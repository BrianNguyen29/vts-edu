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
	getSectionAssessmentIDFunc        func(ctx context.Context, orgID, sectionID string) (string, error)
	questionVersionExistsFunc         func(ctx context.Context, orgID, questionVersionID string) (bool, error)
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
			return AssessmentDetail{ID: assessmentID, Status: "DRAFT"}, nil
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
		getAssessmentItemsWithContentFunc: func(ctx context.Context, orgID, assessmentID string) ([]ItemContentRow, error) {
			return []ItemContentRow{{
				ID: "item-1", AssessmentSectionID: "sec-1", QuestionVersionID: "qv-1", Position: 1, Points: "1.00",
				Prompt: []byte(`{"text":"Q"}`), Choices: []byte(`[]`), AnswerKey: []byte(`{}`), MaxScore: "1.00",
			}}, nil
		},
		getAssessmentTargetsFunc: func(ctx context.Context, orgID, assessmentID string) ([]TargetDetail, error) {
			return []TargetDetail{{ID: "tgt-1", ClassSectionID: "class-1"}}, nil
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
