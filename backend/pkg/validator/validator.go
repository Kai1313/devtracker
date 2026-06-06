package validator

import (
	"reflect"
	"strings"

	apperrors "devtracker/backend/pkg/errors"

	basevalidator "github.com/go-playground/validator/v10"
)

var validate = newValidator()

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
}

func Struct(payload any) error {
	if err := validate.Struct(payload); err != nil {
		if validationErrors, ok := err.(basevalidator.ValidationErrors); ok {
			return apperrors.Validation(format(validationErrors))
		}

		return apperrors.BadRequest(err.Error())
	}

	return nil
}

func newValidator() *basevalidator.Validate {
	instance := basevalidator.New()
	instance.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		if name == "" {
			return field.Name
		}

		return name
	})

	return instance
}

func format(validationErrors basevalidator.ValidationErrors) []FieldError {
	fields := make([]FieldError, 0, len(validationErrors))
	for _, item := range validationErrors {
		fields = append(fields, FieldError{
			Field:   item.Field(),
			Message: messageFor(item),
			Tag:     item.Tag(),
		})
	}

	return fields
}

func messageFor(err basevalidator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "must be a valid email address"
	case "uuid":
		return "must be a valid UUID"
	case "min":
		return "must be at least " + err.Param() + " characters"
	case "max":
		return "must be at most " + err.Param() + " characters"
	default:
		return "is invalid"
	}
}
