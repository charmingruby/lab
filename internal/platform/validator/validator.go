package validator

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	val "github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/en"
)

var instance = sync.OnceValues(newValidator)

func Validate(s any) error {
	v, translator := instance()

	err := v.Struct(s)

	if err == nil {
		return nil
	}

	errs, ok := errors.AsType[val.ValidationErrors](err)
	if !ok {
		return err
	}

	root := reflect.TypeOf(s)

	for root.Kind() == reflect.Pointer {
		root = root.Elem()
	}

	prefix := root.Name() + "."

	fields := make([]FieldError, 0, len(errs))

	for _, field := range errs {
		fields = append(fields, FieldError{
			Field:   strings.TrimPrefix(field.Namespace(), prefix),
			Code:    field.Tag(),
			Message: field.Translate(translator),
		})
	}

	return &ValidationError{
		Fields: fields,
	}
}

func newValidator() (*val.Validate, ut.Translator) {
	v := val.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}

		return name
	})

	translator := ut.New(en.New()).GetFallback()
	_ = translations.RegisterDefaultTranslations(v, translator)

	return v, translator
}
