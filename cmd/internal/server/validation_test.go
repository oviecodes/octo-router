package server

import (
	"bytes"
	"llm-router/types"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func TestFormatValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("formats validation errors correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"role": "invalid", "content": ""}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var message types.Message
		err := c.ShouldBindJSON(&message)

		if err == nil {
			t.Fatal("Expected validation errors")
		}

		errors := FormatValidationErrors(err)

		if len(errors) == 0 {
			t.Fatal("Expected formatted errors")
		}
	})

	t.Run("returns empty for non-validator error", func(t *testing.T) {
		errors := FormatValidationErrors(nil)

		if len(errors) != 0 {
			t.Errorf("Expected 0 errors, got %d", len(errors))
		}
	})
}

func TestGetJSONFieldName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"role": "", "content": "test"}`
	c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var message types.Message
	err := c.ShouldBindJSON(&message)

	if err == nil {
		t.Fatal("Expected validation error")
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatal("Expected validator.ValidationErrors")
	}

	for _, e := range validationErrs {
		fieldName := getJSONFieldName(e)
		if fieldName == "" {
			t.Error("Expected non-empty field name")
		}
	}
}

func TestGetErrorMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		jsonBody string
	}{
		{"required field", `{"role": "", "content": "test"}`},
		{"oneof validation", `{"role": "invalid", "content": "test"}`},
		{"temperature range", `{"messages": [{"role": "user", "content": "test"}], "temperature": 3.0}`},
		{"max_tokens range", `{"messages": [{"role": "user", "content": "test"}], "max_tokens": 0}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var obj any
			if tt.name == "required field" || tt.name == "oneof validation" {
				var message types.Message
				obj = &message
			} else {
				var completion types.Completion
				obj = &completion
			}

			err := c.ShouldBindJSON(obj)

			if err == nil {
				t.Fatal("Expected validation error")
			}

			validationErrs, ok := err.(validator.ValidationErrors)
			if !ok {
				t.Skip("Not a validation error")
			}

			for _, e := range validationErrs {
				message := getErrorMessage(e)
				if message == "" {
					t.Error("Expected non-empty error message")
				}
			}
		})
	}
}

func TestHandleValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 400 with error details", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := `{"role": "invalid", "content": ""}`
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var message types.Message
		err := c.ShouldBindJSON(&message)

		if err == nil {
			t.Fatal("Expected validation error")
		}

		HandleValidationError(c, err)

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}

		body2 := w.Body.String()
		if body2 == "" {
			t.Error("Expected non-empty response body")
		}
	})
}
