package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nedpals/supabase-go"
	"github.com/redis/go-redis/v9"

	"github.com/NayanGattani/linkforge/backend/internal/config"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	supabase   *supabase.Client
	redis      *redis.Client
	baseURL    string
}

func New(cfg config.Config, logger *slog.Logger, sb *supabase.Client, redis *redis.Client) *Server {
	s := &Server{
		logger:   logger,
		supabase: sb,
		redis:    redis,
		baseURL:  cfg.BaseURL,
	}

	r := chi.NewRouter()

	r.Use(middleware.Timeout(cfg.ReadTimeout + cfg.WriteTimeout))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.With(
		requireJSONMiddleware,
		bodySizeLimitMiddleware(2048),
		rateLimitMiddleware(s.redis, cfg.RateLimitWriteLimit, cfg.RateLimitWriteBurst, time.Minute),
	).Post("/api/shorten", s.handleShorten())

	r.With(
		rateLimitMiddleware(s.redis, cfg.RateLimitReadLimit, cfg.RateLimitReadBurst, time.Second),
	).Get("/{id}", s.handleRedirect())
	r.Get("/api/links/{id}/stats", s.handleStats())

	s.httpServer = &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")

	shutdownErr := s.httpServer.Shutdown(ctx)
	if s.redis != nil {
		if err := s.redis.Close(); err != nil {
			s.logger.Error("failed to close redis client", "error", err)
		}
	}

	return shutdownErr
}
