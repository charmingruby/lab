package httpx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	val "github.com/go-playground/validator/v10"
)

type validator struct {
	validator *val.Validate
}

type ctxKey struct{}

func newValidator() *validator {
	return &validator{
		validator: val.New(),
	}
}

func (v *validator) Validate(s any) error {
	err := v.validator.Struct(s)

	if err == nil {
		return nil
	}

	errs := v.unwrapValidationErr(err)

	return errors.New(strings.Join(errs, ", "))
}

func validatorToContext(ctx context.Context, val *validator) context.Context {
	return context.WithValue(ctx, ctxKey{}, val)
}

func validatorFromContext(ctx context.Context) *validator {
	val, ok := ctx.Value(ctxKey{}).(*validator)
	if !ok {
		return newValidator()
	}

	return val
}

func (v *validator) unwrapValidationErr(err error) []string {
	var valErr *val.ValidationErrors
	if !errors.As(err, &valErr) {
		return []string{err.Error()}
	}

	errs := *valErr

	reasonsWrapper := make([]string, 0, len(errs))

	for _, vErr := range errs {
		reason := fmt.Sprintf("field `%s` does not satisfy %s rule", vErr.Field(), vErr.Tag())

		reasonsWrapper = append(reasonsWrapper, reason)
	}

	return reasonsWrapper
}
