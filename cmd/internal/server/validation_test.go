package server

import (
	"bytes"
	"llm-router/types"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMessageValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		jsonBody    string
		expectError bool
	}{
		{
			name:        "valid message",
			jsonBody:    `{"role": "user", "content": "Hello, world!"}`,
			expectError: false,
		},
		{
			name:        "invalid role",
			jsonBody:    `{"role": "invalid", "content": "test"}`,
			expectError: true,
		},
		{
			name:        "missing role",
			jsonBody:    `{"content": "test"}`,
			expectError: true,
		},
		{
			name:        "missing content",
			jsonBody:    `{"role": "user"}`,
			expectError: true,
		},
		{
			name:        "empty content",
			jsonBody:    `{"role": "user", "content": ""}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			var message types.Message
			err := c.ShouldBindJSON(&message)

			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestHandleValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	if w.Body.Len() == 0 {
		t.Error("Expected non-empty response body")
	}
}