package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/auth/internal/devmode"
	"github.com/konstpic/treepage/backend/auth/internal/handler"
	"github.com/konstpic/treepage/backend/auth/internal/service"
	"github.com/konstpic/treepage/backend/pkg/config"
	"github.com/konstpic/treepage/backend/pkg/database"
	"github.com/konstpic/treepage/backend/pkg/health"
	pkgjwt "github.com/konstpic/treepage/backend/pkg/jwt"
	"github.com/konstpic/treepage/backend/pkg/logging"
	"github.com/konstpic/treepage/backend/pkg/metrics"
	"github.com/konstpic/treepage/backend/pkg/middleware"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type AppConfig struct {
	Server   config.ServerConfig   `yaml:"server"`
	Postgres config.PostgresConfig `yaml:"postgres"`
	OIDC     config.OIDCConfig     `yaml:"oidc"`
	JWT      config.JWTConfig      `yaml:"jwt"`
	Logging  config.LoggingConfig  `yaml:"logging"`
	Security config.SecurityConfig `yaml:"security"`
	Frontend struct {
		URL string `yaml:"url"`
	} `yaml:"frontend"`
}

func main() {
	var cfg AppConfig
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	cfg.Postgres.Password = os.Getenv("DB_PASSWORD")
	cfg.OIDC.ClientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	cfg.Security.CSRFSecret = os.Getenv("CSRF_SECRET")

	loader := config.Loader{Path: os.Getenv("CONFIG_PATH")}
	if err := loader.Load(&cfg); err != nil {
		panic(err)
	}
	if cfg.JWT.Secret == "" {
		panic("JWT_SECRET is required")
	}

	logLevel := logging.ResolveLevel(cfg.Logging.Level)
	log, err := logging.New(logLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	db, err := database.Connect(cfg.Postgres, logging.GormLogLevel(logLevel))
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
	}

	jwtMgr, err := pkgjwt.NewManager(cfg.JWT)
	if err != nil {
		log.Fatal("jwt init failed", zap.Error(err))
	}

	authSvc := service.NewAuthService(db, jwtMgr, log)
	authSvc.SetOIDCClaimDefaults(service.OIDCClaimSettings{
		RoleClaim:  cfg.OIDC.RoleClaim,
		GroupClaim: cfg.OIDC.GroupClaim,
		SyncGroups: cfg.OIDC.SyncGroups,
	})

	ctx := context.Background()
	if err := authSvc.BootstrapLocalAdmin(ctx); err != nil {
		log.Fatal("bootstrap local admin failed", zap.Error(err))
	}
	log.Info("local admin bootstrap enabled; password login available for users with a local password")

	var oidcProvider *oidc.Provider
	var oauth2Cfg *oauth2.Config
	if cfg.OIDC.Enabled && cfg.OIDC.IssuerURL != "" {
		ctx := context.Background()
		var err error
		oidcProvider, err = oidc.NewProvider(ctx, cfg.OIDC.IssuerURL)
		if err != nil {
			log.Warn("oidc provider init failed, login disabled", zap.Error(err))
		} else {
			oauth2Cfg = &oauth2.Config{
				ClientID:     cfg.OIDC.ClientID,
				ClientSecret: cfg.OIDC.ClientSecret,
				RedirectURL:  cfg.OIDC.RedirectURL,
				Endpoint:     oidcProvider.Endpoint(),
				Scopes:       splitScopes(cfg.OIDC.Scopes),
			}
		}
	}

	metrics.Register()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.CORS(cfg.Security.AllowedOrigins))
	r.Use(middleware.RateLimitFromEnv("auth", cfg.Security.RateLimitRPS))
	r.Use(middleware.PrometheusHTTP("auth"))
	r.Use(middleware.AccessLogger(logLevel, middleware.ZapAccessLog(log)))

	h := health.NewHandler(func(ctx context.Context) error {
		return database.Ping(ctx, db)
	})
	h.Register(r)

	hdl := handler.New(authSvc, jwtMgr, handler.NewStateStoreFromEnv(handler.NewMemoryStateStore()), oidcProvider, oauth2Cfg, cfg.Frontend.URL, os.Getenv("AUTHENTIK_PUBLIC_URL"), devmode.LocalLoginEnabled(), log)
	hdl.Register(r)

	srv := &http.Server{Addr: cfg.Server.Addr(), Handler: r}
	h.SetReady(true)

	go func() {
		log.Info("backend-auth listening", zap.String("addr", cfg.Server.Addr()))
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

func splitScopes(s string) []string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return []string{"openid", "profile", "email"}
	}
	return parts
}
