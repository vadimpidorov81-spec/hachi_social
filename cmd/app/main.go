package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	domainusers "github.com/hachisocial/hachisocial/internal/domain/users"
	usersapp "github.com/hachisocial/hachisocial/internal/feature/users/application"
	userspostgres "github.com/hachisocial/hachisocial/internal/feature/users/repository/postgres"
	usershttp "github.com/hachisocial/hachisocial/internal/feature/users/transport/http"
	"github.com/hachisocial/hachisocial/internal/platform/clock"
	"github.com/hachisocial/hachisocial/internal/platform/config"
	"github.com/hachisocial/hachisocial/internal/platform/database"
	"github.com/hachisocial/hachisocial/internal/platform/httpserver"
	"github.com/hachisocial/hachisocial/internal/platform/identity"
	"github.com/hachisocial/hachisocial/internal/platform/logging"
	"github.com/hachisocial/hachisocial/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, err := logging.New(cfg.LogDir, cfg.LogLevel, cfg.Environment, cfg.Version)
	if err != nil {
		return err
	}
	defer func() {
		_ = logger.Close()
	}()
	slog.SetDefault(logger.Logger)

	rootContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.Open(rootContext, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	userRepository := userspostgres.New(pool)
	userService := usersapp.NewService(
		userRepository,
		identity.UUIDv4Generator{},
		clock.UTCClock{},
	)
	var principal usershttp.PrincipalProvider = usershttp.DenyPrincipalProvider{}
	if cfg.DevelopmentUser != "" {
		developmentUserID, err := domainusers.ParseID(cfg.DevelopmentUser)
		if err != nil {
			return err
		}
		principal = usershttp.NewStaticPrincipalProvider(developmentUserID, domainusers.RoleUser)
		logger.Warn(
			"development identity enabled",
			slog.String("user_id", developmentUserID.String()),
		)
	}
	userHandler := usershttp.NewHandler(userService, principal)

	router := chi.NewRouter()
	router.Use(httpserver.RequestIDMiddleware)
	router.Use(httpserver.Recoverer(logger.Logger))
	router.Use(httpserver.AccessLog(logger.Logger))
	router.Use(httpserver.SecurityHeaders)
	router.Get("/live", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	})
	router.Get("/ready", func(response http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(response, "not ready", http.StatusServiceUnavailable)
			return
		}
		response.WriteHeader(http.StatusOK)
	})
	router.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.FS(web.Assets))))
	router.Get("/", func(response http.ResponseWriter, request *http.Request) {
		content, err := web.Assets.ReadFile("index.html")
		if err != nil {
			http.Error(response, "frontend unavailable", http.StatusInternalServerError)
			return
		}
		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = response.Write(content)
	})
	router.Mount("/api/v1/users", userHandler.Routes())
	router.Mount("/api/v1/admin/users", userHandler.AdminRoutes())

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server started", slog.String("address", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	select {
	case <-rootContext.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErrors:
		if err != nil {
			return err
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return err
	}
	logger.Info("application stopped")
	return nil
}
