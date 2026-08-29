package api

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) gin.H {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		errorMessages := make([]string, 0, len(ve))
		for _, fe := range ve {
			field := toSnakeCase(fe.Field())
			switch fe.Tag() {
			case "required":
				errorMessages = append(errorMessages, fmt.Sprintf("%s is required", field))
			case "email":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address", field))
			case "min":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s", field, fe.Param()))
			case "max":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must not exceed %s", field, fe.Param()))
			case "oneof":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be one of: %s", field, strings.Join(strings.Fields(fe.Param()), ", ")))
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation on %s", field, fe.Tag()))
			}
		}
		return gin.H{
			"error":   strings.Join(errorMessages, "; "),
			"details": errorMessages,
		}
	}

	return gin.H{"error": err.Error()}
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
