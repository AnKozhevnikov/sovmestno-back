package apperror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Errors []AppError `json:"errors"`
}

// One returns an ErrorResponse with a single error.
func One(code, message string) ErrorResponse {
	return ErrorResponse{Errors: []AppError{{Code: code, Message: message}}}
}

// FromValidation parses go-playground/validator errors into ErrorResponse.
// Returns (response, true) if err contains ValidationErrors, otherwise (_, false).
func FromValidation(err error) (ErrorResponse, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return ErrorResponse{}, false
	}
	var errs []AppError
	for _, fe := range ve {
		errs = append(errs, mapFieldError(fe))
	}
	return ErrorResponse{Errors: errs}, true
}

func mapFieldError(fe validator.FieldError) AppError {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return AppError{Code: "FIELD_REQUIRED", Message: fmt.Sprintf("Field '%s' is required", field)}
	case "email":
		return AppError{Code: "INVALID_EMAIL", Message: "Invalid email format"}
	case "min":
		if field == "password" {
			return AppError{Code: "PASSWORD_TOO_SHORT", Message: fmt.Sprintf("Password must be at least %s characters", fe.Param())}
		}
		return AppError{Code: "VALUE_TOO_SHORT", Message: fmt.Sprintf("Field '%s' must be at least %s characters", field, fe.Param())}
	case "max":
		if field == "password" {
			return AppError{Code: "PASSWORD_TOO_LONG", Message: fmt.Sprintf("Password must be at most %s characters", fe.Param())}
		}
		return AppError{Code: "VALUE_TOO_LONG", Message: fmt.Sprintf("Field '%s' must be at most %s characters", field, fe.Param())}
	case "url":
		return AppError{Code: "INVALID_URL", Message: fmt.Sprintf("Field '%s' must be a valid URL (e.g. https://example.com)", field)}
	case "e164":
		return AppError{Code: "INVALID_PHONE", Message: "Phone must be in E.164 format (e.g. +79001234567)"}
	case "oneof":
		return AppError{Code: "INVALID_VALUE", Message: fmt.Sprintf("Field '%s' must be one of: %s", field, fe.Param())}
	default:
		return AppError{Code: "VALIDATION_ERROR", Message: fmt.Sprintf("Invalid value for field '%s'", field)}
	}
}
