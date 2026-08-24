package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charmingruby/new/internal/shared/customerr"
	"github.com/charmingruby/new/pkg/o11y"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteOKResponse(w http.ResponseWriter, v any) {
	WriteResponse(w, http.StatusOK, v)
}

func WriteCreatedResponse(w http.ResponseWriter, v any) {
	WriteResponse(w, http.StatusCreated, v)
}

func WriteEmptyResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteResponse(w http.ResponseWriter, status int, v any) {
	if v == nil {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"Internal Server Error"}`))
		return
	}

	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	log := o11y.LoggerFromContext(r.Context())

	if customErr, ok := errors.AsType[*customerr.Error](err); ok {
		var originalErrMsg string
		if customErr.Err != nil {
			originalErrMsg = customErr.Err.Error()
		}

		log.Warn("application error",
			"type", customErr.Type,
			"message", customErr.Message,
			"original_error", originalErrMsg,
		)

		writeCustomError(w, customErr)
		return
	}

	log.Error("internal server error", "error", err)

	WriteResponse(w, http.StatusInternalServerError, ErrorResponse{
		Message: "Internal Server Error",
	})
}

func writeCustomError(w http.ResponseWriter, err *customerr.Error) {
	status := mapStatus(err.Type)

	WriteResponse(w, status, ErrorResponse{
		Message: err.Message,
	})
}

func mapStatus(t customerr.ErrorType) int {
	switch t {
	case customerr.TypeNotFound:
		return http.StatusNotFound
	case customerr.TypeValidation:
		return http.StatusBadRequest
	case customerr.TypeConflict:
		return http.StatusConflict
	case customerr.TypeIntegration:
		return http.StatusInternalServerError
	case customerr.TypeUnauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
