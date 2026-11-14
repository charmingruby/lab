package customerr

import (
	"errors"
	"net/http"

	"github.com/charmingruby/lab/internal/platform/httpx"
	"github.com/charmingruby/lab/internal/platform/o11y/logging"
)

type HTTPError struct {
	OriginalError Error
	Message       string
	Status        int
}

func HandleHTTPErr(w http.ResponseWriter, r *http.Request, err error) {
	log := logging.LoggerFromContext(r.Context())

	if cerr, ok := errors.AsType[*Error](err); ok {
		httpErr := mapToHTTP(*cerr)

		log.Warn("application error",
			"type", httpErr.OriginalError.Type,
			"message", httpErr.Message,
			"original_error", httpErr.OriginalError.Error(),
		)

		httpx.WriteErrorResponse(w, httpErr.Status, httpx.ErrorResponse{
			Message: httpErr.Message,
		})

		return
	}

	log.Error("internal server error", "error", err)

	unhandledErr := unhandledHTTPError()

	httpx.WriteErrorResponse(w, unhandledErr.Status, httpx.ErrorResponse{
		Message: unhandledErr.Message,
	})
}

func mapToHTTP(err Error) HTTPError {
	httpErr := HTTPError{
		Message:       err.Message,
		OriginalError: err,
	}

	switch err.Type {
	case TypeNotFound:
		httpErr.Status = http.StatusNotFound
		return httpErr
	case TypeConflict:
		httpErr.Status = http.StatusConflict
		return httpErr
	case TypeValidation:
		httpErr.Status = http.StatusUnprocessableEntity
		return httpErr
	case TypeUnauthorized:
		httpErr.Status = http.StatusUnauthorized
		return httpErr
	case TypeInternal:
		httpErr.Status = http.StatusInternalServerError
		httpErr.Message = "Internal Server Error"
		return httpErr
	}

	return unhandledHTTPError()
}

func unhandledHTTPError() HTTPError {
	return HTTPError{
		Status:  http.StatusInternalServerError,
		Message: "Internal Server Error",
	}
}
