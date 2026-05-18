package client

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntry_GetString(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]any
		key      string
		expected string
	}{
		{"array value", map[string]any{"name": []any{"alice"}}, "name", "alice"},
		{"string value", map[string]any{"name": "alice"}, "name", "alice"},
		{"empty array", map[string]any{"name": []any{}}, "name", ""},
		{"missing key", map[string]any{}, "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{Attrs: tt.attrs}
			assert.Equal(t, tt.expected, e.GetString(tt.key))
		})
	}
}

func TestEntry_GetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]any
		key      string
		expected []string
	}{
		{"array of any", map[string]any{"mail": []any{"a@x.com", "b@x.com"}}, "mail", []string{"a@x.com", "b@x.com"}},
		{"string slice", map[string]any{"mail": []string{"a@x.com"}}, "mail", []string{"a@x.com"}},
		{"single string", map[string]any{"mail": "a@x.com"}, "mail", []string{"a@x.com"}},
		{"missing key", map[string]any{}, "mail", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Entry{Attrs: tt.attrs}
			assert.Equal(t, tt.expected, e.GetStringSlice(tt.key))
		})
	}
}

func TestCheckResponse_Errors(t *testing.T) {
	c := NewClient("https://idm.example.com", "token")

	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{"404 → ErrNotFound", http.StatusNotFound, ErrNotFound},
		{"401 → ErrUnauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"403 → ErrForbidden", http.StatusForbidden, ErrForbidden},
		{"200 → no error", http.StatusOK, nil},
		{"201 → no error", http.StatusCreated, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := emptyResponse(tt.statusCode)
			err := c.checkResponse(resp)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
