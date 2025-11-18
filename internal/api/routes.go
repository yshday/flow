package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/yourusername/issue-tracker/docs"
	"github.com/yourusername/issue-tracker/internal/api/handlers"
	"github.com/yourusername/issue-tracker/internal/api/middleware"
	"github.com/yourusername/issue-tracker/internal/auth"
	"github.com/yourusername/issue-tracker/internal/repository"
	"github.com/yourusername/issue-tracker/internal/service"
	"github.com/yourusername/issue-tracker/pkg/cache"
	"github.com/yourusername/issue-tracker/pkg/email"
	"github.com/yourusername/issue-tracker/pkg/markdown"
	"github.com/yourusername/issue-tracker/pkg/ratelimit"
	"github.com/yourusername/issue-tracker/pkg/storage"
)

// Config holds API configuration
type Config struct {
	DB                   *sql.DB
	Cache                cache.Cache
	JWTSecret            string
	JWTRefreshSecret     string
	JWTAccessTTL         time.Duration
	JWTRefreshTTL        time.Duration
	RateLimitEnabled     bool
	RateLimitPerMinute   int
	RateLimitWindow      time.Duration
	StoragePath          string
	StorageMaxFileSize   int64
	SMTPHost             string
	SMTPPort             string
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
}

// NewRouter creates a new HTTP router with all routes
func NewRouter(config Config) http.Handler {
	// Initialize local storage
	localStorage, err := storage.NewLocalStorage(config.StoragePath)
	if err != nil {
		log.Fatal("Failed to initialize local storage:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(config.DB)
	projectRepo := repository.NewProjectRepository(config.DB)
	boardRepo := repository.NewBoardRepository(config.DB)
	issueRepo := repository.NewIssueRepository(config.DB)
	commentRepo := repository.NewCommentRepository(config.DB)
	labelRepo := repository.NewLabelRepository(config.DB)
	memberRepo := repository.NewProjectMemberRepository(config.DB)
	activityRepo := repository.NewActivityRepository(config.DB)
	milestoneRepo := repository.NewMilestoneRepository(config.DB)
	notificationRepo := repository.NewNotificationRepository(config.DB)
	statisticsRepo := repository.NewStatisticsRepository(config.DB)
	searchRepo := repository.NewSearchRepository(config.DB)
	attachmentRepo := repository.NewAttachmentRepository(config.DB)
	reactionRepo := repository.NewReactionRepository(config.DB)
	mentionRepo := repository.NewMentionRepository(config.DB)
	referenceRepo := repository.NewIssueReferenceRepository(config.DB)
	watcherRepo := repository.NewIssueWatcherRepository(config.DB)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		config.JWTSecret,
		config.JWTRefreshSecret,
		config.JWTAccessTTL,
		config.JWTRefreshTTL,
	)

	// Initialize email client
	emailClient := email.NewClient(email.Config{
		Host:     config.SMTPHost,
		Port:     config.SMTPPort,
		Username: config.SMTPUsername,
		Password: config.SMTPPassword,
		From:     config.SMTPFrom,
	})

	// Initialize markdown renderer
	markdownRenderer := markdown.NewRenderer()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)
	authorizationService := service.NewAuthorizationService(projectRepo, memberRepo)
	projectService := service.NewProjectService(projectRepo, boardRepo, config.DB, config.Cache)
	mentionService := service.NewMentionService(mentionRepo, userRepo, notificationRepo, config.DB)
	referenceService := service.NewIssueReferenceService(referenceRepo, issueRepo, config.DB)
	issueService := service.NewIssueService(issueRepo, watcherRepo, authorizationService, config.DB, config.Cache, markdownRenderer, mentionService, referenceService)
	commentService := service.NewCommentService(commentRepo, issueRepo, authorizationService, config.DB, markdownRenderer, mentionService, referenceService)
	labelService := service.NewLabelService(labelRepo, projectRepo, issueRepo, authorizationService, config.DB, config.Cache)
	boardService := service.NewBoardService(boardRepo, projectRepo, authorizationService, config.DB)
	memberService := service.NewProjectMemberService(memberRepo, projectRepo, userRepo, config.DB)
	activityService := service.NewActivityService(activityRepo, projectRepo, issueRepo, config.DB)
	milestoneService := service.NewMilestoneService(milestoneRepo, projectRepo, authorizationService, config.Cache)
	notificationService := service.NewNotificationService(notificationRepo, userRepo, emailClient)
	statisticsService := service.NewStatisticsService(statisticsRepo, projectRepo, memberRepo, config.Cache)
	searchService := service.NewSearchService(searchRepo, projectRepo, memberRepo, config.Cache)
	attachmentService := service.NewAttachmentService(attachmentRepo, issueRepo, authorizationService, localStorage, config.StorageMaxFileSize)
	reactionService := service.NewReactionService(reactionRepo, issueRepo, commentRepo, authorizationService, config.DB)
	watcherService := service.NewWatcherService(watcherRepo, issueRepo, notificationRepo, authorizationService, config.DB)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	projectHandler := handlers.NewProjectHandler(projectService)
	issueHandler := handlers.NewIssueHandler(issueService)
	commentHandler := handlers.NewCommentHandler(commentService)
	labelHandler := handlers.NewLabelHandler(labelService)
	boardHandler := handlers.NewBoardHandler(boardService)
	memberHandler := handlers.NewProjectMemberHandler(memberService)
	activityHandler := handlers.NewActivityHandler(activityService)
	milestoneHandler := handlers.NewMilestoneHandler(milestoneService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	statisticsHandler := handlers.NewStatisticsHandler(statisticsService)
	searchHandler := handlers.NewSearchHandler(searchService)
	attachmentHandler := handlers.NewAttachmentHandler(attachmentService)
	reactionHandler := handlers.NewReactionHandler(reactionService)
	watcherHandler := handlers.NewWatcherHandler(watcherService)

	// Create router
	mux := http.NewServeMux()

	// Public routes (no authentication required)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)

	// Protected routes (authentication required)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/v1/auth/me", authHandler.GetMe)

	// Project routes
	protectedMux.HandleFunc("POST /api/v1/projects", projectHandler.Create)
	protectedMux.HandleFunc("GET /api/v1/projects", projectHandler.List)
	protectedMux.HandleFunc("GET /api/v1/projects/{id}", projectHandler.GetByID)
	protectedMux.HandleFunc("PUT /api/v1/projects/{id}", projectHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/projects/{id}", projectHandler.Delete)

	// Issue routes
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/issues", issueHandler.Create)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/issues", issueHandler.List)
	protectedMux.HandleFunc("GET /api/v1/issues/{id}", issueHandler.GetByID)
	protectedMux.HandleFunc("GET /api/v1/issues/{projectKey}/{issueNumber}", issueHandler.GetByProjectKey)
	protectedMux.HandleFunc("PUT /api/v1/issues/{id}", issueHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/issues/{id}", issueHandler.Delete)
	protectedMux.HandleFunc("PUT /api/v1/issues/{id}/move", issueHandler.MoveToColumn)

	// Comment routes
	protectedMux.HandleFunc("POST /api/v1/issues/{issueId}/comments", commentHandler.Create)
	protectedMux.HandleFunc("GET /api/v1/issues/{issueId}/comments", commentHandler.List)
	protectedMux.HandleFunc("PUT /api/v1/comments/{id}", commentHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/comments/{id}", commentHandler.Delete)

	// Label routes
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/labels", labelHandler.Create)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/labels", labelHandler.List)
	protectedMux.HandleFunc("GET /api/v1/labels/{id}", labelHandler.GetByID)
	protectedMux.HandleFunc("PUT /api/v1/labels/{id}", labelHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/labels/{id}", labelHandler.Delete)

	// Issue-Label relationship routes
	protectedMux.HandleFunc("POST /api/v1/issues/{issueId}/labels/{labelId}", labelHandler.AddToIssue)
	protectedMux.HandleFunc("DELETE /api/v1/issues/{issueId}/labels/{labelId}", labelHandler.RemoveFromIssue)
	protectedMux.HandleFunc("GET /api/v1/issues/{issueId}/labels", labelHandler.ListByIssueID)

	// Board routes
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/board", boardHandler.List)
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/board/columns", boardHandler.Create)
	protectedMux.HandleFunc("PUT /api/v1/board/columns/{id}", boardHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/board/columns/{id}", boardHandler.Delete)

	// Project Member routes
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/members", memberHandler.ListMembers)
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/members", memberHandler.AddMember)
	protectedMux.HandleFunc("PUT /api/v1/projects/{projectId}/members/{userId}", memberHandler.UpdateMemberRole)
	protectedMux.HandleFunc("DELETE /api/v1/projects/{projectId}/members/{userId}", memberHandler.RemoveMember)

	// Activity routes
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/activities", activityHandler.ListByProject)
	protectedMux.HandleFunc("GET /api/v1/issues/{issueId}/activities", activityHandler.ListByIssue)

	// Milestone routes
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/milestones", milestoneHandler.Create)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/milestones", milestoneHandler.ListByProject)
	protectedMux.HandleFunc("GET /api/v1/milestones/{id}", milestoneHandler.GetByID)
	protectedMux.HandleFunc("PUT /api/v1/milestones/{id}", milestoneHandler.Update)
	protectedMux.HandleFunc("DELETE /api/v1/milestones/{id}", milestoneHandler.Delete)

	// Notification routes
	protectedMux.HandleFunc("GET /api/v1/notifications", notificationHandler.List)
	protectedMux.HandleFunc("GET /api/v1/notifications/unread/count", notificationHandler.GetUnreadCount)
	protectedMux.HandleFunc("GET /api/v1/notifications/{id}", notificationHandler.GetByID)
	protectedMux.HandleFunc("PUT /api/v1/notifications/read", notificationHandler.MarkAsRead)
	protectedMux.HandleFunc("PUT /api/v1/notifications/read/all", notificationHandler.MarkAllAsRead)
	protectedMux.HandleFunc("DELETE /api/v1/notifications/{id}", notificationHandler.Delete)

	// Statistics routes
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/statistics", statisticsHandler.GetProjectStatistics)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/statistics/issues", statisticsHandler.GetIssueStatistics)
	protectedMux.HandleFunc("GET /api/v1/users/me/statistics", statisticsHandler.GetUserActivityStatistics)

	// Search routes
	protectedMux.HandleFunc("GET /api/v1/search", searchHandler.Search)
	protectedMux.HandleFunc("GET /api/v1/search/issues", searchHandler.SearchIssues)
	protectedMux.HandleFunc("GET /api/v1/search/projects", searchHandler.SearchProjects)

	// Attachment routes
	protectedMux.HandleFunc("POST /api/v1/issues/{id}/attachments", attachmentHandler.Upload)
	protectedMux.HandleFunc("GET /api/v1/issues/{id}/attachments", attachmentHandler.List)
	protectedMux.HandleFunc("GET /api/v1/attachments/{id}/download", attachmentHandler.Download)
	protectedMux.HandleFunc("DELETE /api/v1/attachments/{id}", attachmentHandler.Delete)

	// Reaction routes
	protectedMux.HandleFunc("POST /api/v1/reactions/{entity_type}/{entity_id}", reactionHandler.AddReaction)
	protectedMux.HandleFunc("GET /api/v1/reactions/{entity_type}/{entity_id}", reactionHandler.GetReactions)
	protectedMux.HandleFunc("GET /api/v1/reactions/{entity_type}/{entity_id}/summary", reactionHandler.GetReactionSummary)
	protectedMux.HandleFunc("DELETE /api/v1/reactions/{entity_type}/{entity_id}/{emoji}", reactionHandler.RemoveReaction)

	// Watcher routes
	protectedMux.HandleFunc("POST /api/v1/projects/{projectId}/issues/{issueNumber}/watch", watcherHandler.WatchIssue)
	protectedMux.HandleFunc("DELETE /api/v1/projects/{projectId}/issues/{issueNumber}/watch", watcherHandler.UnwatchIssue)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/issues/{issueNumber}/watching", watcherHandler.CheckWatchingStatus)
	protectedMux.HandleFunc("GET /api/v1/projects/{projectId}/issues/{issueNumber}/watchers", watcherHandler.GetWatchers)
	protectedMux.HandleFunc("GET /api/v1/user/watching", watcherHandler.GetWatchedIssues)

	// Apply authentication middleware to protected routes
	mux.Handle("/api/v1/auth/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/projects", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/projects/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/issues", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/issues/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/comments", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/comments/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/labels", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/labels/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/board", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/board/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/milestones", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/milestones/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/notifications", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/notifications/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/users", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/users/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/search", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/search/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/attachments", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/attachments/", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/reactions", middleware.Authenticate(authService)(protectedMux))
	mux.Handle("/api/v1/reactions/", middleware.Authenticate(authService)(protectedMux))

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger documentation
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Apply global middlewares
	middlewares := []middleware.Middleware{
		middleware.Logging(),
		middleware.DefaultCORS(),
	}

	// Add rate limiting if enabled
	if config.RateLimitEnabled {
		limiter := ratelimit.NewLimiter(config.Cache, config.RateLimitPerMinute, config.RateLimitWindow)
		middlewares = append(middlewares, middleware.RateLimit(middleware.RateLimitConfig{
			Limiter: limiter,
		}))
		log.Printf("Rate limiting enabled: %d requests per %v", config.RateLimitPerMinute, config.RateLimitWindow)
	}

	handler := middleware.Chain(mux, middlewares...)

	return handler
}
