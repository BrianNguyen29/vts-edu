package assessments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TransactionManager executes work inside a database transaction.
type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

// Service is the assessments application service contract.
type Service interface {
	ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, *PageInfo, error)

	CreateAssessment(ctx context.Context, actor auth.Actor, classSectionID string, req CreateAssessmentRequest) (AssessmentDetail, error)
	ListAssessmentsByClass(ctx context.Context, actor auth.Actor, classSectionID string) ([]AssessmentListItem, error)
	GetAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (AssessmentDetail, error)
	UpdateAssessment(ctx context.Context, actor auth.Actor, assessmentID string, req UpdateAssessmentRequest) (AssessmentDetail, error)
	CreateSection(ctx context.Context, actor auth.Actor, assessmentID string, req CreateSectionRequest) (SectionDetail, error)
	UpdateSection(ctx context.Context, actor auth.Actor, sectionID string, req UpdateSectionRequest) (SectionDetail, error)
	DeleteSection(ctx context.Context, actor auth.Actor, sectionID string) error

	CreateItem(ctx context.Context, actor auth.Actor, sectionID string, req CreateItemRequest) (ItemDetail, error)
	UpdateItem(ctx context.Context, actor auth.Actor, itemID string, req UpdateItemRequest) (ItemDetail, error)
	DeleteItem(ctx context.Context, actor auth.Actor, itemID string) error

	CreateTarget(ctx context.Context, actor auth.Actor, assessmentID string, req CreateTargetRequest) (TargetDetail, error)
	DeleteTarget(ctx context.Context, actor auth.Actor, assessmentID, targetID string) error

	ReorderSections(ctx context.Context, actor auth.Actor, assessmentID string, req ReorderSectionsRequest) error
	ReorderItems(ctx context.Context, actor auth.Actor, sectionID string, req ReorderItemsRequest) error

	ListQuestions(ctx context.Context, actor auth.Actor, opts ListQuestionsOptions) ([]QuestionPickerItem, *PageInfo, error)
	ListPublications(ctx context.Context, actor auth.Actor, assessmentID string) ([]PublicationSummary, error)

	ValidateAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (ValidationResult, error)
	PublishAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (PublishResult, error)
}

type service struct {
	repo Repository
	tm   TransactionManager
}

// NewService creates the concrete assessments service.
func NewService(repo Repository, tm TransactionManager) Service {
	return &service{repo: repo, tm: tm}
}

func isAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}

func isTeacherOrAdmin(roles []string) bool {
	for _, r := range roles {
		if r == "teacher" || r == "admin" {
			return true
		}
	}
	return false
}

func isDuplicateError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func mapRepoError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// ListAssessments returns the tenant-scoped published/open assessment list.
func (s *service) ListAssessments(ctx context.Context, orgID string, opts ListOptions) ([]AssessmentListItem, *PageInfo, error) {
	queryOpts := opts
	if opts.Limit > 0 {
		queryOpts.Limit = opts.Limit + 1
	}

	rows, err := s.repo.ListPublishedByOrganization(ctx, orgID, queryOpts)
	if err != nil {
		return nil, nil, err
	}

	page := &PageInfo{Limit: opts.Limit, Offset: opts.Offset}
	if opts.Limit > 0 {
		if len(rows) > opts.Limit {
			page.HasMore = true
			last := rows[opts.Limit-1]
			cursor := pagination.Encode(pagination.Cursor{Key: last.CreatedAt, ID: last.ID})
			page.NextCursor = &cursor
			rows = rows[:opts.Limit]
		}
	}

	if opts.Count {
		count, err := s.repo.CountPublishedByOrganization(ctx, orgID, opts)
		if err != nil {
			return nil, nil, err
		}
		page.TotalCount = &count
	}

	items := make([]AssessmentListItem, len(rows))
	for i, r := range rows {
		items[i] = AssessmentListItem{
			ID:              r.ID,
			Title:           r.Title,
			Status:          r.Status,
			DurationMinutes: r.DurationMinutes,
		}
	}
	return items, page, nil
}

func (s *service) CreateAssessment(ctx context.Context, actor auth.Actor, classSectionID string, req CreateAssessmentRequest) (AssessmentDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return AssessmentDetail{}, ErrUnauthorized
	}
	if strings.TrimSpace(req.Title) == "" {
		return AssessmentDetail{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if req.DurationMinutes < 1 {
		return AssessmentDetail{}, fmt.Errorf("%w: duration_minutes must be at least 1", ErrInvalidInput)
	}
	if req.MaxAttempts < 1 {
		req.MaxAttempts = 1
	}

	ok, err := s.repo.IsClassManager(ctx, actor.OrgID, actor.UserID, classSectionID)
	if err != nil {
		return AssessmentDetail{}, err
	}
	if !ok {
		return AssessmentDetail{}, ErrUnauthorized
	}

	var assessment AssessmentDetail
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		assessment, err = s.repo.CreateAssessment(ctx, tx, actor.OrgID, classSectionID, strings.TrimSpace(req.Title), req.DurationMinutes, req.MaxAttempts)
		return err
	})
	if err != nil {
		return AssessmentDetail{}, mapRepoError(err)
	}
	return assessment, nil
}

func (s *service) ListAssessmentsByClass(ctx context.Context, actor auth.Actor, classSectionID string) ([]AssessmentListItem, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, ErrUnauthorized
	}
	ok, err := s.repo.IsClassManager(ctx, actor.OrgID, actor.UserID, classSectionID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrUnauthorized
	}
	return s.repo.ListAssessmentsByClass(ctx, actor.OrgID, classSectionID)
}

func (s *service) GetAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (AssessmentDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return AssessmentDetail{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return AssessmentDetail{}, err
	}
	return s.loadAssessmentDetail(ctx, actor.OrgID, assessmentID)
}

func (s *service) UpdateAssessment(ctx context.Context, actor auth.Actor, assessmentID string, req UpdateAssessmentRequest) (AssessmentDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return AssessmentDetail{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return AssessmentDetail{}, err
	}

	current, err := s.repo.GetAssessment(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return AssessmentDetail{}, mapRepoError(err)
	}
	if current.Status != "DRAFT" {
		return AssessmentDetail{}, ErrNotDraft
	}

	merged := req
	if strings.TrimSpace(merged.Title) == "" {
		merged.Title = current.Title
	}
	if merged.DurationMinutes == nil {
		merged.DurationMinutes = &current.DurationMinutes
	}
	if merged.MaxAttempts == nil {
		merged.MaxAttempts = &current.MaxAttempts
	}
	if merged.Settings == nil {
		merged.Settings = current.Settings
	}

	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.UpdateAssessmentSettings(ctx, tx, actor.OrgID, assessmentID, merged)
	})
	if err != nil {
		return AssessmentDetail{}, mapRepoError(err)
	}
	return s.loadAssessmentDetail(ctx, actor.OrgID, assessmentID)
}

func (s *service) CreateSection(ctx context.Context, actor auth.Actor, assessmentID string, req CreateSectionRequest) (SectionDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return SectionDetail{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return SectionDetail{}, err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return SectionDetail{}, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return SectionDetail{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	var section SectionDetail
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		section, err = s.repo.CreateAssessmentSection(ctx, tx, actor.OrgID, assessmentID, req)
		if isDuplicateError(err) {
			return ErrDuplicateSection
		}
		return err
	})
	if err != nil {
		return SectionDetail{}, mapRepoError(err)
	}
	return section, nil
}

func (s *service) CreateItem(ctx context.Context, actor auth.Actor, sectionID string, req CreateItemRequest) (ItemDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return ItemDetail{}, ErrUnauthorized
	}
	if req.QuestionVersionID == "" {
		return ItemDetail{}, fmt.Errorf("%w: question_version_id is required", ErrInvalidInput)
	}
	if req.Points != "" {
		if _, err := strconv.ParseFloat(req.Points, 64); err != nil {
			return ItemDetail{}, fmt.Errorf("%w: points must be numeric", ErrInvalidInput)
		}
	}

	assessmentID, err := s.repo.GetSectionAssessmentID(ctx, actor.OrgID, sectionID)
	if err != nil {
		return ItemDetail{}, mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return ItemDetail{}, err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return ItemDetail{}, err
	}
	exists, err := s.repo.QuestionVersionExists(ctx, actor.OrgID, req.QuestionVersionID)
	if err != nil {
		return ItemDetail{}, err
	}
	if !exists {
		return ItemDetail{}, fmt.Errorf("%w: question version not found or not published", ErrInvalidInput)
	}

	var item ItemDetail
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		item, err = s.repo.CreateAssessmentItem(ctx, tx, actor.OrgID, assessmentID, sectionID, req)
		if isDuplicateError(err) {
			return ErrDuplicateItem
		}
		return err
	})
	if err != nil {
		return ItemDetail{}, mapRepoError(err)
	}
	return item, nil
}

func (s *service) CreateTarget(ctx context.Context, actor auth.Actor, assessmentID string, req CreateTargetRequest) (TargetDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return TargetDetail{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return TargetDetail{}, err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return TargetDetail{}, err
	}
	if req.ClassSectionID == "" {
		return TargetDetail{}, fmt.Errorf("%w: class_section_id is required", ErrInvalidInput)
	}

	var target TargetDetail
	err := s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		target, err = s.repo.CreateAssessmentTarget(ctx, tx, actor.OrgID, assessmentID, req)
		if isDuplicateError(err) {
			return ErrDuplicateTarget
		}
		return err
	})
	if err != nil {
		return TargetDetail{}, mapRepoError(err)
	}
	return target, nil
}

func (s *service) UpdateSection(ctx context.Context, actor auth.Actor, sectionID string, req UpdateSectionRequest) (SectionDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return SectionDetail{}, ErrUnauthorized
	}
	assessmentID, err := s.repo.GetSectionAssessmentID(ctx, actor.OrgID, sectionID)
	if err != nil {
		return SectionDetail{}, mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return SectionDetail{}, err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return SectionDetail{}, err
	}
	if strings.TrimSpace(req.Title) == "" && req.Position == 0 {
		return SectionDetail{}, fmt.Errorf("%w: title or position required", ErrInvalidInput)
	}

	var section SectionDetail
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		section, err = s.repo.UpdateAssessmentSection(ctx, tx, actor.OrgID, sectionID, req)
		if isDuplicateError(err) {
			return ErrDuplicateSection
		}
		return err
	})
	if err != nil {
		return SectionDetail{}, mapRepoError(err)
	}
	return section, nil
}

func (s *service) DeleteSection(ctx context.Context, actor auth.Actor, sectionID string) error {
	if !isTeacherOrAdmin(actor.Roles) {
		return ErrUnauthorized
	}
	assessmentID, err := s.repo.GetSectionAssessmentID(ctx, actor.OrgID, sectionID)
	if err != nil {
		return mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveAssessmentSection(ctx, tx, actor.OrgID, sectionID)
	})
}

func (s *service) UpdateItem(ctx context.Context, actor auth.Actor, itemID string, req UpdateItemRequest) (ItemDetail, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return ItemDetail{}, ErrUnauthorized
	}
	assessmentID, err := s.repo.GetItemAssessmentID(ctx, actor.OrgID, itemID)
	if err != nil {
		return ItemDetail{}, mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return ItemDetail{}, err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return ItemDetail{}, err
	}
	if req.QuestionVersionID == "" && req.Position == 0 && req.Points == "" {
		return ItemDetail{}, fmt.Errorf("%w: at least one field to update is required", ErrInvalidInput)
	}
	if req.Points != "" {
		if _, err := strconv.ParseFloat(req.Points, 64); err != nil {
			return ItemDetail{}, fmt.Errorf("%w: points must be numeric", ErrInvalidInput)
		}
	}
	if req.QuestionVersionID != "" {
		exists, err := s.repo.QuestionVersionExists(ctx, actor.OrgID, req.QuestionVersionID)
		if err != nil {
			return ItemDetail{}, err
		}
		if !exists {
			return ItemDetail{}, fmt.Errorf("%w: question version not found or not published", ErrInvalidInput)
		}
	}

	var item ItemDetail
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		item, err = s.repo.UpdateAssessmentItem(ctx, tx, actor.OrgID, itemID, req)
		if isDuplicateError(err) {
			return ErrDuplicateItem
		}
		return err
	})
	if err != nil {
		return ItemDetail{}, mapRepoError(err)
	}
	return item, nil
}

func (s *service) DeleteItem(ctx context.Context, actor auth.Actor, itemID string) error {
	if !isTeacherOrAdmin(actor.Roles) {
		return ErrUnauthorized
	}
	assessmentID, err := s.repo.GetItemAssessmentID(ctx, actor.OrgID, itemID)
	if err != nil {
		return mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveAssessmentItem(ctx, tx, actor.OrgID, itemID)
	})
}

func (s *service) DeleteTarget(ctx context.Context, actor auth.Actor, assessmentID, targetID string) error {
	if !isTeacherOrAdmin(actor.Roles) {
		return ErrUnauthorized
	}
	targetAssessmentID, err := s.repo.GetTargetAssessmentID(ctx, actor.OrgID, targetID)
	if err != nil {
		return mapRepoError(err)
	}
	if targetAssessmentID != assessmentID {
		return ErrNotFound
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return err
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ArchiveAssessmentTarget(ctx, tx, actor.OrgID, targetID)
	})
}

func (s *service) ReorderSections(ctx context.Context, actor auth.Actor, assessmentID string, req ReorderSectionsRequest) error {
	if !isTeacherOrAdmin(actor.Roles) {
		return ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return err
	}
	if len(req.SectionIDs) == 0 {
		return fmt.Errorf("%w: section_ids required", ErrInvalidInput)
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ReorderAssessmentSections(ctx, tx, actor.OrgID, assessmentID, req.SectionIDs)
	})
}

func (s *service) ReorderItems(ctx context.Context, actor auth.Actor, sectionID string, req ReorderItemsRequest) error {
	if !isTeacherOrAdmin(actor.Roles) {
		return ErrUnauthorized
	}
	assessmentID, err := s.repo.GetSectionAssessmentID(ctx, actor.OrgID, sectionID)
	if err != nil {
		return mapRepoError(err)
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return err
	}
	if err := s.requireDraft(ctx, actor.OrgID, assessmentID); err != nil {
		return err
	}
	if len(req.ItemIDs) == 0 {
		return fmt.Errorf("%w: item_ids required", ErrInvalidInput)
	}
	return s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return s.repo.ReorderAssessmentItems(ctx, tx, actor.OrgID, sectionID, req.ItemIDs)
	})
}

func (s *service) ListQuestions(ctx context.Context, actor auth.Actor, opts ListQuestionsOptions) ([]QuestionPickerItem, *PageInfo, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, nil, ErrUnauthorized
	}
	queryOpts := opts
	if opts.Limit > 0 {
		queryOpts.Limit = opts.Limit + 1
	}
	rows, err := s.repo.ListQuestions(ctx, actor.OrgID, queryOpts)
	if err != nil {
		return nil, nil, err
	}
	page := &PageInfo{Limit: opts.Limit, Offset: opts.Offset}
	if opts.Limit > 0 {
		if len(rows) > opts.Limit {
			page.HasMore = true
			rows = rows[:opts.Limit]
		}
	}
	return rows, page, nil
}

func (s *service) ListPublications(ctx context.Context, actor auth.Actor, assessmentID string) ([]PublicationSummary, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return nil, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return nil, err
	}
	return s.repo.ListAssessmentPublications(ctx, actor.OrgID, assessmentID)
}

// ValidationResult reports whether an assessment is ready to publish.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

func (s *service) ValidateAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (ValidationResult, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return ValidationResult{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return ValidationResult{}, err
	}

	assessment, err := s.repo.GetAssessment(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return ValidationResult{}, mapRepoError(err)
	}

	var validationErrors []ValidationError
	if assessment.Status != "DRAFT" {
		validationErrors = append(validationErrors, ValidationError{Field: "status", Message: "assessment must be in DRAFT status"})
	}
	if assessment.DurationMinutes < 1 {
		validationErrors = append(validationErrors, ValidationError{Field: "duration_minutes", Message: "duration must be at least 1 minute"})
	}
	if assessment.MaxAttempts < 1 {
		validationErrors = append(validationErrors, ValidationError{Field: "max_attempts", Message: "max_attempts must be at least 1"})
	}
	if assessment.OpensAt != nil && assessment.ClosesAt != nil {
		opens, err1 := time.Parse(time.RFC3339, *assessment.OpensAt)
		closes, err2 := time.Parse(time.RFC3339, *assessment.ClosesAt)
		if err1 == nil && err2 == nil && !closes.After(opens) {
			validationErrors = append(validationErrors, ValidationError{Field: "schedule", Message: "opens_at must be before closes_at"})
		}
	}

	sections, err := s.repo.GetAssessmentSections(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return ValidationResult{}, err
	}
	if len(sections) == 0 {
		validationErrors = append(validationErrors, ValidationError{Field: "sections", Message: "at least one active section is required"})
	}

	items, err := s.repo.GetAssessmentItems(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return ValidationResult{}, err
	}
	if len(items) == 0 {
		validationErrors = append(validationErrors, ValidationError{Field: "items", Message: "at least one active item is required"})
	}
	for _, it := range items {
		if pts, err := strconv.ParseFloat(it.Points, 64); err != nil || pts <= 0 {
			validationErrors = append(validationErrors, ValidationError{Field: "items", Message: fmt.Sprintf("item %s points must be greater than 0", it.ID)})
		}
		published, err := s.repo.IsQuestionVersionPublished(ctx, actor.OrgID, it.QuestionVersionID)
		if err != nil {
			return ValidationResult{}, err
		}
		if !published {
			validationErrors = append(validationErrors, ValidationError{Field: "items", Message: fmt.Sprintf("item %s question version is not published", it.ID)})
		}
	}

	targets, err := s.repo.GetAssessmentTargets(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return ValidationResult{}, err
	}
	if len(targets) == 0 {
		validationErrors = append(validationErrors, ValidationError{Field: "targets", Message: "at least one active target class is required"})
	}
	for _, tgt := range targets {
		active, err := s.repo.IsClassSectionActive(ctx, actor.OrgID, tgt.ClassSectionID)
		if err != nil {
			return ValidationResult{}, err
		}
		if !active {
			validationErrors = append(validationErrors, ValidationError{Field: "targets", Message: fmt.Sprintf("target class %s is not active", tgt.ClassSectionID)})
		}
	}

	return ValidationResult{
		Valid:  len(validationErrors) == 0,
		Errors: validationErrors,
	}, nil
}

func (s *service) PublishAssessment(ctx context.Context, actor auth.Actor, assessmentID string) (PublishResult, error) {
	if !isTeacherOrAdmin(actor.Roles) {
		return PublishResult{}, ErrUnauthorized
	}
	if err := s.requireManager(ctx, actor, assessmentID); err != nil {
		return PublishResult{}, err
	}

	validation, err := s.ValidateAssessment(ctx, actor, assessmentID)
	if err != nil {
		return PublishResult{}, err
	}
	if !validation.Valid {
		return PublishResult{}, fmt.Errorf("%w: assessment is not valid for publish", ErrValidationFailed)
	}

	assessment, err := s.repo.GetAssessment(ctx, actor.OrgID, assessmentID)
	if err != nil {
		return PublishResult{}, mapRepoError(err)
	}

	newStatus := determinePublishStatus(assessment.OpensAt, assessment.ClosesAt)
	snapshot, err := s.buildSnapshot(ctx, actor.OrgID, assessment)
	if err != nil {
		return PublishResult{}, fmt.Errorf("build snapshot: %w", err)
	}

	var result PublishResult
	err = s.tm.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		revision, err := s.repo.PublishAssessment(ctx, tx, actor.OrgID, assessmentID, newStatus)
		if err != nil {
			return err
		}
		if err := s.repo.InsertAssessmentPublication(ctx, tx, actor.OrgID, assessmentID, revision, snapshot); err != nil {
			return err
		}
		result = PublishResult{
			ID:       assessmentID,
			Status:   newStatus,
			Revision: revision,
		}
		return nil
	})
	if err != nil {
		return PublishResult{}, mapRepoError(err)
	}
	return result, nil
}

func (s *service) requireManager(ctx context.Context, actor auth.Actor, assessmentID string) error {
	ok, err := s.repo.IsAssessmentManager(ctx, actor.OrgID, actor.UserID, assessmentID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnauthorized
	}
	return nil
}

func (s *service) requireDraft(ctx context.Context, orgID, assessmentID string) error {
	assessment, err := s.repo.GetAssessment(ctx, orgID, assessmentID)
	if err != nil {
		return err
	}
	if assessment.Status != "DRAFT" {
		return ErrNotDraft
	}
	return nil
}

func (s *service) loadAssessmentDetail(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
	assessment, err := s.repo.GetAssessment(ctx, orgID, assessmentID)
	if err != nil {
		return AssessmentDetail{}, mapRepoError(err)
	}
	sections, err := s.repo.GetAssessmentSections(ctx, orgID, assessmentID)
	if err != nil {
		return AssessmentDetail{}, err
	}
	items, err := s.repo.GetAssessmentItems(ctx, orgID, assessmentID)
	if err != nil {
		return AssessmentDetail{}, err
	}
	targets, err := s.repo.GetAssessmentTargets(ctx, orgID, assessmentID)
	if err != nil {
		return AssessmentDetail{}, err
	}

	sectionMap := make(map[string]*SectionDetail, len(sections))
	for i := range sections {
		sectionMap[sections[i].ID] = &sections[i]
	}
	for i := range items {
		section := sectionMap[items[i].AssessmentSectionID]
		if section == nil {
			// Item references section that may have been archived; skip.
			continue
		}
		section.Items = append(section.Items, items[i])
	}

	assessment.Sections = sections
	assessment.Targets = targets
	return assessment, nil
}

func determinePublishStatus(opensAt, closesAt *string) string {
	now := time.Now().UTC()
	if opensAt != nil {
		t, err := time.Parse(time.RFC3339, *opensAt)
		if err == nil && t.After(now) {
			return "PUBLISHED"
		}
	}
	if closesAt != nil {
		t, err := time.Parse(time.RFC3339, *closesAt)
		if err == nil && !t.After(now) {
			return "CLOSED"
		}
	}
	return "OPEN"
}

func (s *service) buildSnapshot(ctx context.Context, orgID string, assessment AssessmentDetail) (json.RawMessage, error) {
	sections, err := s.repo.GetAssessmentSections(ctx, orgID, assessment.ID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.GetAssessmentItemsWithContent(ctx, orgID, assessment.ID)
	if err != nil {
		return nil, err
	}
	targets, err := s.repo.GetAssessmentTargets(ctx, orgID, assessment.ID)
	if err != nil {
		return nil, err
	}

	type snapshotItem struct {
		ID                string          `json:"id"`
		QuestionVersionID string          `json:"question_version_id"`
		Position          int             `json:"position"`
		Points            string          `json:"points"`
		Prompt            json.RawMessage `json:"prompt"`
		Choices           json.RawMessage `json:"choices"`
		AnswerKey         json.RawMessage `json:"answer_key"`
		MaxScore          string          `json:"max_score"`
	}

	type snapshotSection struct {
		ID       string         `json:"id"`
		Title    string         `json:"title"`
		Position int            `json:"position"`
		Items    []snapshotItem `json:"items"`
	}

	sectionMap := make(map[string]*snapshotSection, len(sections))
	for i, sec := range sections {
		sectionMap[sec.ID] = &snapshotSection{
			ID:       sec.ID,
			Title:    sec.Title,
			Position: sec.Position,
			Items:    []snapshotItem{},
		}
		_ = i
	}
	for _, it := range items {
		sec := sectionMap[it.AssessmentSectionID]
		if sec == nil {
			continue
		}
		sec.Items = append(sec.Items, snapshotItem{
			ID:                it.ID,
			QuestionVersionID: it.QuestionVersionID,
			Position:          it.Position,
			Points:            it.Points,
			Prompt:            it.Prompt,
			Choices:           it.Choices,
			AnswerKey:         it.AnswerKey,
			MaxScore:          it.MaxScore,
		})
	}

	var sectionList []snapshotSection
	for _, sec := range sections {
		if s := sectionMap[sec.ID]; s != nil {
			sectionList = append(sectionList, *s)
		}
	}

	snapshot := struct {
		ID              string            `json:"id"`
		Title           string            `json:"title"`
		DurationMinutes int               `json:"duration_minutes"`
		MaxAttempts     int               `json:"max_attempts"`
		Instructions    string            `json:"instructions,omitempty"`
		OpensAt         *string           `json:"opens_at,omitempty"`
		ClosesAt        *string           `json:"closes_at,omitempty"`
		Settings        json.RawMessage   `json:"settings,omitempty"`
		Revision        int               `json:"revision"`
		Sections        []snapshotSection `json:"sections"`
		Targets         []TargetDetail    `json:"targets"`
	}{
		ID:              assessment.ID,
		Title:           assessment.Title,
		DurationMinutes: assessment.DurationMinutes,
		MaxAttempts:     assessment.MaxAttempts,
		Instructions:    assessment.Instructions,
		OpensAt:         assessment.OpensAt,
		ClosesAt:        assessment.ClosesAt,
		Settings:        assessment.Settings,
		Revision:        assessment.Revision,
		Sections:        sectionList,
		Targets:         targets,
	}

	return json.Marshal(snapshot)
}
