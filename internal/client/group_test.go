package client

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateGroup(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/group", r.URL.Path)
		return emptyResponse(http.StatusCreated), nil
	})

	g, err := c.CreateGroup(context.Background(), "developers", "Dev team")
	require.NoError(t, err)
	assert.Equal(t, "developers", g.ID)
	assert.Equal(t, "Dev team", g.Description)
}

func TestGetGroup(t *testing.T) {
	t.Run("success with SPN normalization", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":        []any{"developers"},
					"description": []any{"Dev team"},
					"member":      []any{"alice@idm.example.com", "bob@idm.example.com"},
				},
			}), nil
		})

		g, err := c.GetGroup(context.Background(), "developers")
		require.NoError(t, err)
		assert.Equal(t, "developers", g.ID)
		// SPN should be normalized to plain usernames
		assert.Equal(t, []string{"alice", "bob"}, g.Members)
	})

	t.Run("member without domain is preserved", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, map[string]any{
				"attrs": map[string]any{
					"name":   []any{"ops"},
					"member": []any{"charlie"},
				},
			}), nil
		})

		g, err := c.GetGroup(context.Background(), "ops")
		require.NoError(t, err)
		assert.Equal(t, []string{"charlie"}, g.Members)
	})

	t.Run("not found", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return emptyResponse(http.StatusNotFound), nil
		})

		_, err := c.GetGroup(context.Background(), "nobody")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestUpdateGroup(t *testing.T) {
	t.Run("with members", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPatch, r.Method)
			return emptyResponse(http.StatusOK), nil
		})
		assert.NoError(t, c.UpdateGroup(context.Background(), "developers", "Updated", []string{"alice"}))
	})

	t.Run("nil members skipped", func(t *testing.T) {
		c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
			return emptyResponse(http.StatusOK), nil
		})
		assert.NoError(t, c.UpdateGroup(context.Background(), "developers", "Updated", nil))
	})
}

func TestDeleteGroup(t *testing.T) {
	c := newTestClient(t, func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodDelete, r.Method)
		return emptyResponse(http.StatusNoContent), nil
	})
	assert.NoError(t, c.DeleteGroup(context.Background(), "developers"))
}
