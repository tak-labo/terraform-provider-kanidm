package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateServiceAccount(t *testing.T) {
	callCount := 0
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		callCount++
		switch callCount {
		case 1: // POST /v1/service_account
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/service_account", r.URL.Path)
			return emptyResponse(http.StatusCreated), nil
		case 2: // POST /v1/service_account/{id}/_api_token
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "_api_token")
			return jsonResponse(http.StatusOK, map[string]any{"token": "eyJ.test.token"}), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", callCount)
		}
	})

	sa, err := c.CreateServiceAccount(context.Background(), "iac-bot", "IaC Bot", []string{"idm_admins"})
	require.NoError(t, err)
	assert.Equal(t, "iac-bot", sa.ID)
	assert.Equal(t, "IaC Bot", sa.DisplayName)
	assert.Equal(t, "eyJ.test.token", sa.APIToken)
}

func TestGetServiceAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":             []any{"iac-bot"},
					"displayname":      []any{"IaC Bot"},
					"entry_managed_by": []any{"idm_admins"},
				},
			}), nil
		})

		sa, err := c.GetServiceAccount(context.Background(), "iac-bot")
		require.NoError(t, err)
		assert.Equal(t, "iac-bot", sa.ID)
		assert.Equal(t, "IaC Bot", sa.DisplayName)
		assert.Equal(t, []string{"idm_admins"}, sa.EntryManagedBy)
		assert.Empty(t, sa.APIToken) // not returned in GET
	})

	t.Run("not found", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return emptyResponse(http.StatusNotFound), nil
		})

		_, err := c.GetServiceAccount(context.Background(), "nobody")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestGenerateServiceAccountToken(t *testing.T) {
	t.Run("JSON token", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "_api_token")
			return jsonResponse(http.StatusOK, map[string]any{"token": "eyJ.abc.def"}), nil
		})

		token, err := c.GenerateServiceAccountToken(context.Background(), "iac-bot", "iac-managed", nil)
		require.NoError(t, err)
		assert.Equal(t, "eyJ.abc.def", token)
	})

	t.Run("plain JWT string", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       stringBody(`"eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJib2IifQ.sig"`),
				Header:     make(http.Header),
			}, nil
		})

		token, err := c.GenerateServiceAccountToken(context.Background(), "iac-bot", "iac-managed", nil)
		require.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJib2IifQ.sig", token)
	})
}
