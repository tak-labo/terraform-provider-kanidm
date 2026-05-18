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

func TestCreateOAuth2BasicClient(t *testing.T) {
	callCount := 0
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		callCount++
		switch callCount {
		case 1: // POST /v1/oauth2/_basic
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/oauth2/_basic", r.URL.Path)
			return emptyResponse(http.StatusCreated), nil
		case 2: // GET /v1/oauth2/{name}/_basic_secret
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "_basic_secret")
			return jsonResponse(http.StatusOK, "super-secret"), nil
		default:
			return nil, fmt.Errorf("unexpected call %d", callCount)
		}
	})

	client, err := c.CreateOAuth2BasicClient(context.Background(), "grafana", "Grafana", "https://grafana.example.com")
	require.NoError(t, err)
	assert.Equal(t, "grafana", client.Name)
	assert.Equal(t, "super-secret", client.ClientSecret)
	assert.False(t, client.IsPublic)
}

func TestGetOAuth2Client(t *testing.T) {
	t.Run("basic client detected by attribute presence", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":                     []any{"grafana"},
					"displayname":              []any{"Grafana"},
					"oauth2_rs_origin_landing": []any{"https://grafana.example.com/"},
					"oauth2_rs_basic_secret":   []any{"hidden"},
				},
			}), nil
		})

		oauth2, err := c.GetOAuth2Client(context.Background(), "grafana")
		require.NoError(t, err)
		assert.Equal(t, "grafana", oauth2.Name)
		assert.False(t, oauth2.IsPublic)
		assert.Equal(t, "https://grafana.example.com", oauth2.Origin) // trailing slash stripped
	})

	t.Run("public client", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":                     []any{"mobile-app"},
					"displayname":              []any{"Mobile App"},
					"oauth2_rs_origin_landing": []any{"https://app.example.com/"},
					// no oauth2_rs_basic_secret = public
				},
			}), nil
		})

		oauth2, err := c.GetOAuth2Client(context.Background(), "mobile-app")
		require.NoError(t, err)
		assert.True(t, oauth2.IsPublic)
	})

	t.Run("not found", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return emptyResponse(http.StatusNotFound), nil
		})

		_, err := c.GetOAuth2Client(context.Background(), "nobody")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestGetOAuth2BasicSecret(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/oauth2/grafana/_basic_secret", r.URL.Path)
		return jsonResponse(http.StatusOK, "my-secret"), nil
	})

	secret, err := c.GetOAuth2BasicSecret(context.Background(), "grafana")
	require.NoError(t, err)
	assert.Equal(t, "my-secret", secret)
}

func TestDeleteOAuth2ScopeMap_NotFound(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		return emptyResponse(http.StatusNotFound), nil
	})

	err := c.DeleteOAuth2ScopeMap(context.Background(), "grafana", "admins")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
}
