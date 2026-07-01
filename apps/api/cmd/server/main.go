package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/BrianNguyen29/vts-edu/apps/api/internal/app"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/academics"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/admin"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/assessments"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/attempts"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/auth"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/gradebook"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/features/resources"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/csrf"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/db"
	vtsmiddleware "github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/middleware"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/ratelimit"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/scheduler"
	"github.com/BrianNguyen29/vts-edu/apps/api/internal/platform/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type server struct {
	cfg       *app.Config
	dbPool    *db.Pool
	txManager *db.TxManager
}

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	limiter := ratelimit.New(cfg.RateLimitEnabled, cfg.RateLimitRPS, cfg.RateLimitBurst, cfg.RateLimitTTL, cfg.RateLimitCleanup)
	defer limiter.Stop()

	ctx := context.Background()
	var pool *db.Pool
	if !cfg.DatabaseSkip {
		pool, err = db.NewPool(ctx, cfg.DatabaseURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
		if err != nil {
			slog.Error("failed to connect to database", "error", err)
			os.Exit(1)
		}
		defer pool.Close()
		slog.Info("database pool initialized")
	} else {
		slog.Warn("database skipped via DB_SKIP; /readyz will report db unavailable")
	}

	srv := &server{
		cfg:       cfg,
		dbPool:    pool,
		txManager: db.NewTxManager(pool),
	}

	var authHandler *auth.Handler
	var attemptsHandler *attempts.Handler
	var assessmentsHandler *assessments.Handler
	var adminHandler *admin.Handler
	var academicsHandler *academics.Handler
	var gradebookHandler *gradebook.Handler
	var resourcesHandler *resources.Handler
	var academicsSvc academics.Service
	var authIssuer *auth.TokenIssuer
	var sched *scheduler.Scheduler
	if !cfg.DatabaseSkip {
		authIssuer = auth.NewTokenIssuer(cfg.JWTSigningKey, "vts-edu-api", "vts-edu-web", cfg.AccessTokenTTL)
		authRepo := auth.NewRepository(pool.Pool)
		authSvc := auth.NewService(authRepo, srv.txManager, authIssuer, cfg.RefreshTokenTTL)
		authHandler = auth.NewHandler(authSvc)

		attemptsRepo := attempts.NewRepository(pool.Pool)
		attemptsSvc := attempts.NewService(attemptsRepo, srv.txManager)
		attemptsHandler = attempts.NewHandler(attemptsSvc, authIssuer)

		assessmentsRepo := assessments.NewRepository(pool.Pool)
		assessmentsSvc := assessments.NewService(assessmentsRepo, srv.txManager)
		assessmentsHandler = assessments.NewHandler(assessmentsSvc, authIssuer)

		if cfg.SchedulerEnabled {
			sched = scheduler.New(cfg.SchedulerInterval)
			sched.Register(assessments.NewTransitionJob(assessmentsRepo))
			sched.Start()
		}

		adminRepo := admin.NewRepository(pool.Pool)
		adminSvc := admin.NewService(adminRepo, srv.txManager)
		adminHandler = admin.NewHandler(adminSvc, authIssuer)

		academicsRepo := academics.NewRepository(pool.Pool)
		academicsSvc = academics.NewService(academicsRepo, srv.txManager)
		academicsHandler = academics.NewHandler(academicsSvc, authIssuer)

		gradebookRepo := gradebook.NewRepository(pool.Pool)
		gradebookSvc := gradebook.NewService(gradebookRepo, &gradebook.AcademicAccessAdapter{Repo: academicsRepo})
		gradebookHandler = gradebook.NewHandler(gradebookSvc, authIssuer)

		storageProvider, err := buildResourceStorageProvider(cfg)
		if err != nil {
			slog.Error("failed to initialize resource storage", "error", err)
			os.Exit(1)
		}
		resourcesRepo := resources.NewRepository(pool.Pool)
		resourcesSvc := resources.NewService(resourcesRepo, storageProvider, cfg.MaxUploadSize)
		resourcesHandler = resources.NewHandler(resourcesSvc, authIssuer)
	}

	if sched != nil {
		defer sched.Stop()
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(vtsmiddleware.RequestLogger)
	r.Use(ratelimit.Middleware(limiter))
	r.Use(corsMiddleware(cfg.FrontendOrigins))

	// Huma feasibility spike: bounded, academics-only, mounted under a
	// sub-prefix to keep existing routes untouched. The spike will be a
	// no-op if academicsSvc is nil (DB_SKIP mode).
	if academicsSvc != nil && authIssuer != nil {
		_ = academics.MountHumaSpike(r, academics.HumaSpikeDeps{Svc: academicsSvc, Issuer: authIssuer})
	}

	r.Get("/healthz", srv.healthHandler)
	r.Get("/readyz", srv.readyHandler)

	r.Route("/api/v1", func(r chi.Router) {
		// CSRF token endpoint: GET /api/v1/auth/csrf-token
		r.Get("/auth/csrf-token", srv.csrfTokenHandler)

		// Auth endpoints. Implemented when the database is available.
		if authHandler != nil {
			r.Post("/auth/login", authHandler.Login)
			r.Get("/me", authHandler.Me)
			r.Get("/me/teaching/classes", academicsHandler.ListMyTeachingClasses)
			r.Post("/auth/refresh", authHandler.Refresh)
			r.Post("/auth/logout", authHandler.Logout)
			r.Post("/auth/change-password", authHandler.ChangePassword)
		} else {
			r.Post("/auth/login", srv.loginPlaceholderHandler)
			r.Get("/me", srv.mePlaceholderHandler)
			r.Get("/me/teaching/classes", srv.academicsPlaceholderHandler)
			r.Post("/auth/refresh", srv.refreshPlaceholderHandler)
			r.Post("/auth/logout", srv.logoutPlaceholderHandler)
		}

		// Attempt runtime endpoints.
		// Request-time expiration reconciliation happens inside submit/save handlers.
		if attemptsHandler != nil {
			r.Get("/me/assessments", attemptsHandler.ListAssignedAssessments)
			r.Get("/me/attempts", attemptsHandler.ListAttemptHistory)
			r.Post("/assessments/{assessment_id}/attempts", attemptsHandler.StartAttempt)

			r.Get("/attempts/{attempt_id}", attemptsHandler.GetAttempt)
			r.Get("/attempts/{attempt_id}/result", attemptsHandler.GetAttemptResult)
			r.Put("/attempts/{attempt_id}/answers/{attempt_item_id}", attemptsHandler.SaveAnswer)
			r.Post("/attempts/{attempt_id}/submit", attemptsHandler.SubmitAttempt)
		} else {
			r.Get("/me/assessments", srv.meAssessmentsPlaceholderHandler)
			r.Get("/me/attempts", srv.meAttemptsPlaceholderHandler)
			r.Post("/assessments/{assessment_id}/attempts", srv.startAttemptPlaceholderHandler)

			r.Get("/attempts/{attempt_id}", srv.getAttemptPlaceholderHandler)
			r.Get("/attempts/{attempt_id}/result", srv.getAttemptResultPlaceholderHandler)
			r.Put("/attempts/{attempt_id}/answers/{attempt_item_id}", srv.saveAnswerPlaceholderHandler)
			r.Post("/attempts/{attempt_id}/submit", srv.submitPlaceholderHandler)
		}

		// Teacher/admin assessment endpoints.
		if assessmentsHandler != nil {
			r.Get("/assessments", assessmentsHandler.ListAssessments)

			r.Post("/classes/{class_id}/assessments", assessmentsHandler.CreateAssessment)
			r.Get("/classes/{class_id}/assessments", assessmentsHandler.ListAssessmentsByClass)
			r.Get("/assessments/{id}", assessmentsHandler.GetAssessment)
			r.Get("/assessments/{id}/preview", assessmentsHandler.PreviewAssessment)
			r.Patch("/assessments/{id}", assessmentsHandler.UpdateAssessment)
			r.Post("/assessments/{id}/sections", assessmentsHandler.CreateSection)
			r.Post("/assessments/{id}/sections/{section_id}/duplicate", assessmentsHandler.DuplicateSection)
			r.Patch("/assessment-sections/{section_id}", assessmentsHandler.UpdateSection)
			r.Delete("/assessment-sections/{section_id}", assessmentsHandler.DeleteSection)
			r.Post("/assessments/{id}/sections/reorder", assessmentsHandler.ReorderSections)
			r.Post("/assessment-sections/{section_id}/items", assessmentsHandler.CreateItem)
			r.Post("/assessment-sections/{section_id}/items/{item_id}/duplicate", assessmentsHandler.DuplicateItem)
			r.Patch("/assessment-items/{item_id}", assessmentsHandler.UpdateItem)
			r.Delete("/assessment-items/{item_id}", assessmentsHandler.DeleteItem)
			r.Post("/assessment-sections/{section_id}/items/reorder", assessmentsHandler.ReorderItems)
			r.Post("/assessments/{id}/targets", assessmentsHandler.CreateTarget)
			r.Delete("/assessments/{id}/targets/{target_id}", assessmentsHandler.DeleteTarget)
			r.Post("/assessments/{id}/validate", assessmentsHandler.ValidateAssessment)
			r.Post("/assessments/{id}/publish", assessmentsHandler.PublishAssessment)
			r.Get("/assessments/{id}/publications", assessmentsHandler.ListPublications)

			r.Get("/assessments/{id}/attempts", gradebookHandler.ListAssessmentAttempts)
			r.Get("/assessments/{id}/results", gradebookHandler.GetAssessmentResults)
			r.Get("/assessments/{id}/attempts/export", gradebookHandler.ExportAssessmentAttemptsCSV)

			r.Get("/questions", assessmentsHandler.ListQuestions)
		} else {
			r.Get("/assessments", srv.listAssessmentsPlaceholderHandler)

			r.Post("/classes/{class_id}/assessments", srv.listAssessmentsPlaceholderHandler)
			r.Get("/classes/{class_id}/assessments", srv.listAssessmentsPlaceholderHandler)
			r.Get("/assessments/{id}", srv.listAssessmentsPlaceholderHandler)
			r.Patch("/assessments/{id}", srv.listAssessmentsPlaceholderHandler)
			r.Post("/assessments/{id}/sections", srv.listAssessmentsPlaceholderHandler)
			r.Post("/assessment-sections/{section_id}/items", srv.listAssessmentsPlaceholderHandler)
			r.Post("/assessments/{id}/targets", srv.listAssessmentsPlaceholderHandler)
			r.Post("/assessments/{id}/validate", srv.listAssessmentsPlaceholderHandler)
			r.Post("/assessments/{id}/publish", srv.listAssessmentsPlaceholderHandler)
		}

		// Admin endpoints.
		if adminHandler != nil {
			r.Get("/users", adminHandler.ListUsers)
			r.Post("/users", adminHandler.CreateUser)
			r.Post("/users/imports", adminHandler.ImportUsers)
			r.Put("/users/{user_id}/roles", adminHandler.UpdateRoles)
			r.Post("/users/{user_id}/reset-password", adminHandler.ResetPassword)
			r.Get("/organizations/current", adminHandler.GetOrganization)
			r.Patch("/organizations/current", adminHandler.UpdateOrganization)
			r.Get("/audit-logs", adminHandler.ListAuditLogs)
			r.Get("/audit-logs/export", adminHandler.ExportAuditLogs)
		} else {
			r.Get("/users", srv.adminPlaceholderHandler)
			r.Post("/users", srv.adminPlaceholderHandler)
			r.Put("/users/{user_id}/roles", srv.adminPlaceholderHandler)
			r.Post("/users/{user_id}/reset-password", srv.adminPlaceholderHandler)
			r.Get("/organizations/current", srv.adminPlaceholderHandler)
			r.Patch("/organizations/current", srv.adminPlaceholderHandler)
			r.Get("/audit-logs", srv.adminPlaceholderHandler)
			r.Get("/audit-logs/export", srv.adminPlaceholderHandler)
		}

		// Academics endpoints.
		if academicsHandler != nil {
			r.Get("/academic-terms", academicsHandler.ListTerms)
			r.Post("/academic-terms", academicsHandler.CreateTerm)
			r.Patch("/academic-terms/{term_id}", academicsHandler.UpdateTerm)
			r.Delete("/academic-terms/{term_id}", academicsHandler.ArchiveTerm)

			r.Get("/subjects", academicsHandler.ListSubjects)
			r.Post("/subjects", academicsHandler.CreateSubject)
			r.Patch("/subjects/{subject_id}", academicsHandler.UpdateSubject)
			r.Delete("/subjects/{subject_id}", academicsHandler.ArchiveSubject)

			r.Get("/courses", academicsHandler.ListCourses)
			r.Post("/courses", academicsHandler.CreateCourse)
			r.Patch("/courses/{course_id}", academicsHandler.UpdateCourse)
			r.Delete("/courses/{course_id}", academicsHandler.ArchiveCourse)

			r.Get("/classes", academicsHandler.ListClasses)
			r.Post("/classes", academicsHandler.CreateClass)
			r.Patch("/classes/{class_id}", academicsHandler.UpdateClass)
			r.Delete("/classes/{class_id}", academicsHandler.ArchiveClass)
			r.Get("/classes/{class_id}/teachers", academicsHandler.ListClassTeachers)
			r.Post("/classes/{class_id}/teachers", academicsHandler.AddClassTeacher)
			r.Post("/classes/{class_id}/teachers/bulk", academicsHandler.BulkAssignTeachers)
			r.Delete("/classes/{class_id}/teachers/{user_id}", academicsHandler.RemoveClassTeacher)
			r.Get("/classes/{class_id}/enrollments", academicsHandler.ListEnrollments)
			r.Post("/classes/{class_id}/enrollments", academicsHandler.EnrollStudent)
			r.Post("/classes/{class_id}/enrollments/bulk", academicsHandler.BulkEnrollStudents)
			r.Delete("/classes/{class_id}/enrollments/{user_id}", academicsHandler.UnenrollStudent)

			r.Get("/classes/{class_id}/gradebook", gradebookHandler.GetClassGradebook)
			r.Get("/classes/{class_id}/gradebook/export", gradebookHandler.ExportClassGradebookCSV)
		} else {
			r.Get("/academic-terms", srv.academicsPlaceholderHandler)
			r.Post("/academic-terms", srv.academicsPlaceholderHandler)
			r.Delete("/academic-terms/{term_id}", srv.academicsPlaceholderHandler)
			r.Get("/subjects", srv.academicsPlaceholderHandler)
			r.Post("/subjects", srv.academicsPlaceholderHandler)
			r.Delete("/subjects/{subject_id}", srv.academicsPlaceholderHandler)
			r.Get("/courses", srv.academicsPlaceholderHandler)
			r.Post("/courses", srv.academicsPlaceholderHandler)
			r.Delete("/courses/{course_id}", srv.academicsPlaceholderHandler)
			r.Get("/classes", srv.academicsPlaceholderHandler)
			r.Post("/classes", srv.academicsPlaceholderHandler)
			r.Delete("/classes/{class_id}", srv.academicsPlaceholderHandler)
			r.Get("/classes/{class_id}/teachers", srv.academicsPlaceholderHandler)
			r.Post("/classes/{class_id}/teachers", srv.academicsPlaceholderHandler)
			r.Delete("/classes/{class_id}/teachers/{user_id}", srv.academicsPlaceholderHandler)
			r.Get("/classes/{class_id}/enrollments", srv.academicsPlaceholderHandler)
			r.Post("/classes/{class_id}/enrollments", srv.academicsPlaceholderHandler)
			r.Delete("/classes/{class_id}/enrollments/{user_id}", srv.academicsPlaceholderHandler)
		}

		// Resources endpoints (org/class file materials).
		if resourcesHandler != nil {
			r.Get("/resources", resourcesHandler.ListResources)
			r.Post("/resources", resourcesHandler.CreateResource)
			r.Post("/resources/{id}/publish", resourcesHandler.PublishResource)
			r.Delete("/resources/{id}", resourcesHandler.ArchiveResource)
			r.Post("/resources/{id}/files", resourcesHandler.UploadFile)
			r.Get("/resources/{id}/download", resourcesHandler.DownloadFile)
		} else {
			r.Get("/resources", srv.academicsPlaceholderHandler)
			r.Post("/resources", srv.academicsPlaceholderHandler)
			r.Post("/resources/{id}/publish", srv.academicsPlaceholderHandler)
			r.Delete("/resources/{id}", srv.academicsPlaceholderHandler)
		}
	})

	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr, "environment", cfg.Environment, "db_skip", cfg.DatabaseSkip)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := ""
			for _, o := range origins {
				if o == origin {
					allowed = o
					break
				}
			}

			if allowed != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowed)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Request-ID")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) readyHandler(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{"http": "ok"}
	status := http.StatusOK

	if s.cfg.DatabaseSkip {
		checks["db"] = "skipped"
	} else if err := s.dbPool.Ping(r.Context()); err != nil {
		checks["db"] = "unavailable"
		status = http.StatusServiceUnavailable
	} else {
		checks["db"] = "ok"
	}

	writeJSON(w, status, map[string]any{"status": "ready", "checks": checks})
}

func (s *server) csrfTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := csrf.Generate()
	if err != nil {
		http.Error(w, "failed to generate csrf token", http.StatusInternalServerError)
		return
	}
	csrf.SetCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": string(token)})
}

func (s *server) loginPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: validate credentials, create refresh session, set HttpOnly refresh cookie.
	writeJSON(w, http.StatusOK, map[string]string{"message": "login placeholder"})
}

func (s *server) mePlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "me placeholder; database unavailable"})
}

func (s *server) refreshPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
		return
	}
	// TODO: rotate refresh token and return new access JWT.
	writeJSON(w, http.StatusOK, map[string]string{"message": "refresh placeholder"})
}

func (s *server) logoutPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
		return
	}
	// TODO: revoke refresh session, clear refresh cookie.
	writeJSON(w, http.StatusOK, map[string]string{"message": "logout placeholder"})
}

func (s *server) listAssessmentsPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "assessments placeholder; database unavailable"})
}

func (s *server) adminPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "admin placeholder; database unavailable"})
}

func (s *server) academicsPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "academics placeholder; database unavailable"})
}

func (s *server) meAssessmentsPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "me assessments placeholder; database unavailable"})
}

func (s *server) meAttemptsPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "me attempts placeholder; database unavailable"})
}

func (s *server) startAttemptPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "start attempt placeholder; database unavailable"})
}

func (s *server) getAttemptPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attempt_id")
	// TODO: load attempt, check ownership, return snapshot metadata.
	writeJSON(w, http.StatusOK, map[string]any{
		"attempt_id": attemptID,
		"status":     "IN_PROGRESS",
	})
}

func (s *server) getAttemptResultPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	attemptID := chi.URLParam(r, "attempt_id")
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{
		"attempt_id": attemptID,
		"message":    "attempt result placeholder; database unavailable",
	})
}

func (s *server) saveAnswerPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
		return
	}
	attemptID := chi.URLParam(r, "attempt_id")
	itemID := chi.URLParam(r, "attempt_item_id")
	// TODO:
	// 1. Validate attempt ownership and IN_PROGRESS status.
	// 2. Check request-time expiration: if server time > expires_at, reject/expire.
	// 3. Optimistic revision update on attempt_answers.
	writeJSON(w, http.StatusOK, map[string]any{
		"attempt_id":      attemptID,
		"attempt_item_id": itemID,
		"revision":        1,
		"message":         "save answer placeholder",
	})
}

func (s *server) submitPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	if !csrf.Validate(r) {
		http.Error(w, "invalid csrf token", http.StatusForbidden)
		return
	}
	attemptID := chi.URLParam(r, "attempt_id")
	// TODO:
	// 1. Lock attempt, validate status/deadline/ownership.
	// 2. Request-time expiration: if expired, transition to EXPIRED.
	// 3. Otherwise transition to SUBMITTED.
	// 4. Synchronous MCQ/simple grading inside the request transaction.
	// 5. If grading is complex, enqueue River job and return grading_status=QUEUED.
	writeJSON(w, http.StatusOK, map[string]any{
		"attempt_id":     attemptID,
		"status":         "SUBMITTED",
		"grading_status": "FINALIZED",
		"message":        "submit placeholder; MCQ/simple grading is synchronous for demo",
	})
}

func buildResourceStorageProvider(cfg *app.Config) (storage.Provider, error) {
	switch cfg.ResourceStorageType {
	case "local", "":
		return storage.NewLocalProvider(cfg.ResourceLocalPath)
	default:
		return storage.NewLocalProvider(cfg.ResourceLocalPath)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
