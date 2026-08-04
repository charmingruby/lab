package httpx

import (
	"net/http"

	"github.com/charmingruby/new/pkg/o11y"
	"github.com/charmingruby/new/pkg/validator"
)

func withValidator(v *validator.Validator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := validator.WithValidator(r.Context(), v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withO11y(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := o11y.Log.With(
			"path", r.URL.Path,
			"method", r.Method,
		)

		ctx := o11y.WithLogger(r.Context(), log)

		log.InfoContext(ctx, "request started")
		defer log.InfoContext(ctx, "request finished")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
