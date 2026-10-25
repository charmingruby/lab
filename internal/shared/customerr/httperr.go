package customerr

import "net/http"

type HTTPError struct {
	OriginalError Error
	Message       string
	Status        int
}

var UnhandledHTTPError = HTTPError{
	Status:  http.StatusInternalServerError,
	Message: "Internal Server Error",
}

func MapToHTTP(err Error) HTTPError {
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
	case TypeInternal, TypeIntegration:
		httpErr.Status = http.StatusInternalServerError
		httpErr.Message = "Internal Server Error"
		return httpErr
	}

	return UnhandledHTTPError
}
