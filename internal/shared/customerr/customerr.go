package customerr

import (
	"errors"
)

type ErrorType string

const (
	TypeNotFound     = "NOT_FOUND"
	TypeInternal     = "INTERNAL"
	TypeValidation   = "VALIDATION"
	TypeConflict     = "CONFLICT"
	TypeIntegration  = "INTEGRATION"
	TypeUnauthorized = "UNAUTHORIZED"
)

type Error struct {
	Details any       `json:"_"`
	Err     error     `json:"-"`
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Code    string    `json:"code"`
}

func New(t ErrorType, message string, err error) *Error {
	return &Error{
		Type:    t,
		Message: message,
		Err:     err,
	}
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NotFound(message string) *Error {
	return New(TypeNotFound, message, nil)
}

func IsNotFound(err error) bool {
	return IsType(err, TypeNotFound)
}

func Integration(err error) *Error {
	return New(TypeIntegration, err.Error(), nil)
}

func IsInvalidOperation(err error) bool {
	return IsType(err, TypeIntegration)
}

func Conflict(message string) *Error {
	return New(TypeConflict, message, nil)
}

func IsConflict(err error) bool {
	return IsType(err, TypeConflict)
}

func Validation(message string) *Error {
	return New(TypeValidation, message, nil)
}

func IsValidation(err error) bool {
	return IsType(err, TypeValidation)
}

func Internal(message string, err error) *Error {
	return New(TypeInternal, message, err)
}

func IsInternal(err error) bool {
	return IsType(err, TypeInternal)
}

func Unauthorized(message string, err error) *Error {
	return New(TypeUnauthorized, message, err)
}

func IsUnauthorized(err error) bool {
	return IsType(err, TypeUnauthorized)
}

func (e *Error) WithDetails(details any) *Error {
	e.Details = details
	return e
}

func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

func IsType(err error, t ErrorType) bool {
	var e *Error
	return errors.As(err, &e) && e.Type == t
}
