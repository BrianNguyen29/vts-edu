package assessments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	assessmentsqlc "github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments/sqlc"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for the assessments feature.
type Repository interface {
	ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error)
	CountPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) (int64, error)

	CreateAssessment(ctx context.Context, tx pgx.Tx, orgID, classSectionID, title string, durationMinutes, maxAttempts int) (AssessmentDetail, error)
	ListAssessmentsByClass(ctx context.Context, orgID, classSectionID string) ([]AssessmentListItem, error)
	GetAssessment(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error)
	GetSectionAssessmentID(ctx context.Context, orgID, sectionID string) (string, error)
	GetAssessmentSections(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error)
	GetAssessmentItems(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error)
	GetAssessmentTargets(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error)
	UpdateAssessmentSettings(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req UpdateAssessmentRequest) error
	CreateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateSectionRequest) (SectionDetail, error)
	CreateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string, req CreateItemRequest) (ItemDetail, error)
	CreateAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateTargetRequest) (TargetDetail, error)

	DuplicateSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string) (SectionDetail, error)
	DuplicateItem(ctx context.Context, tx pgx.Tx, orgID, sectionID, itemID string) (ItemDetail, error)

	GetAssessmentSection(ctx context.Context, orgID, sectionID string) (SectionDetail, error)
	UpdateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string, req UpdateSectionRequest) (SectionDetail, error)
	ArchiveAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string) error

	GetAssessmentItem(ctx context.Context, orgID, itemID string) (ItemDetail, error)
	GetItemAssessmentID(ctx context.Context, orgID, itemID string) (string, error)
	UpdateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string, req UpdateItemRequest) (ItemDetail, error)
	ArchiveAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string) error

	GetAssessmentTarget(ctx context.Context, orgID, targetID string) (TargetDetail, error)
	GetTargetAssessmentID(ctx context.Context, orgID, targetID string) (string, error)
	ArchiveAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, targetID string) error

	ReorderAssessmentSections(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, sectionIDs []string) error
	ReorderAssessmentItems(ctx context.Context, tx pgx.Tx, orgID, sectionID string, itemIDs []string) error

	GetAssessmentItemsWithContent(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error)
	PublishAssessment(ctx context.Context, tx pgx.Tx, orgID, assessmentID, newStatus string) (int, error)
	InsertAssessmentPublication(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, version int, snapshot json.RawMessage) error
	ListAssessmentPublications(ctx context.Context, orgID, assessmentID string) ([]PublicationSummary, error)
	CountAssessmentSections(ctx context.Context, orgID, assessmentID string) (int64, error)
	CountAssessmentItems(ctx context.Context, orgID, assessmentID string) (int64, error)
	CountAssessmentTargets(ctx context.Context, orgID, assessmentID string) (int64, error)
	QuestionVersionExists(ctx context.Context, orgID, questionVersionID string) (bool, error)
	IsQuestionVersionPublished(ctx context.Context, orgID, questionVersionID string) (bool, error)
	IsClassSectionActive(ctx context.Context, orgID, classSectionID string) (bool, error)
	ListQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) ([]QuestionPickerItem, error)
	CountQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) (int64, error)
	IsClassManager(ctx context.Context, orgID, userID, classSectionID string) (bool, error)
	IsAssessmentManager(ctx context.Context, orgID, userID, assessmentID string) (bool, error)

	TransitionAssessmentsToOpen(ctx context.Context) (int64, error)
	TransitionAssessmentsToClosed(ctx context.Context) (int64, error)

	// Question bank editor
	CreateQuestionBank(ctx context.Context, tx pgx.Tx, orgID, title string) (QuestionBank, error)
	ListQuestionBanksByOrganization(ctx context.Context, orgID string, opts ListQuestionBanksOptions) ([]QuestionBank, error)
	GetQuestionBank(ctx context.Context, orgID, bankID string) (QuestionBank, error)

	CreateQuestion(ctx context.Context, tx pgx.Tx, bankID string) (QuestionBankQuestion, error)
	ListQuestionsInBank(ctx context.Context, bankID string, opts ListQuestionBanksOptions) ([]QuestionBankQuestion, error)
	GetQuestionWithBank(ctx context.Context, questionID string) (QuestionBankQuestion, string, error)
	GetQuestion(ctx context.Context, bankID, questionID string) (QuestionBankQuestion, error)

	CreateQuestionVersion(ctx context.Context, tx pgx.Tx, questionID string, req CreateQuestionVersionRequest, maxScore string, version int) (QuestionVersion, error)
	GetQuestionVersion(ctx context.Context, orgID, versionID string) (QuestionVersion, error)
	GetLatestVersionNumber(ctx context.Context, tx pgx.Tx, questionID string) (int, error)
	PublishQuestionVersion(ctx context.Context, tx pgx.Tx, versionID string) (QuestionVersion, error)
}

type sqlcRepository struct {
	queries *assessmentsqlc.Queries
}

// NewRepository creates a new assessments repository backed by generated sqlc
// queries. It preserves the existing Repository interface.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &sqlcRepository{queries: assessmentsqlc.New(pool)}
}

func toUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func textPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func tsPtr(t pgtype.Timestamptz) *string {
	if t.Valid {
		s := t.Time.UTC().Format(time.RFC3339)
		return &s
	}
	return nil
}

func tsFromString(s string) (pgtype.Timestamptz, error) {
	if s == "" {
		return pgtype.Timestamptz{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return pgtype.Timestamptz{}, err
	}
	return pgtype.Timestamptz{Time: t, Valid: true}, nil
}

func uuidPtr(u pgtype.UUID) *string {
	if u.Valid {
		s := u.String()
		return &s
	}
	return nil
}

func numericString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	f, err := n.Float64Value()
	if err != nil {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", f.Float64)
}

func toNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func decodeAssessmentCursor(cursor string) (string, pgtype.UUID, error) {
	if cursor == "" {
		return "", pgtype.UUID{}, nil
	}
	c, err := pagination.Decode(cursor)
	if err != nil {
		return "", pgtype.UUID{}, err
	}
	id, err := toUUID(c.ID)
	if err != nil {
		return "", pgtype.UUID{}, pagination.ErrInvalidCursor
	}
	return c.Key, id, nil
}

func (r *sqlcRepository) ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error) {
	var orgUUID pgtype.UUID
	if err := orgUUID.Scan(orgID); err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}

	key, cursorID, err := decodeAssessmentCursor(opts.Cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor: %w", err)
	}

	rows, err := r.queries.ListPublishedByOrganization(ctx, assessmentsqlc.ListPublishedByOrganizationParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
		CursorKey:      key,
		CursorID:       cursorID,
		PageOffset:     int32(opts.Offset),
		PageLimit:      int32(opts.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list assessments: %w", err)
	}

	list := make([]Assessment, len(rows))
	for i, row := range rows {
		list[i] = Assessment{
			ID:              row.ID.String(),
			Title:           row.Title,
			Status:          row.Status,
			DurationMinutes: int(row.DurationMinutes),
			CreatedAt:       row.CreatedAt.Time.Format(time.RFC3339),
		}
	}
	return list, nil
}

func (r *sqlcRepository) CountPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) (int64, error) {
	var orgUUID pgtype.UUID
	if err := orgUUID.Scan(orgID); err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}

	count, err := r.queries.CountPublishedByOrganization(ctx, assessmentsqlc.CountPublishedByOrganizationParams{
		OrganizationID: orgUUID,
		SearchQuery:    opts.Query,
	})
	if err != nil {
		return 0, fmt.Errorf("count assessments: %w", err)
	}
	return count, nil
}

func assessmentDetailFromRow(row assessmentsqlc.GetAssessmentRow) AssessmentDetail {
	var settings json.RawMessage
	if len(row.SettingsJson) > 0 {
		settings = row.SettingsJson
	}
	return AssessmentDetail{
		ID:              row.ID.String(),
		ClassSectionID:  uuidPtr(row.ClassSectionID),
		Title:           row.Title,
		Status:          row.Status,
		DurationMinutes: int(row.DurationMinutes),
		MaxAttempts:     int(row.MaxAttempts),
		Revision:        int(row.Revision),
		Instructions:    textString(row.Instructions),
		OpensAt:         tsPtr(row.OpensAt),
		ClosesAt:        tsPtr(row.ClosesAt),
		Settings:        settings,
		Sections:        []SectionDetail{},
		Targets:         []TargetDetail{},
		CreatedAt:       row.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:       row.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}
}

func textString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func (r *sqlcRepository) CreateAssessment(ctx context.Context, tx pgx.Tx, orgID, classSectionID, title string, durationMinutes, maxAttempts int) (AssessmentDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classSectionID)
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("invalid class id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateAssessment(ctx, assessmentsqlc.CreateAssessmentParams{
		OrganizationID:  orgUUID,
		ClassSectionID:  classUUID,
		Title:           title,
		DurationMinutes: int32(durationMinutes),
		MaxAttempts:     int32(maxAttempts),
	})
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("create assessment: %w", err)
	}
	return assessmentDetailFromRow(assessmentsqlc.GetAssessmentRow{
		ID:              row.ID,
		OrganizationID:  row.OrganizationID,
		ClassSectionID:  row.ClassSectionID,
		Title:           row.Title,
		DurationMinutes: row.DurationMinutes,
		MaxAttempts:     row.MaxAttempts,
		SettingsJson:    row.SettingsJson,
		Status:          row.Status,
		Revision:        row.Revision,
		Instructions:    row.Instructions,
		OpensAt:         row.OpensAt,
		ClosesAt:        row.ClosesAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}), nil
}

func (r *sqlcRepository) ListAssessmentsByClass(ctx context.Context, orgID, classSectionID string) ([]AssessmentListItem, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classSectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid class id: %w", err)
	}
	rows, err := r.queries.ListAssessmentsByClass(ctx, assessmentsqlc.ListAssessmentsByClassParams{
		OrganizationID: orgUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("list assessments by class: %w", err)
	}
	list := make([]AssessmentListItem, len(rows))
	for i, row := range rows {
		list[i] = AssessmentListItem{
			ID:              row.ID.String(),
			Title:           row.Title,
			Status:          row.Status,
			DurationMinutes: int(row.DurationMinutes),
		}
	}
	return list, nil
}

func (r *sqlcRepository) GetAssessment(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("invalid assessment id: %w", err)
	}
	row, err := r.queries.GetAssessment(ctx, assessmentsqlc.GetAssessmentParams{
		ID:             id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return AssessmentDetail{}, ErrNotFound
	}
	if err != nil {
		return AssessmentDetail{}, fmt.Errorf("get assessment: %w", err)
	}
	return assessmentDetailFromRow(row), nil
}

func (r *sqlcRepository) GetSectionAssessmentID(ctx context.Context, orgID, sectionID string) (string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(sectionID)
	if err != nil {
		return "", fmt.Errorf("invalid section id: %w", err)
	}
	row, err := r.queries.GetSectionAssessmentID(ctx, assessmentsqlc.GetSectionAssessmentIDParams{
		SectionID:      id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get section assessment id: %w", err)
	}
	return row.String(), nil
}

func (r *sqlcRepository) GetAssessmentSections(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.GetAssessmentSections(ctx, assessmentsqlc.GetAssessmentSectionsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("get assessment sections: %w", err)
	}
	sections := make([]SectionDetail, len(rows))
	for i, row := range rows {
		var settings json.RawMessage
		if len(row.SettingsJson) > 0 {
			settings = row.SettingsJson
		}
		sections[i] = SectionDetail{
			ID:       row.ID.String(),
			Title:    row.Title,
			Position: int(row.Position),
			Settings: settings,
			Items:    []ItemDetail{},
		}
	}
	return sections, nil
}

func (r *sqlcRepository) GetAssessmentItems(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.GetAssessmentItems(ctx, assessmentsqlc.GetAssessmentItemsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("get assessment items: %w", err)
	}
	items := make([]ItemDetail, len(rows))
	for i, row := range rows {
		items[i] = ItemDetail{
			ID:                  row.ID.String(),
			AssessmentSectionID: row.AssessmentSectionID.String(),
			QuestionVersionID:   row.QuestionVersionID.String(),
			Position:            int(row.Position),
			Points:              numericString(row.Points),
		}
	}
	return items, nil
}

// ItemContentRow extends an item with question-version content for snapshots.
type ItemContentRow struct {
	ID                  string
	AssessmentSectionID string
	QuestionVersionID   string
	Position            int
	Points              string
	Prompt              json.RawMessage
	Choices             json.RawMessage
	AnswerKey           json.RawMessage
	MaxScore            string
	QuestionType        string
}

func (r *sqlcRepository) GetAssessmentItemsWithContent(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.GetAssessmentItemsWithContent(ctx, assessmentsqlc.GetAssessmentItemsWithContentParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("get assessment items with content: %w", err)
	}
	items := make([]ItemContentRow, len(rows))
	for i, row := range rows {
		items[i] = ItemContentRow{
			ID:                  row.ID.String(),
			AssessmentSectionID: row.AssessmentSectionID.String(),
			QuestionVersionID:   row.QuestionVersionID.String(),
			Position:            int(row.Position),
			Points:              numericString(row.Points),
			Prompt:              row.PromptJson,
			Choices:             row.ChoicesJson,
			AnswerKey:           row.AnswerKeyJson,
			MaxScore:            numericString(row.MaxScore),
			QuestionType:        row.QuestionType,
		}
	}
	return items, nil
}

func (r *sqlcRepository) GetAssessmentTargets(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.GetAssessmentTargets(ctx, assessmentsqlc.GetAssessmentTargetsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("get assessment targets: %w", err)
	}
	targets := make([]TargetDetail, len(rows))
	for i, row := range rows {
		targets[i] = TargetDetail{
			ID:             row.ID.String(),
			ClassSectionID: row.ClassSectionID.String(),
		}
	}
	return targets, nil
}

func (r *sqlcRepository) GetAssessmentSection(ctx context.Context, orgID, sectionID string) (SectionDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(sectionID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid section id: %w", err)
	}
	row, err := r.queries.GetAssessmentSection(ctx, assessmentsqlc.GetAssessmentSectionParams{
		SectionID:      id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SectionDetail{}, ErrNotFound
	}
	if err != nil {
		return SectionDetail{}, fmt.Errorf("get assessment section: %w", err)
	}
	return sectionDetailFromRow(row), nil
}

func sectionDetailFromRow(row assessmentsqlc.GetAssessmentSectionRow) SectionDetail {
	var settings json.RawMessage
	if len(row.SettingsJson) > 0 {
		settings = row.SettingsJson
	}
	return SectionDetail{
		ID:       row.ID.String(),
		Title:    row.Title,
		Position: int(row.Position),
		Settings: settings,
		Items:    []ItemDetail{},
	}
}

func (r *sqlcRepository) UpdateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string, req UpdateSectionRequest) (SectionDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(sectionID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid section id: %w", err)
	}
	row, err := r.queries.WithTx(tx).UpdateAssessmentSection(ctx, assessmentsqlc.UpdateAssessmentSectionParams{
		Title:          req.Title,
		Position:       int32(req.Position),
		SectionID:      id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SectionDetail{}, ErrNotFound
	}
	if err != nil {
		return SectionDetail{}, fmt.Errorf("update assessment section: %w", err)
	}
	return sectionDetailFromRow(assessmentsqlc.GetAssessmentSectionRow{
		ID:           row.ID,
		AssessmentID: row.AssessmentID,
		Title:        row.Title,
		Position:     row.Position,
		SettingsJson: row.SettingsJson,
		Status:       row.Status,
	}), nil
}

func (r *sqlcRepository) ArchiveAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(sectionID)
	if err != nil {
		return fmt.Errorf("invalid section id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveAssessmentSection(ctx, assessmentsqlc.ArchiveAssessmentSectionParams{
		SectionID:      id,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive assessment section: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) GetAssessmentItem(ctx context.Context, orgID, itemID string) (ItemDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(itemID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid item id: %w", err)
	}
	row, err := r.queries.GetAssessmentItem(ctx, assessmentsqlc.GetAssessmentItemParams{
		ItemID:         id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ItemDetail{}, ErrNotFound
	}
	if err != nil {
		return ItemDetail{}, fmt.Errorf("get assessment item: %w", err)
	}
	return itemDetailFromRow(row), nil
}

func itemDetailFromRow(row assessmentsqlc.GetAssessmentItemRow) ItemDetail {
	return ItemDetail{
		ID:                  row.ID.String(),
		AssessmentSectionID: row.AssessmentSectionID.String(),
		QuestionVersionID:   row.QuestionVersionID.String(),
		Position:            int(row.Position),
		Points:              numericString(row.Points),
	}
}

func (r *sqlcRepository) GetItemAssessmentID(ctx context.Context, orgID, itemID string) (string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(itemID)
	if err != nil {
		return "", fmt.Errorf("invalid item id: %w", err)
	}
	row, err := r.queries.GetItemAssessmentID(ctx, assessmentsqlc.GetItemAssessmentIDParams{
		ItemID:         id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get item assessment id: %w", err)
	}
	return row.String(), nil
}

func (r *sqlcRepository) UpdateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string, req UpdateItemRequest) (ItemDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(itemID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid item id: %w", err)
	}
	var qvUUID pgtype.UUID
	if req.QuestionVersionID != "" {
		qvUUID, err = toUUID(req.QuestionVersionID)
		if err != nil {
			return ItemDetail{}, fmt.Errorf("invalid question version id: %w", err)
		}
	}
	var points pgtype.Numeric
	if req.Points != "" {
		points, err = toNumeric(req.Points)
		if err != nil {
			return ItemDetail{}, fmt.Errorf("invalid points: %w", err)
		}
	}
	row, err := r.queries.WithTx(tx).UpdateAssessmentItem(ctx, assessmentsqlc.UpdateAssessmentItemParams{
		QuestionVersionID: qvUUID,
		Points:            points,
		Position:          int32(req.Position),
		ItemID:            id,
		OrganizationID:    orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ItemDetail{}, ErrNotFound
	}
	if err != nil {
		return ItemDetail{}, fmt.Errorf("update assessment item: %w", err)
	}
	return itemDetailFromRow(assessmentsqlc.GetAssessmentItemRow{
		ID:                  row.ID,
		AssessmentSectionID: row.AssessmentSectionID,
		QuestionVersionID:   row.QuestionVersionID,
		Position:            row.Position,
		Points:              row.Points,
		Status:              row.Status,
	}), nil
}

func (r *sqlcRepository) ArchiveAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(itemID)
	if err != nil {
		return fmt.Errorf("invalid item id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveAssessmentItem(ctx, assessmentsqlc.ArchiveAssessmentItemParams{
		ItemID:         id,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive assessment item: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) GetAssessmentTarget(ctx context.Context, orgID, targetID string) (TargetDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return TargetDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(targetID)
	if err != nil {
		return TargetDetail{}, fmt.Errorf("invalid target id: %w", err)
	}
	row, err := r.queries.GetAssessmentTarget(ctx, assessmentsqlc.GetAssessmentTargetParams{
		TargetID:       id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return TargetDetail{}, ErrNotFound
	}
	if err != nil {
		return TargetDetail{}, fmt.Errorf("get assessment target: %w", err)
	}
	return TargetDetail{
		ID:             row.ID.String(),
		ClassSectionID: row.ClassSectionID.String(),
	}, nil
}

func (r *sqlcRepository) GetTargetAssessmentID(ctx context.Context, orgID, targetID string) (string, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return "", fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(targetID)
	if err != nil {
		return "", fmt.Errorf("invalid target id: %w", err)
	}
	row, err := r.queries.GetTargetAssessmentID(ctx, assessmentsqlc.GetTargetAssessmentIDParams{
		TargetID:       id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get target assessment id: %w", err)
	}
	return row.String(), nil
}

func (r *sqlcRepository) ArchiveAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, targetID string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(targetID)
	if err != nil {
		return fmt.Errorf("invalid target id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).ArchiveAssessmentTarget(ctx, assessmentsqlc.ArchiveAssessmentTargetParams{
		TargetID:       id,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return fmt.Errorf("archive assessment target: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) ReorderAssessmentSections(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, sectionIDs []string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	assessUUID, err := toUUID(assessmentID)
	if err != nil {
		return fmt.Errorf("invalid assessment id: %w", err)
	}
	// Verify all provided IDs belong to this assessment and are active.
	activeRows, err := r.queries.WithTx(tx).GetAssessmentSections(ctx, assessmentsqlc.GetAssessmentSectionsParams{
		OrganizationID: orgUUID,
		AssessmentID:   assessUUID,
	})
	if err != nil {
		return fmt.Errorf("list active sections: %w", err)
	}
	activeSet := make(map[string]bool, len(activeRows))
	for _, row := range activeRows {
		activeSet[row.ID.String()] = true
	}
	if len(sectionIDs) != len(activeSet) {
		return fmt.Errorf("%w: reorder must include all active sections", ErrInvalidInput)
	}
	for _, sid := range sectionIDs {
		if !activeSet[sid] {
			return fmt.Errorf("%w: unknown or inactive section %s", ErrInvalidInput, sid)
		}
	}
	// Assign positions sequentially to avoid unique-constraint conflicts.
	for i, sid := range sectionIDs {
		sectionUUID, err := toUUID(sid)
		if err != nil {
			return fmt.Errorf("invalid section id: %w", err)
		}
		if _, err := r.queries.WithTx(tx).UpdateAssessmentSectionPosition(ctx, assessmentsqlc.UpdateAssessmentSectionPositionParams{
			Position:       int32((i + 1) * 10),
			SectionID:      sectionUUID,
			OrganizationID: orgUUID,
		}); err != nil {
			return fmt.Errorf("update section position: %w", err)
		}
	}
	return nil
}

func (r *sqlcRepository) ReorderAssessmentItems(ctx context.Context, tx pgx.Tx, orgID, sectionID string, itemIDs []string) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	sectionUUID, err := toUUID(sectionID)
	if err != nil {
		return fmt.Errorf("invalid section id: %w", err)
	}
	activeRows, err := r.queries.WithTx(tx).GetAssessmentItemsBySection(ctx, assessmentsqlc.GetAssessmentItemsBySectionParams{
		OrganizationID: orgUUID,
		SectionID:      sectionUUID,
	})
	if err != nil {
		return fmt.Errorf("list active items: %w", err)
	}
	activeSet := make(map[string]bool, len(activeRows))
	for _, row := range activeRows {
		activeSet[row.ID.String()] = true
	}
	if len(itemIDs) != len(activeSet) {
		return fmt.Errorf("%w: reorder must include all active items in the section", ErrInvalidInput)
	}
	for _, iid := range itemIDs {
		if !activeSet[iid] {
			return fmt.Errorf("%w: unknown or inactive item %s", ErrInvalidInput, iid)
		}
	}
	for i, iid := range itemIDs {
		itemUUID, err := toUUID(iid)
		if err != nil {
			return fmt.Errorf("invalid item id: %w", err)
		}
		if _, err := r.queries.WithTx(tx).UpdateAssessmentItemPosition(ctx, assessmentsqlc.UpdateAssessmentItemPositionParams{
			Position:       int32((i + 1) * 10),
			ItemID:         itemUUID,
			OrganizationID: orgUUID,
		}); err != nil {
			return fmt.Errorf("update item position: %w", err)
		}
	}
	return nil
}

func (r *sqlcRepository) ListQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) ([]QuestionPickerItem, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.ListQuestions(ctx, assessmentsqlc.ListQuestionsParams{
		OrganizationID: orgUUID,
		BankID:         opts.BankID,
		SearchQuery:    opts.Query,
		PageLimit:      int32(opts.Limit),
		PageOffset:     int32(opts.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list questions: %w", err)
	}
	items := make([]QuestionPickerItem, 0, len(rows))
	for _, row := range rows {
		// COALESCE gives a nil UUID for questions without a published version.
		if !row.QuestionVersionID.Valid || row.QuestionVersionID.String() == "00000000-0000-0000-0000-000000000000" {
			continue
		}
		qvID := row.QuestionVersionID.String()
		prompt := ""
		if s, ok := row.PromptText.(string); ok {
			prompt = s
		}
		items = append(items, QuestionPickerItem{
			ID:                    row.ID.String(),
			QuestionBankID:        row.QuestionBankID.String(),
			QuestionVersionID:     qvID,
			QuestionVersionStatus: row.QuestionVersionStatus,
			QuestionType:          row.QuestionType,
			Prompt:                prompt,
		})
	}
	return items, nil
}

func (r *sqlcRepository) CountQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	count, err := r.queries.CountQuestions(ctx, assessmentsqlc.CountQuestionsParams{
		OrganizationID: orgUUID,
		BankID:         opts.BankID,
		SearchQuery:    opts.Query,
	})
	if err != nil {
		return 0, fmt.Errorf("count questions: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) ListAssessmentPublications(ctx context.Context, orgID, assessmentID string) ([]PublicationSummary, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.ListAssessmentPublications(ctx, assessmentsqlc.ListAssessmentPublicationsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return nil, fmt.Errorf("list assessment publications: %w", err)
	}
	pubs := make([]PublicationSummary, len(rows))
	for i, row := range rows {
		pubs[i] = PublicationSummary{
			ID:          row.ID.String(),
			Version:     int(row.Version),
			Status:      row.Status,
			PublishedAt: row.PublishedAt.Time.UTC().Format(time.RFC3339),
		}
	}
	return pubs, nil
}

func (r *sqlcRepository) IsQuestionVersionPublished(ctx context.Context, orgID, questionVersionID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	qvUUID, err := toUUID(questionVersionID)
	if err != nil {
		return false, fmt.Errorf("invalid question version id: %w", err)
	}
	ok, err := r.queries.IsQuestionVersionPublished(ctx, assessmentsqlc.IsQuestionVersionPublishedParams{
		QuestionVersionID: qvUUID,
		OrganizationID:    orgUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check question version published: %w", err)
	}
	return ok, nil
}

func (r *sqlcRepository) IsClassSectionActive(ctx context.Context, orgID, classSectionID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	classUUID, err := toUUID(classSectionID)
	if err != nil {
		return false, fmt.Errorf("invalid class section id: %w", err)
	}
	ok, err := r.queries.IsClassSectionActive(ctx, assessmentsqlc.IsClassSectionActiveParams{
		ClassSectionID: classUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check class section active: %w", err)
	}
	return ok, nil
}

func (r *sqlcRepository) UpdateAssessmentSettings(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req UpdateAssessmentRequest) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return fmt.Errorf("invalid assessment id: %w", err)
	}
	duration := int32(0)
	if req.DurationMinutes != nil {
		duration = int32(*req.DurationMinutes)
	}
	maxAttempts := int32(0)
	if req.MaxAttempts != nil {
		maxAttempts = int32(*req.MaxAttempts)
	}
	opensAt, err := tsFromString(req.OpensAt)
	if err != nil {
		return fmt.Errorf("invalid opens_at: %w", err)
	}
	closesAt, err := tsFromString(req.ClosesAt)
	if err != nil {
		return fmt.Errorf("invalid closes_at: %w", err)
	}
	rows, err := r.queries.WithTx(tx).UpdateAssessmentSettings(ctx, assessmentsqlc.UpdateAssessmentSettingsParams{
		Title:           req.Title,
		DurationMinutes: duration,
		MaxAttempts:     maxAttempts,
		Instructions:    toText(req.Instructions),
		OpensAt:         opensAt,
		ClosesAt:        closesAt,
		SettingsJson:    req.Settings,
		ID:              id,
		OrganizationID:  orgUUID,
	})
	if err != nil {
		return fmt.Errorf("update assessment settings: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlcRepository) CreateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateSectionRequest) (SectionDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid assessment id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateAssessmentSection(ctx, assessmentsqlc.CreateAssessmentSectionParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
		Title:          req.Title,
		Position:       int32(req.Position),
	})
	if err != nil {
		return SectionDetail{}, fmt.Errorf("create assessment section: %w", err)
	}
	return SectionDetail{
		ID:       row.ID.String(),
		Title:    row.Title,
		Position: int(row.Position),
		Items:    []ItemDetail{},
	}, nil
}

func (r *sqlcRepository) CreateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string, req CreateItemRequest) (ItemDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	assessUUID, err := toUUID(assessmentID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid assessment id: %w", err)
	}
	sectionUUID, err := toUUID(sectionID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid section id: %w", err)
	}
	qvUUID, err := toUUID(req.QuestionVersionID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid question version id: %w", err)
	}
	points := "1.00"
	if req.Points != "" {
		points = req.Points
	}
	pointsNum, err := toNumeric(points)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid points: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateAssessmentItem(ctx, assessmentsqlc.CreateAssessmentItemParams{
		OrganizationID:      orgUUID,
		AssessmentID:        assessUUID,
		AssessmentSectionID: sectionUUID,
		QuestionVersionID:   qvUUID,
		Position:            int32(req.Position),
		Points:              pointsNum,
	})
	if err != nil {
		return ItemDetail{}, fmt.Errorf("create assessment item: %w", err)
	}
	return ItemDetail{
		ID:                row.ID.String(),
		QuestionVersionID: row.QuestionVersionID.String(),
		Position:          int(row.Position),
		Points:            numericString(row.Points),
	}, nil
}

func itemDetailFromCreateRow(row assessmentsqlc.CreateAssessmentItemRow) ItemDetail {
	return ItemDetail{
		ID:                row.ID.String(),
		QuestionVersionID: row.QuestionVersionID.String(),
		Position:          int(row.Position),
		Points:            numericString(row.Points),
	}
}

func (r *sqlcRepository) DuplicateSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string) (SectionDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	assessUUID, err := toUUID(assessmentID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid assessment id: %w", err)
	}
	sectionUUID, err := toUUID(sectionID)
	if err != nil {
		return SectionDetail{}, fmt.Errorf("invalid section id: %w", err)
	}

	source, err := r.queries.WithTx(tx).GetAssessmentSection(ctx, assessmentsqlc.GetAssessmentSectionParams{
		SectionID:      sectionUUID,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return SectionDetail{}, ErrNotFound
	}
	if err != nil {
		return SectionDetail{}, fmt.Errorf("get source section: %w", err)
	}
	if source.AssessmentID != assessUUID || source.Status != "ACTIVE" {
		return SectionDetail{}, ErrNotFound
	}

	sections, err := r.queries.WithTx(tx).GetAssessmentSections(ctx, assessmentsqlc.GetAssessmentSectionsParams{
		OrganizationID: orgUUID,
		AssessmentID:   assessUUID,
	})
	if err != nil {
		return SectionDetail{}, fmt.Errorf("list sections: %w", err)
	}
	maxPos := 0
	for _, sec := range sections {
		if int(sec.Position) > maxPos {
			maxPos = int(sec.Position)
		}
	}

	newSection, err := r.queries.WithTx(tx).CreateAssessmentSection(ctx, assessmentsqlc.CreateAssessmentSectionParams{
		OrganizationID: orgUUID,
		AssessmentID:   assessUUID,
		Title:          source.Title + " (copy)",
		Position:       int32(maxPos + 10),
	})
	if err != nil {
		return SectionDetail{}, fmt.Errorf("create duplicated section: %w", err)
	}

	sourceItems, err := r.queries.WithTx(tx).GetAssessmentItemsBySection(ctx, assessmentsqlc.GetAssessmentItemsBySectionParams{
		OrganizationID: orgUUID,
		SectionID:      sectionUUID,
	})
	if err != nil {
		return SectionDetail{}, fmt.Errorf("list source items: %w", err)
	}

	items := make([]ItemDetail, len(sourceItems))
	for i, it := range sourceItems {
		row, err := r.queries.WithTx(tx).CreateAssessmentItem(ctx, assessmentsqlc.CreateAssessmentItemParams{
			OrganizationID:      orgUUID,
			AssessmentID:        assessUUID,
			AssessmentSectionID: newSection.ID,
			QuestionVersionID:   it.QuestionVersionID,
			Position:            it.Position,
			Points:              it.Points,
		})
		if err != nil {
			return SectionDetail{}, fmt.Errorf("duplicate item: %w", err)
		}
		items[i] = itemDetailFromCreateRow(row)
	}

	var settings json.RawMessage
	if len(newSection.SettingsJson) > 0 {
		settings = newSection.SettingsJson
	}
	return SectionDetail{
		ID:       newSection.ID.String(),
		Title:    newSection.Title,
		Position: int(newSection.Position),
		Settings: settings,
		Items:    items,
	}, nil
}

func (r *sqlcRepository) DuplicateItem(ctx context.Context, tx pgx.Tx, orgID, sectionID, itemID string) (ItemDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	sectionUUID, err := toUUID(sectionID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid section id: %w", err)
	}
	itemUUID, err := toUUID(itemID)
	if err != nil {
		return ItemDetail{}, fmt.Errorf("invalid item id: %w", err)
	}

	source, err := r.queries.WithTx(tx).GetAssessmentItem(ctx, assessmentsqlc.GetAssessmentItemParams{
		ItemID:         itemUUID,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ItemDetail{}, ErrNotFound
	}
	if err != nil {
		return ItemDetail{}, fmt.Errorf("get source item: %w", err)
	}
	if source.AssessmentSectionID != sectionUUID || source.Status != "ACTIVE" {
		return ItemDetail{}, ErrNotFound
	}

	assessmentID, err := r.queries.WithTx(tx).GetSectionAssessmentID(ctx, assessmentsqlc.GetSectionAssessmentIDParams{
		SectionID:      sectionUUID,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return ItemDetail{}, fmt.Errorf("get section assessment id: %w", err)
	}

	sectionItems, err := r.queries.WithTx(tx).GetAssessmentItemsBySection(ctx, assessmentsqlc.GetAssessmentItemsBySectionParams{
		OrganizationID: orgUUID,
		SectionID:      sectionUUID,
	})
	if err != nil {
		return ItemDetail{}, fmt.Errorf("list section items: %w", err)
	}
	maxPos := 0
	for _, it := range sectionItems {
		if int(it.Position) > maxPos {
			maxPos = int(it.Position)
		}
	}

	row, err := r.queries.WithTx(tx).CreateAssessmentItem(ctx, assessmentsqlc.CreateAssessmentItemParams{
		OrganizationID:      orgUUID,
		AssessmentID:        assessmentID,
		AssessmentSectionID: sectionUUID,
		QuestionVersionID:   source.QuestionVersionID,
		Position:            int32(maxPos + 10),
		Points:              source.Points,
	})
	if err != nil {
		return ItemDetail{}, fmt.Errorf("create duplicated item: %w", err)
	}
	return itemDetailFromCreateRow(row), nil
}

func (r *sqlcRepository) CreateAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateTargetRequest) (TargetDetail, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return TargetDetail{}, fmt.Errorf("invalid organization id: %w", err)
	}
	assessUUID, err := toUUID(assessmentID)
	if err != nil {
		return TargetDetail{}, fmt.Errorf("invalid assessment id: %w", err)
	}
	classUUID, err := toUUID(req.ClassSectionID)
	if err != nil {
		return TargetDetail{}, fmt.Errorf("invalid class id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateAssessmentTarget(ctx, assessmentsqlc.CreateAssessmentTargetParams{
		OrganizationID: orgUUID,
		AssessmentID:   assessUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return TargetDetail{}, fmt.Errorf("create assessment target: %w", err)
	}
	return TargetDetail{
		ID:             row.ID.String(),
		ClassSectionID: row.ClassSectionID.String(),
	}, nil
}

func (r *sqlcRepository) PublishAssessment(ctx context.Context, tx pgx.Tx, orgID, assessmentID, newStatus string) (int, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return 0, fmt.Errorf("invalid assessment id: %w", err)
	}
	rows, err := r.queries.WithTx(tx).PublishAssessment(ctx, assessmentsqlc.PublishAssessmentParams{
		NewStatus:      newStatus,
		ID:             id,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("publish assessment: %w", err)
	}
	if rows == 0 {
		return 0, ErrNotFound
	}
	rev, err := r.queries.WithTx(tx).GetAssessmentRevision(ctx, assessmentsqlc.GetAssessmentRevisionParams{
		ID:             id,
		OrganizationID: orgUUID,
	})
	if err != nil {
		return 0, fmt.Errorf("get assessment revision: %w", err)
	}
	return int(rev), nil
}

func (r *sqlcRepository) InsertAssessmentPublication(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, version int, snapshot json.RawMessage) error {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return fmt.Errorf("invalid assessment id: %w", err)
	}
	_, err = r.queries.WithTx(tx).InsertAssessmentPublication(ctx, assessmentsqlc.InsertAssessmentPublicationParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
		Version:        int32(version),
		SnapshotJson:   snapshot,
	})
	if err != nil {
		return fmt.Errorf("insert assessment publication: %w", err)
	}
	return nil
}

func (r *sqlcRepository) CountAssessmentSections(ctx context.Context, orgID, assessmentID string) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return 0, fmt.Errorf("invalid assessment id: %w", err)
	}
	count, err := r.queries.CountAssessmentSections(ctx, assessmentsqlc.CountAssessmentSectionsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return 0, fmt.Errorf("count sections: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) CountAssessmentItems(ctx context.Context, orgID, assessmentID string) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return 0, fmt.Errorf("invalid assessment id: %w", err)
	}
	count, err := r.queries.CountAssessmentItems(ctx, assessmentsqlc.CountAssessmentItemsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return 0, fmt.Errorf("count items: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) CountAssessmentTargets(ctx context.Context, orgID, assessmentID string) (int64, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return 0, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(assessmentID)
	if err != nil {
		return 0, fmt.Errorf("invalid assessment id: %w", err)
	}
	count, err := r.queries.CountAssessmentTargets(ctx, assessmentsqlc.CountAssessmentTargetsParams{
		OrganizationID: orgUUID,
		AssessmentID:   id,
	})
	if err != nil {
		return 0, fmt.Errorf("count targets: %w", err)
	}
	return count, nil
}

func (r *sqlcRepository) QuestionVersionExists(ctx context.Context, orgID, questionVersionID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	qvUUID, err := toUUID(questionVersionID)
	if err != nil {
		return false, fmt.Errorf("invalid question version id: %w", err)
	}
	exists, err := r.queries.QuestionVersionExists(ctx, assessmentsqlc.QuestionVersionExistsParams{
		QuestionVersionID: qvUUID,
		OrganizationID:    orgUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check question version: %w", err)
	}
	return exists, nil
}

func (r *sqlcRepository) IsClassManager(ctx context.Context, orgID, userID, classSectionID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user id: %w", err)
	}
	classUUID, err := toUUID(classSectionID)
	if err != nil {
		return false, fmt.Errorf("invalid class id: %w", err)
	}
	ok, err := r.queries.IsClassManager(ctx, assessmentsqlc.IsClassManagerParams{
		OrganizationID: orgUUID,
		UserID:         userUUID,
		ClassSectionID: classUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check class manager: %w", err)
	}
	return ok.Bool, nil
}

func (r *sqlcRepository) IsAssessmentManager(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return false, fmt.Errorf("invalid organization id: %w", err)
	}
	userUUID, err := toUUID(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user id: %w", err)
	}
	assessUUID, err := toUUID(assessmentID)
	if err != nil {
		return false, fmt.Errorf("invalid assessment id: %w", err)
	}
	ok, err := r.queries.IsAssessmentManager(ctx, assessmentsqlc.IsAssessmentManagerParams{
		OrganizationID: orgUUID,
		UserID:         userUUID,
		AssessmentID:   assessUUID,
	})
	if err != nil {
		return false, fmt.Errorf("check assessment manager: %w", err)
	}
	return ok.Bool, nil
}

func (r *sqlcRepository) TransitionAssessmentsToOpen(ctx context.Context) (int64, error) {
	n, err := r.queries.TransitionAssessmentsToOpen(ctx)
	if err != nil {
		return 0, fmt.Errorf("transition assessments to open: %w", err)
	}
	return n, nil
}

func (r *sqlcRepository) TransitionAssessmentsToClosed(ctx context.Context) (int64, error) {
	n, err := r.queries.TransitionAssessmentsToClosed(ctx)
	if err != nil {
		return 0, fmt.Errorf("transition assessments to closed: %w", err)
	}
	return n, nil
}

// ----- Question bank editor -----

func questionBankFromRow(row assessmentsqlc.QuestionBank) QuestionBank {
	return QuestionBank{
		ID:             row.ID.String(),
		OrganizationID: row.OrganizationID.String(),
		Title:          row.Title,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:      row.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}
}

func (r *sqlcRepository) CreateQuestionBank(ctx context.Context, tx pgx.Tx, orgID, title string) (QuestionBank, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return QuestionBank{}, fmt.Errorf("invalid organization id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateQuestionBank(ctx, assessmentsqlc.CreateQuestionBankParams{
		OrganizationID: orgUUID,
		Title:          title,
	})
	if err != nil {
		return QuestionBank{}, fmt.Errorf("create question bank: %w", err)
	}
	return questionBankFromRow(row), nil
}

func (r *sqlcRepository) ListQuestionBanksByOrganization(ctx context.Context, orgID string, opts ListQuestionBanksOptions) ([]QuestionBank, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization id: %w", err)
	}
	rows, err := r.queries.ListQuestionBanksByOrganization(ctx, assessmentsqlc.ListQuestionBanksByOrganizationParams{
		OrganizationID:  orgUUID,
		IncludeArchived: opts.IncludeArchived,
		PageOffset:      int32(opts.Offset),
		PageLimit:       int32(opts.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list question banks: %w", err)
	}
	banks := make([]QuestionBank, len(rows))
	for i, row := range rows {
		banks[i] = questionBankFromRow(row)
	}
	return banks, nil
}

func (r *sqlcRepository) GetQuestionBank(ctx context.Context, orgID, bankID string) (QuestionBank, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return QuestionBank{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(bankID)
	if err != nil {
		return QuestionBank{}, fmt.Errorf("invalid bank id: %w", err)
	}
	row, err := r.queries.GetQuestionBank(ctx, assessmentsqlc.GetQuestionBankParams{
		ID:             id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionBank{}, ErrNotFound
	}
	if err != nil {
		return QuestionBank{}, fmt.Errorf("get question bank: %w", err)
	}
	return questionBankFromRow(row), nil
}

func (r *sqlcRepository) CreateQuestion(ctx context.Context, tx pgx.Tx, bankID string) (QuestionBankQuestion, error) {
	bankUUID, err := toUUID(bankID)
	if err != nil {
		return QuestionBankQuestion{}, fmt.Errorf("invalid bank id: %w", err)
	}
	row, err := r.queries.WithTx(tx).CreateQuestion(ctx, bankUUID)
	if err != nil {
		return QuestionBankQuestion{}, fmt.Errorf("create question: %w", err)
	}
	return QuestionBankQuestion{
		ID:             row.ID.String(),
		QuestionBankID: row.QuestionBankID.String(),
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:      row.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}

func (r *sqlcRepository) ListQuestionsInBank(ctx context.Context, bankID string, opts ListQuestionBanksOptions) ([]QuestionBankQuestion, error) {
	bankUUID, err := toUUID(bankID)
	if err != nil {
		return nil, fmt.Errorf("invalid bank id: %w", err)
	}
	rows, err := r.queries.ListQuestionsInBank(ctx, assessmentsqlc.ListQuestionsInBankParams{
		BankID:          bankUUID,
		IncludeArchived: opts.IncludeArchived,
		PageOffset:      int32(opts.Offset),
		PageLimit:       int32(opts.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list questions in bank: %w", err)
	}
	items := make([]QuestionBankQuestion, len(rows))
	for i, row := range rows {
		item := QuestionBankQuestion{
			ID:             row.ID.String(),
			QuestionBankID: row.QuestionBankID.String(),
			Status:         row.Status,
			CreatedAt:      row.CreatedAt.Time.UTC().Format(time.RFC3339),
			UpdatedAt:      row.UpdatedAt.Time.UTC().Format(time.RFC3339),
		}
		if row.LatestVersionID.Valid {
			s := row.LatestVersionID.String()
			item.LatestVersionID = &s
		}
		if row.LatestVersionStatus != "" {
			s := row.LatestVersionStatus
			item.LatestVersionStatus = &s
		}
		if row.LatestVersion > 0 {
			n := int(row.LatestVersion)
			item.LatestVersion = &n
		}
		if row.QuestionType != "" {
			s := row.QuestionType
			item.QuestionType = &s
		}
		items[i] = item
	}
	return items, nil
}

func (r *sqlcRepository) GetQuestionWithBank(ctx context.Context, questionID string) (QuestionBankQuestion, string, error) {
	id, err := toUUID(questionID)
	if err != nil {
		return QuestionBankQuestion{}, "", fmt.Errorf("invalid question id: %w", err)
	}
	row, err := r.queries.GetQuestionWithBank(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionBankQuestion{}, "", ErrNotFound
	}
	if err != nil {
		return QuestionBankQuestion{}, "", fmt.Errorf("get question with bank: %w", err)
	}
	bankID := row.QuestionBankID.String()
	orgID := row.OrganizationID.String()
	return QuestionBankQuestion{
		ID:             row.ID.String(),
		QuestionBankID: bankID,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:      row.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}, orgID, nil
}

func (r *sqlcRepository) GetQuestion(ctx context.Context, bankID, questionID string) (QuestionBankQuestion, error) {
	bankUUID, err := toUUID(bankID)
	if err != nil {
		return QuestionBankQuestion{}, fmt.Errorf("invalid bank id: %w", err)
	}
	id, err := toUUID(questionID)
	if err != nil {
		return QuestionBankQuestion{}, fmt.Errorf("invalid question id: %w", err)
	}
	row, err := r.queries.GetQuestion(ctx, assessmentsqlc.GetQuestionParams{
		BankID: bankUUID,
		ID:     id,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionBankQuestion{}, ErrNotFound
	}
	if err != nil {
		return QuestionBankQuestion{}, fmt.Errorf("get question: %w", err)
	}
	return QuestionBankQuestion{
		ID:             row.ID.String(),
		QuestionBankID: row.QuestionBankID.String(),
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:      row.UpdatedAt.Time.UTC().Format(time.RFC3339),
	}, nil
}

func (r *sqlcRepository) CreateQuestionVersion(ctx context.Context, tx pgx.Tx, questionID string, req CreateQuestionVersionRequest, maxScore string, version int) (QuestionVersion, error) {
	qUUID, err := toUUID(questionID)
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("invalid question id: %w", err)
	}
	maxScoreNum, err := toNumeric(maxScore)
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("invalid max score: %w", err)
	}
	prompt := []byte(req.Prompt)
	if len(prompt) == 0 {
		prompt = []byte("{}")
	}
	var choices []byte
	if len(req.Choices) > 0 {
		choices = []byte(req.Choices)
	}
	var answerKey []byte
	if len(req.AnswerKey) > 0 {
		answerKey = []byte(req.AnswerKey)
	}
	status := "DRAFT"
	if req.Publish {
		status = "PUBLISHED"
	}
	row, err := r.queries.WithTx(tx).CreateQuestionVersion(ctx, assessmentsqlc.CreateQuestionVersionParams{
		QuestionID:    qUUID,
		Version:       int32(version),
		PromptJson:    prompt,
		ChoicesJson:   choices,
		AnswerKeyJson: answerKey,
		MaxScore:      maxScoreNum,
		Status:        status,
		QuestionType:  req.QuestionType,
	})
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("create question version: %w", err)
	}
	return questionVersionFromRow(row.ID, row.QuestionID, row.Version, row.QuestionType, row.PromptJson, row.ChoicesJson, row.AnswerKeyJson, row.MaxScore, row.Status, row.CreatedAt), nil
}

func (r *sqlcRepository) GetQuestionVersion(ctx context.Context, orgID, versionID string) (QuestionVersion, error) {
	orgUUID, err := toUUID(orgID)
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("invalid organization id: %w", err)
	}
	id, err := toUUID(versionID)
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("invalid version id: %w", err)
	}
	row, err := r.queries.GetQuestionVersion(ctx, assessmentsqlc.GetQuestionVersionParams{
		VersionID:      id,
		OrganizationID: orgUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionVersion{}, ErrNotFound
	}
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("get question version: %w", err)
	}
	return questionVersionFromRow(row.ID, row.QuestionID, row.Version, row.QuestionType, row.PromptJson, row.ChoicesJson, row.AnswerKeyJson, row.MaxScore, row.Status, row.CreatedAt), nil
}

func (r *sqlcRepository) GetLatestVersionNumber(ctx context.Context, tx pgx.Tx, questionID string) (int, error) {
	qUUID, err := toUUID(questionID)
	if err != nil {
		return 0, fmt.Errorf("invalid question id: %w", err)
	}
	n, err := r.queries.WithTx(tx).GetLatestVersionNumber(ctx, qUUID)
	if err != nil {
		return 0, fmt.Errorf("get latest version number: %w", err)
	}
	return int(n), nil
}

func (r *sqlcRepository) PublishQuestionVersion(ctx context.Context, tx pgx.Tx, versionID string) (QuestionVersion, error) {
	id, err := toUUID(versionID)
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("invalid version id: %w", err)
	}
	row, err := r.queries.WithTx(tx).PublishQuestionVersion(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return QuestionVersion{}, ErrNotFound
	}
	if err != nil {
		return QuestionVersion{}, fmt.Errorf("publish question version: %w", err)
	}
	return questionVersionFromRow(row.ID, row.QuestionID, row.Version, row.QuestionType, row.PromptJson, row.ChoicesJson, row.AnswerKeyJson, row.MaxScore, row.Status, row.CreatedAt), nil
}

func questionVersionFromRow(
	id, questionID pgtype.UUID,
	version int32,
	questionType string,
	prompt, choices, answerKey []byte,
	maxScore pgtype.Numeric,
	status string,
	createdAt pgtype.Timestamptz,
) QuestionVersion {
	var promptJSON json.RawMessage
	if len(prompt) > 0 {
		promptJSON = prompt
	} else {
		promptJSON = json.RawMessage("{}")
	}
	var choicesJSON json.RawMessage
	if len(choices) > 0 {
		choicesJSON = choices
	}
	var answerKeyJSON json.RawMessage
	if len(answerKey) > 0 {
		answerKeyJSON = answerKey
	}
	return QuestionVersion{
		ID:           id.String(),
		QuestionID:   questionID.String(),
		Version:      int(version),
		QuestionType: questionType,
		Prompt:       promptJSON,
		Choices:      choicesJSON,
		AnswerKey:    answerKeyJSON,
		MaxScore:     numericString(maxScore),
		Status:       status,
		CreatedAt:    createdAt.Time.UTC().Format(time.RFC3339),
	}
}
