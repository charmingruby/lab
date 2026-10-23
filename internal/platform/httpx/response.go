package httpx

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charmingruby/lab/internal/platform/logging"
	"github.com/charmingruby/lab/internal/shared/customerr"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteOKResponse(w http.ResponseWriter, v any) {
	writeResponse(w, http.StatusOK, v)
}

func WriteCreatedResponse(w http.ResponseWriter, v any) {
	writeResponse(w, http.StatusCreated, v)
}

func WriteEmptyResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func WriteBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	HandleErr(w, r, customerr.Validation(err.Error()))
}

func HandleErr(w http.ResponseWriter, r *http.Request, err error) {
	log := logging.LoggerFromContext(r.Context())

	if cerr, ok := errors.AsType[*customerr.Error](err); ok {
		httpErr := customerr.MapToHTTP(*cerr)

		log.Warn("application error",
			"type", httpErr.OriginalError.Type,
			"message", httpErr.Message,
			"original_error", httpErr.OriginalError.Error(),
		)

		writeResponse(w, httpErr.Status, ErrorResponse{
			Message: httpErr.Message,
		})

		return
	}

	log.Error("internal server error", "error", err)

	writeResponse(w, customerr.UnhandledHTTPError.Status, ErrorResponse{
		Message: customerr.UnhandledHTTPError.Message,
	})
}

func writeResponse(w http.ResponseWriter, status int, v any) {
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
