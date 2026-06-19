package main

import (
	"context"
	"crypto/subtle"
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
	"github.com/konstpic/treepage/backend/pkg/internalauth"
	"github.com/konstpic/treepage/backend/pkg/logging"
	"github.com/konstpic/treepage/backend/pkg/middleware"
	"github.com/konstpic/treepage/backend/sync/internal/syncer"
	"go.uber.org/zap"
)

type AppConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Postgres config.PostgresConfig `yaml:"postgres"`
	Git      config.GitConfig      `yaml:"git"`
	Logging  config.LoggingConfig  `yaml:"logging"`
	Security config.SecurityConfig `yaml:"security"`
}

func main() {
	var cfg AppConfig
	cfg.Postgres.Password = os.Getenv("DB_PASSWORD")

	loader := config.Loader{Path: os.Getenv("CONFIG_PATH")}
	if err := loader.Load(&cfg); err != nil {
		panic(err)
	}

	logLevel := logging.ResolveLevel(cfg.Logging.Level)
	log, _ := logging.New(logLevel)
	defer log.Sync()

	db, err := database.Connect(cfg.Postgres, logging.GormLogLevel(logLevel))
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
	}

	workDir := cfg.Git.WorkDir
	if workDir == "" {
		workDir = "/tmp/treepage-repos"
	}
	syncSvc := syncer.New(db, workDir, os.Getenv("GIT_ACCESS_TOKEN"), log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := cfg.Git.SyncInterval
	if interval == 0 {
		interval = 5 * time.Minute
	}
	go syncSvc.RunScheduled(ctx, interval)

	internalToken := internalauth.TokenFromEnv()
	if internalToken == "" {
		log.Warn("INTERNAL_SERVICE_TOKEN is not set — sync API endpoints are disabled")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.RateLimit(cfg.Security.RateLimitRPS))
	r.Use(middleware.AccessLogger(logLevel, middleware.ZapAccessLog(log)))

	h := health.NewHandler(func(c context.Context) error {
		return database.Ping(c, db)
	})
	h.Register(r)

	internal := r.Group("/api/sync")
	internal.Use(internalauth.Middleware(internalToken))

	internal.POST("/repositories/:id/publish", func(c *gin.Context) {
		var input syncer.PublishInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := syncSvc.PublishDocument(c.Request.Context(), c.Param("id"), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	internal.POST("/repositories/:id", func(c *gin.Context) {
		id := c.Param("id")
		result, err := syncSvc.SyncRepository(c.Request.Context(), id, "manual")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":            "completed",
			"files_processed":   result.FilesProcessed,
			"conflicts_skipped": result.ConflictsSkipped,
		})
	})

	r.POST("/api/sync/webhook/:id", func(c *gin.Context) {
		secret := os.Getenv("GIT_WEBHOOK_SECRET")
		got := c.GetHeader("X-Webhook-Secret")
		if secret == "" || subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid webhook secret"})
			return
		}
		id := c.Param("id")
		go func() {
			_, _ = syncSvc.SyncRepository(context.Background(), id, "webhook")
		}()
		c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
	})

	srv := &http.Server{Addr: cfg.Server.Addr(), Handler: r}
	h.SetReady(true)

	go func() {
		log.Info("backend-sync listening", zap.String("addr", cfg.Server.Addr()))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	h.SetReady(false)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	cancel()
	_ = srv.Shutdown(shutdownCtx)
}
