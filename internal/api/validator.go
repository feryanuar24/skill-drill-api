package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) map[string][]string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errorMessages := make(map[string][]string)
		for _, fe := range ve {
			field := toSnakeCase(fe.Field())
			var msg string
			switch fe.Tag() {
			case "required":
				msg = fmt.Sprintf("%s is required", field)
			case "email":
				msg = fmt.Sprintf("%s must be a valid email address", field)
			case "min":
				msg = fmt.Sprintf("%s must be at least %s", field, fe.Param())
			case "max":
				msg = fmt.Sprintf("%s must not exceed %s", field, fe.Param())
			case "oneof":
				msg = fmt.Sprintf("%s must be one of: %s", field, strings.Join(strings.Fields(fe.Param()), ", "))
			default:
				msg = fmt.Sprintf("%s failed validation on %s", field, fe.Tag())
			}
			errorMessages[field] = append(errorMessages[field], msg)
		}
		return errorMessages
	}

	return map[string][]string{
		"error": {err.Error()},
	}
}

func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
