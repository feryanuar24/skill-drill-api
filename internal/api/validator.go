package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
		_ = v.RegisterValidation("password", validatePassword)
	}
}

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
			case "password":
				msg = fmt.Sprintf("%s must be at least 8 characters long and contain a combination of uppercase letters, lowercase letters, numbers, and symbols", field)
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

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	var (
		hasUpper  bool
		hasLower  bool
		hasNumber bool
		hasSymbol bool
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSymbol
}
