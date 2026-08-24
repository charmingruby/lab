package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/charmingruby/new/pkg/validator"
)

var (
	ErrInvalidPayload = errors.New("invalid payload")
	ErrMissingParam   = errors.New("missing param")
)

func ParseRequest[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	var obj T

	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
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

func GetPageQueryParam(r *http.Request) int {
	const defaultPage = 1

	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		return defaultPage
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < defaultPage {
		return defaultPage
	}

	return page
}
