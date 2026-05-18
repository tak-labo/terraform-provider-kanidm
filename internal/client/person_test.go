package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePerson(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/person", r.URL.Path)
		return emptyResponse(http.StatusCreated), nil
	})

	p, err := c.CreatePerson(context.Background(), "alice", "Alice Smith")
	require.NoError(t, err)
	assert.Equal(t, "alice", p.ID)
	assert.Equal(t, "Alice Smith", p.DisplayName)
}

func TestGetPerson(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/person/alice", r.URL.Path)
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":        []any{"alice"},
					"displayname": []any{"Alice Smith"},
					"mail":        []any{"alice@example.com"},
				},
			}), nil
		})

		p, err := c.GetPerson(context.Background(), "alice")
		require.NoError(t, err)
		assert.Equal(t, "alice", p.ID)
		assert.Equal(t, "Alice Smith", p.DisplayName)
		assert.Equal(t, []string{"alice@example.com"}, p.Mail)
	})

	t.Run("not found", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return emptyResponse(http.StatusNotFound), nil
		})

		_, err := c.GetPerson(context.Background(), "nobody")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestUpdatePerson(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/v1/person/alice", r.URL.Path)
		return emptyResponse(http.StatusOK), nil
	})

	err := c.UpdatePerson(context.Background(), "alice", "Alice Updated", []string{"new@example.com"}, nil)
	assert.NoError(t, err)
}

func TestDeletePerson(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodDelete, r.Method)
			return emptyResponse(http.StatusNoContent), nil
		})
		assert.NoError(t, c.DeletePerson(context.Background(), "alice"))
	})
}

func TestCreatePersonCredentialResetToken(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "_credential/_update_intent")
		return jsonResponse(http.StatusOK, map[string]any{"token": "reset-token-abc"}), nil
	})

	token, err := c.CreatePersonCredentialResetToken(context.Background(), "alice", nil)
	require.NoError(t, err)
	assert.Equal(t, "reset-token-abc", token)
}
