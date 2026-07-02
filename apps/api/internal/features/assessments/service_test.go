package assessments

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/jackc/pgx/v5"
)

type fakeRepo struct {
	createAssessmentFunc              func(ctx context.Context, tx pgx.Tx, orgID, classSectionID, title string, durationMinutes, maxAttempts int) (AssessmentDetail, error)
	listAssessmentsByClassFunc        func(ctx context.Context, orgID, classSectionID string) ([]AssessmentListItem, error)
	getAssessmentFunc                 func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error)
	updateAssessmentSettingsFunc      func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req UpdateAssessmentRequest) error
	createSectionFunc                 func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateSectionRequest) (SectionDetail, error)
	createItemFunc                    func(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string, req CreateItemRequest) (ItemDetail, error)
	createTargetFunc                  func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateTargetRequest) (TargetDetail, error)
	duplicateSectionFunc              func(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string) (SectionDetail, error)
	duplicateItemFunc                 func(ctx context.Context, tx pgx.Tx, orgID, sectionID, itemID string) (ItemDetail, error)
	getSectionAssessmentIDFunc        func(ctx context.Context, orgID, sectionID string) (string, error)
	updateAssessmentSectionFunc       func(ctx context.Context, tx pgx.Tx, orgID, sectionID string, req UpdateSectionRequest) (SectionDetail, error)
	archiveAssessmentSectionFunc      func(ctx context.Context, tx pgx.Tx, orgID, sectionID string) error
	getItemAssessmentIDFunc           func(ctx context.Context, orgID, itemID string) (string, error)
	updateAssessmentItemFunc          func(ctx context.Context, tx pgx.Tx, orgID, itemID string, req UpdateItemRequest) (ItemDetail, error)
	archiveAssessmentItemFunc         func(ctx context.Context, tx pgx.Tx, orgID, itemID string) error
	getTargetAssessmentIDFunc         func(ctx context.Context, orgID, targetID string) (string, error)
	archiveAssessmentTargetFunc       func(ctx context.Context, tx pgx.Tx, orgID, targetID string) error
	reorderAssessmentSectionsFunc     func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, sectionIDs []string) error
	reorderAssessmentItemsFunc        func(ctx context.Context, tx pgx.Tx, orgID, sectionID string, itemIDs []string) error
	listQuestionsFunc                 func(ctx context.Context, orgID string, opts ListQuestionsOptions) ([]QuestionPickerItem, error)
	countQuestionsFunc                func(ctx context.Context, orgID string, opts ListQuestionsOptions) (int64, error)
	listAssessmentPublicationsFunc    func(ctx context.Context, orgID, assessmentID string) ([]PublicationSummary, error)
	questionVersionExistsFunc         func(ctx context.Context, orgID, questionVersionID string) (bool, error)
	isQuestionVersionPublishedFunc    func(ctx context.Context, orgID, questionVersionID string) (bool, error)
	isClassSectionActiveFunc          func(ctx context.Context, orgID, classSectionID string) (bool, error)
	countSectionsFunc                 func(ctx context.Context, orgID, assessmentID string) (int64, error)
	countItemsFunc                    func(ctx context.Context, orgID, assessmentID string) (int64, error)
	countTargetsFunc                  func(ctx context.Context, orgID, assessmentID string) (int64, error)
	publishAssessmentFunc             func(ctx context.Context, tx pgx.Tx, orgID, assessmentID, newStatus string) (int, error)
	insertPublicationFunc             func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, version int, snapshot json.RawMessage) error
	getAssessmentItemsWithContentFunc func(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error)
	getAssessmentSectionsFunc         func(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error)
	getAssessmentItemsFunc            func(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error)
	getAssessmentTargetsFunc          func(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error)
	isClassManagerFunc                func(ctx context.Context, orgID, userID, classSectionID string) (bool, error)
	isAssessmentManagerFunc           func(ctx context.Context, orgID, userID, assessmentID string) (bool, error)
}

func (f *fakeRepo) ListPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) ([]Assessment, error) {
	return nil, nil
}

func (f *fakeRepo) CountPublishedByOrganization(ctx context.Context, orgID string, opts ListOptions) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) CreateAssessment(ctx context.Context, tx pgx.Tx, orgID, classSectionID, title string, durationMinutes, maxAttempts int) (AssessmentDetail, error) {
	if f.createAssessmentFunc != nil {
		return f.createAssessmentFunc(ctx, tx, orgID, classSectionID, title, durationMinutes, maxAttempts)
	}
	return AssessmentDetail{}, nil
}

func (f *fakeRepo) ListAssessmentsByClass(ctx context.Context, orgID, classSectionID string) ([]AssessmentListItem, error) {
	if f.listAssessmentsByClassFunc != nil {
		return f.listAssessmentsByClassFunc(ctx, orgID, classSectionID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessment(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
	if f.getAssessmentFunc != nil {
		return f.getAssessmentFunc(ctx, orgID, assessmentID)
	}
	return AssessmentDetail{}, nil
}

func (f *fakeRepo) GetSectionAssessmentID(ctx context.Context, orgID, sectionID string) (string, error) {
	if f.getSectionAssessmentIDFunc != nil {
		return f.getSectionAssessmentIDFunc(ctx, orgID, sectionID)
	}
	return "", nil
}

func (f *fakeRepo) GetAssessmentSections(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
	if f.getAssessmentSectionsFunc != nil {
		return f.getAssessmentSectionsFunc(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessmentItems(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error) {
	if f.getAssessmentItemsFunc != nil {
		return f.getAssessmentItemsFunc(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessmentItemsWithContent(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error) {
	if f.getAssessmentItemsWithContentFunc != nil {
		return f.getAssessmentItemsWithContentFunc(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessmentTargets(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
	if f.getAssessmentTargetsFunc != nil {
		return f.getAssessmentTargetsFunc(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) GetAssessmentSection(ctx context.Context, orgID, sectionID string) (SectionDetail, error) {
	return SectionDetail{}, nil
}

func (f *fakeRepo) UpdateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string, req UpdateSectionRequest) (SectionDetail, error) {
	if f.updateAssessmentSectionFunc != nil {
		return f.updateAssessmentSectionFunc(ctx, tx, orgID, sectionID, req)
	}
	return SectionDetail{}, nil
}

func (f *fakeRepo) ArchiveAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, sectionID string) error {
	if f.archiveAssessmentSectionFunc != nil {
		return f.archiveAssessmentSectionFunc(ctx, tx, orgID, sectionID)
	}
	return nil
}

func (f *fakeRepo) GetAssessmentItem(ctx context.Context, orgID, itemID string) (ItemDetail, error) {
	return ItemDetail{}, nil
}

func (f *fakeRepo) GetItemAssessmentID(ctx context.Context, orgID, itemID string) (string, error) {
	if f.getItemAssessmentIDFunc != nil {
		return f.getItemAssessmentIDFunc(ctx, orgID, itemID)
	}
	return "", nil
}

func (f *fakeRepo) UpdateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string, req UpdateItemRequest) (ItemDetail, error) {
	if f.updateAssessmentItemFunc != nil {
		return f.updateAssessmentItemFunc(ctx, tx, orgID, itemID, req)
	}
	return ItemDetail{}, nil
}

func (f *fakeRepo) ArchiveAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, itemID string) error {
	if f.archiveAssessmentItemFunc != nil {
		return f.archiveAssessmentItemFunc(ctx, tx, orgID, itemID)
	}
	return nil
}

func (f *fakeRepo) GetAssessmentTarget(ctx context.Context, orgID, targetID string) (TargetDetail, error) {
	return TargetDetail{}, nil
}

func (f *fakeRepo) GetTargetAssessmentID(ctx context.Context, orgID, targetID string) (string, error) {
	if f.getTargetAssessmentIDFunc != nil {
		return f.getTargetAssessmentIDFunc(ctx, orgID, targetID)
	}
	return "", nil
}

func (f *fakeRepo) ArchiveAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, targetID string) error {
	if f.archiveAssessmentTargetFunc != nil {
		return f.archiveAssessmentTargetFunc(ctx, tx, orgID, targetID)
	}
	return nil
}

func (f *fakeRepo) ReorderAssessmentSections(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, sectionIDs []string) error {
	if f.reorderAssessmentSectionsFunc != nil {
		return f.reorderAssessmentSectionsFunc(ctx, tx, orgID, assessmentID, sectionIDs)
	}
	return nil
}

func (f *fakeRepo) ReorderAssessmentItems(ctx context.Context, tx pgx.Tx, orgID, sectionID string, itemIDs []string) error {
	if f.reorderAssessmentItemsFunc != nil {
		return f.reorderAssessmentItemsFunc(ctx, tx, orgID, sectionID, itemIDs)
	}
	return nil
}

func (f *fakeRepo) ListQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) ([]QuestionPickerItem, error) {
	if f.listQuestionsFunc != nil {
		return f.listQuestionsFunc(ctx, orgID, opts)
	}
	return nil, nil
}

func (f *fakeRepo) CountQuestions(ctx context.Context, orgID string, opts ListQuestionsOptions) (int64, error) {
	if f.countQuestionsFunc != nil {
		return f.countQuestionsFunc(ctx, orgID, opts)
	}
	return 0, nil
}

func (f *fakeRepo) ListAssessmentPublications(ctx context.Context, orgID, assessmentID string) ([]PublicationSummary, error) {
	if f.listAssessmentPublicationsFunc != nil {
		return f.listAssessmentPublicationsFunc(ctx, orgID, assessmentID)
	}
	return nil, nil
}

func (f *fakeRepo) IsQuestionVersionPublished(ctx context.Context, orgID, questionVersionID string) (bool, error) {
	if f.isQuestionVersionPublishedFunc != nil {
		return f.isQuestionVersionPublishedFunc(ctx, orgID, questionVersionID)
	}
	return false, nil
}

func (f *fakeRepo) IsClassSectionActive(ctx context.Context, orgID, classSectionID string) (bool, error) {
	if f.isClassSectionActiveFunc != nil {
		return f.isClassSectionActiveFunc(ctx, orgID, classSectionID)
	}
	return false, nil
}

func (f *fakeRepo) UpdateAssessmentSettings(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req UpdateAssessmentRequest) error {
	if f.updateAssessmentSettingsFunc != nil {
		return f.updateAssessmentSettingsFunc(ctx, tx, orgID, assessmentID, req)
	}
	return nil
}

func (f *fakeRepo) CreateAssessmentSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateSectionRequest) (SectionDetail, error) {
	if f.createSectionFunc != nil {
		return f.createSectionFunc(ctx, tx, orgID, assessmentID, req)
	}
	return SectionDetail{}, nil
}

func (f *fakeRepo) CreateAssessmentItem(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string, req CreateItemRequest) (ItemDetail, error) {
	if f.createItemFunc != nil {
		return f.createItemFunc(ctx, tx, orgID, assessmentID, sectionID, req)
	}
	return ItemDetail{}, nil
}

func (f *fakeRepo) CreateAssessmentTarget(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, req CreateTargetRequest) (TargetDetail, error) {
	if f.createTargetFunc != nil {
		return f.createTargetFunc(ctx, tx, orgID, assessmentID, req)
	}
	return TargetDetail{}, nil
}

func (f *fakeRepo) DuplicateSection(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string) (SectionDetail, error) {
	if f.duplicateSectionFunc != nil {
		return f.duplicateSectionFunc(ctx, tx, orgID, assessmentID, sectionID)
	}
	return SectionDetail{}, nil
}

func (f *fakeRepo) DuplicateItem(ctx context.Context, tx pgx.Tx, orgID, sectionID, itemID string) (ItemDetail, error) {
	if f.duplicateItemFunc != nil {
		return f.duplicateItemFunc(ctx, tx, orgID, sectionID, itemID)
	}
	return ItemDetail{}, nil
}

func (f *fakeRepo) PublishAssessment(ctx context.Context, tx pgx.Tx, orgID, assessmentID, newStatus string) (int, error) {
	if f.publishAssessmentFunc != nil {
		return f.publishAssessmentFunc(ctx, tx, orgID, assessmentID, newStatus)
	}
	return 0, nil
}

func (f *fakeRepo) InsertAssessmentPublication(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, version int, snapshot json.RawMessage) error {
	if f.insertPublicationFunc != nil {
		return f.insertPublicationFunc(ctx, tx, orgID, assessmentID, version, snapshot)
	}
	return nil
}

func (f *fakeRepo) CountAssessmentSections(ctx context.Context, orgID, assessmentID string) (int64, error) {
	if f.countSectionsFunc != nil {
		return f.countSectionsFunc(ctx, orgID, assessmentID)
	}
	return 0, nil
}

func (f *fakeRepo) CountAssessmentItems(ctx context.Context, orgID, assessmentID string) (int64, error) {
	if f.countItemsFunc != nil {
		return f.countItemsFunc(ctx, orgID, assessmentID)
	}
	return 0, nil
}

func (f *fakeRepo) CountAssessmentTargets(ctx context.Context, orgID, assessmentID string) (int64, error) {
	if f.countTargetsFunc != nil {
		return f.countTargetsFunc(ctx, orgID, assessmentID)
	}
	return 0, nil
}

func (f *fakeRepo) QuestionVersionExists(ctx context.Context, orgID, questionVersionID string) (bool, error) {
	if f.questionVersionExistsFunc != nil {
		return f.questionVersionExistsFunc(ctx, orgID, questionVersionID)
	}
	return false, nil
}

func (f *fakeRepo) IsClassManager(ctx context.Context, orgID, userID, classSectionID string) (bool, error) {
	if f.isClassManagerFunc != nil {
		return f.isClassManagerFunc(ctx, orgID, userID, classSectionID)
	}
	return false, nil
}

func (f *fakeRepo) IsAssessmentManager(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
	if f.isAssessmentManagerFunc != nil {
		return f.isAssessmentManagerFunc(ctx, orgID, userID, assessmentID)
	}
	return false, nil
}

func (f *fakeRepo) TransitionAssessmentsToOpen(ctx context.Context) (int64, error) {
	return 0, nil
}

func (f *fakeRepo) TransitionAssessmentsToClosed(ctx context.Context) (int64, error) {
	return 0, nil
}

// Question bank editor stubs
func (f *fakeRepo) CreateQuestionBank(ctx context.Context, tx pgx.Tx, orgID, title string) (QuestionBank, error) {
	return QuestionBank{ID: "bank-id", OrganizationID: orgID, Title: title, Status: "ACTIVE"}, nil
}

func (f *fakeRepo) ListQuestionBanksByOrganization(ctx context.Context, orgID string, opts ListQuestionBanksOptions) ([]QuestionBank, error) {
	return nil, nil
}

func (f *fakeRepo) GetQuestionBank(ctx context.Context, orgID, bankID string) (QuestionBank, error) {
	return QuestionBank{ID: bankID, OrganizationID: orgID, Status: "ACTIVE"}, nil
}

func (f *fakeRepo) CreateQuestion(ctx context.Context, tx pgx.Tx, bankID string) (QuestionBankQuestion, error) {
	return QuestionBankQuestion{ID: "q-id", QuestionBankID: bankID, Status: "ACTIVE"}, nil
}

func (f *fakeRepo) ListQuestionsInBank(ctx context.Context, bankID string, opts ListQuestionBanksOptions) ([]QuestionBankQuestion, error) {
	return nil, nil
}

func (f *fakeRepo) GetQuestionWithBank(ctx context.Context, questionID string) (QuestionBankQuestion, string, error) {
	return QuestionBankQuestion{ID: questionID, Status: "ACTIVE"}, "org-id", nil
}

func (f *fakeRepo) GetQuestion(ctx context.Context, bankID, questionID string) (QuestionBankQuestion, error) {
	return QuestionBankQuestion{ID: questionID, QuestionBankID: bankID, Status: "ACTIVE"}, nil
}

func (f *fakeRepo) CreateQuestionVersion(ctx context.Context, tx pgx.Tx, questionID string, req CreateQuestionVersionRequest, maxScore string, version int) (QuestionVersion, error) {
	return QuestionVersion{ID: "v-id", QuestionID: questionID, Version: version, QuestionType: req.QuestionType, Status: "DRAFT"}, nil
}

func (f *fakeRepo) GetQuestionVersion(ctx context.Context, orgID, versionID string) (QuestionVersion, error) {
	return QuestionVersion{ID: versionID, QuestionID: "q-id", Status: "DRAFT"}, nil
}

func (f *fakeRepo) GetLatestVersionNumber(ctx context.Context, tx pgx.Tx, questionID string) (int, error) {
	return 0, nil
}

func (f *fakeRepo) PublishQuestionVersion(ctx context.Context, tx pgx.Tx, versionID string) (QuestionVersion, error) {
	return QuestionVersion{ID: versionID, Status: "PUBLISHED"}, nil
}

type stubTxManager struct{}

func (stubTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	return fn(ctx, nil)
}

func TestService_CreateAssessment_Unauthorized(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateAssessment(context.Background(), auth.Actor{Roles: []string{"student"}}, "class-1", CreateAssessmentRequest{Title: "Test", DurationMinutes: 30})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_CreateAssessment_InvalidInput(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "class-1", CreateAssessmentRequest{Title: "", DurationMinutes: 30})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_ValidateAssessment_MissingSectionsItemsTargets(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT", DurationMinutes: 30, MaxAttempts: 1}, nil
		},
		getAssessmentSectionsFunc: func(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
			return nil, nil
		},
		getAssessmentItemsFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error) {
			return nil, nil
		},
		getAssessmentTargetsFunc: func(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
			return nil, nil
		},
		countSectionsFunc: func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 0, nil },
		countItemsFunc:    func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 0, nil },
		countTargetsFunc:  func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 0, nil },
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.ValidateAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected validation to fail")
	}
	if len(result.Errors) != 3 {
		t.Fatalf("expected 3 validation errors, got %d", len(result.Errors))
	}
}

func TestService_PublishAssessment_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Title: "Quiz", Status: "DRAFT", DurationMinutes: 30, MaxAttempts: 1, Revision: 1}, nil
		},
		countSectionsFunc: func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
		countItemsFunc:    func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
		countTargetsFunc:  func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
		getAssessmentSectionsFunc: func(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
			return []SectionDetail{{ID: "sec-1", Title: "Part A", Position: 1}}, nil
		},
		getAssessmentItemsFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error) {
			return []ItemDetail{{ID: "item-1", AssessmentSectionID: "sec-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00"}}, nil
		},
		getAssessmentItemsWithContentFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error) {
			return []ItemContentRow{{
				ID: "item-1", AssessmentSectionID: "sec-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00",
				Prompt: []byte(`{"text":"Q"}`), Choices: []byte(`[]`), AnswerKey: []byte(`{}`), MaxScore: "1.00",
			}}, nil
		},
		getAssessmentTargetsFunc: func(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
			return []TargetDetail{{ID: "tgt-1", ClassSectionID: "class-1"}}, nil
		},
		isQuestionVersionPublishedFunc: func(ctx context.Context, orgID, questionVersionID string) (bool, error) {
			return true, nil
		},
		isClassSectionActiveFunc: func(ctx context.Context, orgID, classSectionID string) (bool, error) {
			return true, nil
		},
		publishAssessmentFunc: func(ctx context.Context, tx pgx.Tx, orgID, assessmentID, newStatus string) (int, error) {
			return 2, nil
		},
		insertPublicationFunc: func(ctx context.Context, tx pgx.Tx, orgID, assessmentID string, version int, snapshot json.RawMessage) error {
			if version != 2 {
				t.Errorf("version = %d, want 2", version)
			}
			if len(snapshot) == 0 {
				t.Error("expected non-empty snapshot")
			}
			return nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	result, err := svc.PublishAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "OPEN" {
		t.Errorf("status = %q, want OPEN", result.Status)
	}
	if result.Revision != 2 {
		t.Errorf("revision = %d, want 2", result.Revision)
	}
}

func TestService_PublishAssessment_NotDraft(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "OPEN"}, nil
		},
		countSectionsFunc: func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
		countItemsFunc:    func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
		countTargetsFunc:  func(ctx context.Context, orgID, assessmentID string) (int64, error) { return 1, nil },
	}
	svc := NewService(repo, stubTxManager{})
	_, err := svc.PublishAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got %v", err)
	}
}

func TestService_UpdateSection_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getSectionAssessmentIDFunc: func(ctx context.Context, orgID, sectionID string) (string, error) {
			return "assess-1", nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		updateAssessmentSectionFunc: func(ctx context.Context, tx pgx.Tx, orgID, sectionID string, req UpdateSectionRequest) (SectionDetail, error) {
			return SectionDetail{ID: sectionID, Title: req.Title, Position: req.Position}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	section, err := svc.UpdateSection(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "sec-1", UpdateSectionRequest{Title: "Updated", Position: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section.Title != "Updated" {
		t.Errorf("title = %q, want Updated", section.Title)
	}
}

func TestService_UpdateItem_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getItemAssessmentIDFunc: func(ctx context.Context, orgID, itemID string) (string, error) {
			return "assess-1", nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		questionVersionExistsFunc: func(ctx context.Context, orgID, questionVersionID string) (bool, error) {
			return true, nil
		},
		updateAssessmentItemFunc: func(ctx context.Context, tx pgx.Tx, orgID, itemID string, req UpdateItemRequest) (ItemDetail, error) {
			return ItemDetail{ID: itemID, QuestionVersionID: req.QuestionVersionID, Points: req.Points, Position: req.Position}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	item, err := svc.UpdateItem(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "item-1", UpdateItemRequest{QuestionVersionID: "qv-2", Points: "2.00", Position: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.QuestionVersionID != "qv-2" {
		t.Errorf("question_version_id = %q, want qv-2", item.QuestionVersionID)
	}
}

func TestService_DeleteTarget_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getTargetAssessmentIDFunc: func(ctx context.Context, orgID, targetID string) (string, error) {
			return "assess-1", nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		archiveAssessmentTargetFunc: func(ctx context.Context, tx pgx.Tx, orgID, targetID string) error {
			return nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	if err := svc.DeleteTarget(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1", "tgt-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListPublications_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		listAssessmentPublicationsFunc: func(ctx context.Context, orgID, assessmentID string) ([]PublicationSummary, error) {
			return []PublicationSummary{{ID: "pub-1", Version: 1, Status: "OPEN", PublishedAt: "2026-06-30T00:00:00Z"}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	pubs, err := svc.ListPublications(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pubs) != 1 {
		t.Fatalf("expected 1 publication, got %d", len(pubs))
	}
}

func TestService_GetAssessment_NestsItemsUnderSections(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		getAssessmentSectionsFunc: func(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
			return []SectionDetail{
				{ID: "sec-1", Title: "Part A", Position: 1, Items: []ItemDetail{}},
				{ID: "sec-2", Title: "Part B", Position: 2, Items: []ItemDetail{}},
			}, nil
		},
		getAssessmentItemsFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemDetail, error) {
			return []ItemDetail{
				{ID: "item-1", AssessmentSectionID: "sec-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00"},
				{ID: "item-2", AssessmentSectionID: "sec-2", QuestionVersionID: "qv-2", Position: 1, Points: "2.00"},
			}, nil
		},
		getAssessmentTargetsFunc: func(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
			return []TargetDetail{{ID: "tgt-1", ClassSectionID: "class-1"}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	detail, err := svc.GetAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(detail.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(detail.Sections))
	}
	if len(detail.Sections[0].Items) != 1 || detail.Sections[0].Items[0].ID != "item-1" {
		t.Errorf("expected sec-1 to have item-1, got %+v", detail.Sections[0].Items)
	}
	if len(detail.Sections[1].Items) != 1 || detail.Sections[1].Items[0].ID != "item-2" {
		t.Errorf("expected sec-2 to have item-2, got %+v", detail.Sections[1].Items)
	}
	if len(detail.Targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(detail.Targets))
	}
}

func TestService_DuplicateSection_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		duplicateSectionFunc: func(ctx context.Context, tx pgx.Tx, orgID, assessmentID, sectionID string) (SectionDetail, error) {
			return SectionDetail{ID: "sec-copy", Title: "Part A (copy)", Position: 20, Items: []ItemDetail{{ID: "item-copy", QuestionVersionID: "qv-1", Position: 10, Points: "1.00"}}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	section, err := svc.DuplicateSection(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1", "sec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section.Title != "Part A (copy)" {
		t.Errorf("title = %q, want Part A (copy)", section.Title)
	}
	if len(section.Items) != 1 {
		t.Errorf("expected 1 duplicated item, got %d", len(section.Items))
	}
}

func TestService_DuplicateSection_NotDraft(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "OPEN"}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	_, err := svc.DuplicateSection(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1", "sec-1")
	if !errors.Is(err, ErrNotDraft) {
		t.Fatalf("expected ErrNotDraft, got %v", err)
	}
}

func TestService_DuplicateItem_Success(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getSectionAssessmentIDFunc: func(ctx context.Context, orgID, sectionID string) (string, error) {
			return "assess-1", nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
		},
		duplicateItemFunc: func(ctx context.Context, tx pgx.Tx, orgID, sectionID, itemID string) (ItemDetail, error) {
			return ItemDetail{ID: "item-copy", QuestionVersionID: "qv-1", Position: 20, Points: "1.00"}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	item, err := svc.DuplicateItem(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "sec-1", "item-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Position != 20 {
		t.Errorf("position = %d, want 20", item.Position)
	}
}

func TestService_PreviewAssessment_HidesAnswerKey(t *testing.T) {
	repo := &fakeRepo{
		isAssessmentManagerFunc: func(ctx context.Context, orgID, userID, assessmentID string) (bool, error) {
			return true, nil
		},
		getAssessmentFunc: func(ctx context.Context, orgID, assessmentID string) (AssessmentDetail, error) {
			return AssessmentDetail{ID: assessmentID, Title: "Quiz", Status: "DRAFT", DurationMinutes: 30, MaxAttempts: 1}, nil
		},
		getAssessmentSectionsFunc: func(ctx context.Context, orgID, assessmentID string) ([]SectionDetail, error) {
			return []SectionDetail{{ID: "sec-1", Title: "Part A", Position: 1, Items: []ItemDetail{}}}, nil
		},
		getAssessmentItemsWithContentFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error) {
			return []ItemContentRow{{
				ID: "item-1", AssessmentSectionID: "sec-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00",
				Prompt: []byte(`{"text":"Q"}`), Choices: []byte(`[{"id":"a","text":"A"}]`), AnswerKey: []byte(`{"correct":"a"}`), MaxScore: "1.00",
			}}, nil
		},
	}
	svc := NewService(repo, stubTxManager{})
	preview, err := svc.PreviewAssessment(context.Background(), auth.Actor{Roles: []string{"teacher"}}, "assess-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preview.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(preview.Sections))
	}
	if len(preview.Sections[0].Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(preview.Sections[0].Items))
	}
	item := preview.Sections[0].Items[0]
	if string(item.Prompt) == "" {
		t.Error("expected prompt to be present")
	}
	if string(item.Choices) == "" {
		t.Error("expected choices to be present")
	}
}

func TestService_CreateQuestionBank_EmptyTitle(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateQuestionBank(context.Background(), auth.Actor{Roles: []string{"teacher"}}, CreateQuestionBankRequest{Title: "  "})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_CreateQuestionBank_Unauthorized(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateQuestionBank(context.Background(), auth.Actor{Roles: []string{"student"}}, CreateQuestionBankRequest{Title: "Bank"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestService_CreateQuestion_MCQMissingChoices(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateQuestion(context.Background(), auth.Actor{Roles: []string{"teacher"}, OrgID: "org-id", UserID: "user-id"}, "bank-id", CreateQuestionRequest{
		QuestionType: "multiple_choice",
		Prompt:       json.RawMessage(`{"text":"x"}`),
		AnswerKey:    json.RawMessage(`{"correct_option":"A"}`),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_CreateQuestion_ShortAnswerMissingAcceptedAnswers(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateQuestion(context.Background(), auth.Actor{Roles: []string{"teacher"}, OrgID: "org-id", UserID: "user-id"}, "bank-id", CreateQuestionRequest{
		QuestionType: "short_answer",
		Prompt:       json.RawMessage(`{"text":"x"}`),
		AnswerKey:    json.RawMessage(`{}`),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestService_CreateQuestion_InvalidType(t *testing.T) {
	svc := NewService(&fakeRepo{}, stubTxManager{})
	_, err := svc.CreateQuestion(context.Background(), auth.Actor{Roles: []string{"teacher"}, OrgID: "org-id", UserID: "user-id"}, "bank-id", CreateQuestionRequest{
		QuestionType: "matching",
		Prompt:       json.RawMessage(`{"text":"x"}`),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
