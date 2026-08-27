package httpx

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/lab/internal/shared/core"
	"github.com/charmingruby/lab/pkg/validator"
)

var (
	ErrInvalidPayload = errors.New("invalid payload")
	ErrMissingParam   = errors.New("missing param")
)

func ParseRequest[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	var obj T

	if err := json.UnmarshalRead(r.Body, &obj); err != nil {
		WriteResponse(w, http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("%s: %s", ErrInvalidPayload.Error(), err.Error()),
		})

		return nil, err
	}

	val := validator.ValidatorFromContext(r.Context())
	if err := val.Validate(obj); err != nil {
		WriteResponse(w, http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("%s: %s", ErrInvalidPayload.Error(), err.Error()),
		})

		return nil, err
	}

	return &obj, nil
}

func GetPathParam(r *http.Request, key string) (string, error) {
	param := chi.URLParam(r, key)
	if param == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingParam, key)
	}

	return param, nil
}

func GetQueryParam(r *http.Request, key string) (string, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingParam, key)
	}

	return param, nil
}

func GetPaginationParams(r *http.Request) core.PaginationParams {
	params := core.PaginationParams{
		Page:  1,
		Limit: core.DefaultPageSize,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	return params.Validate()
}
