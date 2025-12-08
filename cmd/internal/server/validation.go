package server

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors converts validator errors into readable format
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errors = append(errors, ValidationError{
				Field:   getJSONFieldName(e),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

// getJSONFieldName extracts the JSON field name from the validation error
func getJSONFieldName(e validator.FieldError) string {
	// Convert StructField to json field name
	// e.g., "Messages[0].Content" -> "messages[0].content"
	field := e.Namespace()

	// Remove the struct name prefix (e.g., "Completion.")
	parts := strings.SplitN(field, ".", 2)
	if len(parts) == 2 {
		field = parts[1]
	}

	// Convert to lowercase for JSON convention
	return strings.ToLower(field)
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		if e.Type().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, e.Param())
	case "max":
		if e.Type().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// HandleValidationError is a helper to send validation errors as JSON response
func HandleValidationError(c *gin.Context, err error) {
	errors := FormatValidationErrors(err)
	c.JSON(400, gin.H{
		"error":   "Validation failed",
		"details": errors,
	})
}
