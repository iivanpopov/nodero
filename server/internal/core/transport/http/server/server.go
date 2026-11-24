package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"server/internal/core/transport/http/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type Config struct {
	Port            string
	ShutdownTimeout time.Duration
	Title           string
	Version         string
}

type Server struct {
	httpServer      *http.Server
	mux             *http.ServeMux
	api             huma.API
	log             *slog.Logger
	shutdownTimeout time.Duration
}

func New(cfg Config, log *slog.Logger, middlewares ...middleware.Middleware) *Server {
	mux := http.NewServeMux()

	humaCfg := huma.DefaultConfig(cfg.Title, cfg.Version)
	humaCfg.DocsRenderer = huma.DocsRendererScalar
	humaCfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		middleware.BearerAuthScheme: {
			Type:        "http",
			Scheme:      "bearer",
			Description: "Session token passed in the Authorization header as 'Bearer <token>'",
		},
	}
	api := humago.New(mux, humaCfg)

	handler := middleware.Chain(mux, middlewares...)

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
		mux:             mux,
		api:             api,
		log:             log,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

func (s *Server) APIVersionGroup(prefix string) huma.API {
	return huma.NewGroup(s.api, prefix)
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.log.Info("http server listening", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server listen: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutting down http server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	return <-errCh
}
