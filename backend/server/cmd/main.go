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
	"github.com/konstpic/treepage/backend/pkg/config"
	"github.com/konstpic/treepage/backend/pkg/database"
	"github.com/konstpic/treepage/backend/pkg/health"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"github.com/konstpic/treepage/backend/pkg/logging"
	"github.com/konstpic/treepage/backend/pkg/middleware"
	"github.com/konstpic/treepage/backend/server/internal/handler"
	"github.com/konstpic/treepage/backend/server/internal/llm"
	"github.com/konstpic/treepage/backend/server/internal/search"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
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

	log, _ := logging.New(cfg.Logging.Level)
	defer log.Sync()

	db, err := database.Connect(cfg.Postgres, logger.Info)
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
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
	searcher := search.NewPostgresSearcher(db)

	llmClient := llm.NewClient(llm.LoadConfigFromEnv(
		os.Getenv("LLM_ENABLED") == "true",
		os.Getenv("LLM_API_URL"),
		os.Getenv("LLM_API_KEY"),
		os.Getenv("LLM_MODEL"),
	))

	syncURL := os.Getenv("SYNC_SERVICE_URL")
	if syncURL == "" {
		syncURL = "http://backend-sync:8083"
	}
	syncClient := syncclient.New(syncURL)

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

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS system_settings (
		key VARCHAR(128) PRIMARY KEY,
		value JSONB NOT NULL DEFAULT '{}',
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_by UUID REFERENCES users(id) ON DELETE SET NULL
	)`).Error; err != nil {
		log.Warn("system_settings bootstrap", zap.Error(err))
	}
	_ = db.Exec(`INSERT INTO system_settings (key, value) VALUES
		('platform', '{"search_default_limit":20,"search_max_limit":100,"cache_enabled":false,"logging_level":"info","auto_translate_docs":true}'),
		('auth', '{"oidc_enabled":true,"local_auth_fallback":false}'),
		('git', '{"access_token_ref":"GIT_ACCESS_TOKEN","webhook_secret_ref":"GIT_WEBHOOK_SECRET","default_sync_interval_seconds":300,"default_sync_mode":"scheduled"}'),
		('ui_theme', '"fox_white"'),
		('ui_language', '"en"')
	ON CONFLICT (key) DO NOTHING`)

	_ = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
		slug VARCHAR(256) NOT NULL,
		title VARCHAR(512) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		root_path VARCHAR(512) NOT NULL,
		audience VARCHAR(64) NOT NULL DEFAULT 'developer',
		focus TEXT NOT NULL DEFAULT '',
		status VARCHAR(32) NOT NULL DEFAULT 'draft',
		source_hash VARCHAR(64),
		outline_json JSONB NOT NULL DEFAULT '[]',
		content_markdown TEXT NOT NULL DEFAULT '',
		error_message TEXT NOT NULL DEFAULT '',
		enhanced BOOLEAN NOT NULL DEFAULT false,
		created_by UUID REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		generated_at TIMESTAMPTZ,
		UNIQUE (space_id, slug)
	)`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_books_space ON books(space_id, updated_at DESC)`)

	_ = db.Exec(`CREATE TABLE IF NOT EXISTS document_translations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
		locale VARCHAR(8) NOT NULL,
		source_hash VARCHAR(64) NOT NULL,
		title VARCHAR(512) NOT NULL DEFAULT '',
		content TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (document_id, locale)
	)`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_document_translations_doc ON document_translations(document_id, locale)`)
	_ = db.Exec(`CREATE TABLE IF NOT EXISTS book_translations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
		locale VARCHAR(8) NOT NULL,
		source_hash VARCHAR(64) NOT NULL,
		title VARCHAR(512) NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		content TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (book_id, locale)
	)`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_book_translations_book ON book_translations(book_id, locale)`)
	_ = db.Exec(`UPDATE system_settings SET value = value || '{"auto_translate_docs": true}'::jsonb
		WHERE key = 'platform' AND NOT (value ? 'auto_translate_docs')`)

	_ = db.Exec(`CREATE TABLE IF NOT EXISTS space_groups (
		space_id UUID NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
		group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
		role_id UUID NOT NULL REFERENCES roles(id),
		PRIMARY KEY (space_id, group_id)
	)`).Error
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_space_groups_group ON space_groups(group_id)`).Error

	_ = db.Exec(`ALTER TABLE oidc_providers ADD COLUMN IF NOT EXISTS sync_groups BOOLEAN NOT NULL DEFAULT false`).Error

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(middleware.RateLimit(cfg.Security.RateLimitRPS))

	h := health.NewHandler(func(ctx context.Context) error {
		return database.Ping(ctx, db)
	})
	h.Register(r)

	hdl := handler.New(spaces, docs, repos, audit, searcher, llmClient, admin, db, jwtMgr, syncClient, cfg.Security.EnableAuditLog)
	hdl.Register(r)
	adminHdl := handler.NewAdminHandler(hdl, admin, groups, spaces, repos, audit, syncClient, cfg.Security.EnableAuditLog)
	adminHdl.Register(r)

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
