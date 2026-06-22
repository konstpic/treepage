package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/blob"
	"github.com/konstpic/treepage/backend/pkg/config"
	"github.com/konstpic/treepage/backend/pkg/database"
	"github.com/konstpic/treepage/backend/pkg/embeddings"
	"github.com/konstpic/treepage/backend/pkg/health"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"github.com/konstpic/treepage/backend/pkg/internalauth"
	"github.com/konstpic/treepage/backend/pkg/logging"
	"github.com/konstpic/treepage/backend/pkg/metrics"
	"github.com/konstpic/treepage/backend/pkg/middleware"
	"github.com/konstpic/treepage/backend/pkg/migrate"
	"github.com/konstpic/treepage/backend/pkg/notify"
	"github.com/konstpic/treepage/backend/server/internal/handler"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/rag"
	"github.com/konstpic/treepage/backend/server/internal/search"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
	"go.uber.org/zap"
)

type AppConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Postgres config.PostgresConfig `yaml:"postgres"`
	JWT      config.JWTConfig      `yaml:"jwt"`
	Logging  config.LoggingConfig  `yaml:"logging"`
	Security config.SecurityConfig `yaml:"security"`
	Search   struct {
		DefaultLimit int `yaml:"default_limit"`
		MaxLimit     int `yaml:"max_limit"`
	} `yaml:"search"`
}

func main() {
	var cfg AppConfig
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	cfg.Postgres.Password = os.Getenv("DB_PASSWORD")

	loader := config.Loader{Path: os.Getenv("CONFIG_PATH")}
	if err := loader.Load(&cfg); err != nil {
		panic(err)
	}
	if cfg.JWT.Secret == "" {
		panic("JWT_SECRET is required")
	}

	logLevel := logging.ResolveLevel(cfg.Logging.Level)
	log, _ := logging.New(logLevel)
	defer log.Sync()

	db, err := database.Connect(cfg.Postgres, logging.GormLogLevel(logLevel))
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
	}

	migrationsDir := migrate.ResolveDir()
	migResult, err := migrate.Run(context.Background(), db, migrationsDir)
	if err != nil {
		log.Fatal("database migrations failed", zap.String("dir", migrationsDir), zap.Error(err))
	}
	if len(migResult.Applied) > 0 {
		log.Info("database migrations applied", zap.Strings("versions", migResult.Applied))
	}
	if len(migResult.Skipped) > 0 {
		log.Info("database migrations skipped (already applied)", zap.Int("count", len(migResult.Skipped)))
	}

	jwtMgr, err := pkgjwt.NewManager(cfg.JWT)
	if err != nil {
		log.Fatal("jwt init failed", zap.Error(err))
	}

	spaces := service.NewSpaceService(db)
	docs := service.NewDocumentService(db)
	repos := service.NewRepositoryService(db)
	audit := service.NewAuditService(db)
	admin := service.NewAdminService(db)
	groups := service.NewGroupService(db)
	prefs := service.NewUserPrefsService(db)
	notifications := service.NewNotificationService(db, notify.NewWebhookFromEnv(), notify.NewSlackFromEnv(), notify.NewEmailFromEnv())
	attachmentsDir := os.Getenv("ATTACHMENTS_DIR")
	if attachmentsDir == "" {
		attachmentsDir = "/data/attachments"
	}
	blobStore, err := blob.NewFromEnv(attachmentsDir)
	if err != nil {
		log.Fatal("attachment storage init failed", zap.Error(err))
	}
	attachments := service.NewAttachmentService(db, blobStore)
	_ = attachments.EnsureDir()
	pageACL := service.NewPageACLService(db)
	comments := service.NewCommentService(db)
	analyticsSvc := service.NewAnalyticsService(db)
	searcher := search.NewFromEnv(db, log)

	llmClient := llm.NewClient(llm.LoadConfigFromEnv(
		os.Getenv("LLM_ENABLED") == "true",
		os.Getenv("LLM_API_URL"),
		os.Getenv("LLM_API_KEY"),
		os.Getenv("LLM_MODEL"),
	))
	embedClient := embeddings.NewClient(embeddings.LoadConfigFromEnv())

	syncURL := os.Getenv("SYNC_SERVICE_URL")
	if syncURL == "" {
		syncURL = "http://backend-sync:8083"
	}
	syncClient := syncclient.New(syncURL)
	ragSvc := rag.New(db, llmClient, embedClient, spaces, pageACL)
	ragWorker := rag.NewWorker(ragSvc, db, log)
	ragWorker.Start(context.Background())

	if osSearcher, ok := searcher.(*search.OpenSearchSearcher); ok {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()
			if err := osSearcher.EnsureIndex(ctx); err != nil {
				log.Warn("opensearch index setup failed", zap.Error(err))
				return
			}
			n, err := osSearcher.BackfillFromDB(ctx)
			if err != nil {
				log.Warn("opensearch backfill failed", zap.Int("documents", n), zap.Error(err))
			} else if n > 0 {
				log.Info("opensearch backfill completed", zap.Int("documents", n))
			}
		}()
	}

	welcomeCfg := service.LoadWelcomeBootstrapConfig()
	if welcomeRepo, created, err := service.BootstrapWelcomeSpace(context.Background(), db, welcomeCfg, log); err != nil {
		log.Warn("welcome space bootstrap failed", zap.Error(err))
	} else if created && welcomeRepo != nil {
		go func(repoID string) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			for attempt := 1; attempt <= 5; attempt++ {
				code, body, err := syncClient.TriggerSync(ctx, repoID)
				if err == nil && code >= 200 && code < 300 {
					log.Info("welcome space initial sync completed", zap.Int("status", code))
					return
				}
				log.Warn("welcome space sync retry",
					zap.Int("attempt", attempt),
					zap.Int("status", code),
					zap.Error(err),
					zap.ByteString("body", body),
				)
				time.Sleep(time.Duration(attempt) * 3 * time.Second)
			}
		}(welcomeRepo.ID)
	}

	metrics.Register()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(middleware.RateLimitFromEnv("server", cfg.Security.RateLimitRPS))
	r.Use(middleware.PrometheusHTTP("server"))
	r.Use(middleware.AccessLogger(logLevel, middleware.ZapAccessLog(log)))

	h := health.NewHandler(func(ctx context.Context) error {
		return database.Ping(ctx, db)
	})
	h.Register(r)

	hdl := handler.New(spaces, docs, repos, audit, prefs, notifications, attachments, pageACL, comments, analyticsSvc, ragSvc, searcher, llmClient, admin, db, jwtMgr, syncClient, cfg.Security.EnableAuditLog)
	hdl.Register(r)
	adminHdl := handler.NewAdminHandler(hdl, admin, groups, spaces, repos, audit, syncClient, ragWorker, cfg.Security.EnableAuditLog)
	adminHdl.Register(r)

	internalToken := internalauth.TokenFromEnv()
	if internalToken != "" {
		internal := r.Group("/api/internal")
		internal.Use(internalauth.Middleware(internalToken))
		hdl.RegisterInternal(internal)
	}

	srv := &http.Server{Addr: cfg.Server.Addr(), Handler: r}
	h.SetReady(true)

	go func() {
		log.Info("backend-server listening", zap.String("addr", cfg.Server.Addr()))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	h.SetReady(false)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
