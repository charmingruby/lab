package httpx

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/lab/pkg/validator"
)

type Server struct {
	http.Server
}

func NewServer(port string, validator *validator.Validator) (*Server, chi.Router) {
	addr := ":" + port

	r := chi.NewRouter()
	r.Use(withValidator(validator), withO11y)

	var apiRouter chi.Router
	r.Route("/api", func(router chi.Router) {
		apiRouter = router
	})

	registerProbes(apiRouter)

	return &Server{
		Server: http.Server{
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Addr:         addr,
			Handler:      r,
		},
	}, apiRouter
}

func (s *Server) Start() error {
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Close(ctx context.Context) error {
	return s.Shutdown(ctx)
}

func registerProbes(r chi.Router) {
	r.Get("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
