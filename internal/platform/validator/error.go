package validator

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return "validation failed"
}
