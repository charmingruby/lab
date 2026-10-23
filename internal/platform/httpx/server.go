package httpx

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	http.Server
}

func NewServer(port string) (*Server, chi.Router) {
	addr := ":" + port

	r := chi.NewRouter()
	r.Use(withValidator(newValidator()), withLogging)

	registerProbes(r)

	return &Server{
		Server: http.Server{
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Addr:         addr,
			Handler:      r,
		},
	}, r
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
	r.Get("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func withValidator(v *validator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := validatorToContext(r.Context(), v)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logging.Log.With(
			"path", r.URL.Path,
			"method", r.Method,
		)

		ctx := logging.WithLogger(r.Context(), log)

		log.InfoContext(ctx, "request started")
		defer log.InfoContext(ctx, "request finished")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
