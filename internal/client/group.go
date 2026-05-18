package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Group represents a Kanidm group
type Group struct {
	ID          string
	Description string
	Members     []string
	UnixGID     *int64
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, name, description string) (*Group, error) {
	attrs := map[string]any{
		"name": []string{name},
	}

	if description != "" {
		attrs["description"] = []string{description}
	}

	req := NewCreateRequest(attrs)

	resp, err := c.doRequest(ctx, "POST", "/v1/group", req)
	if err != nil {
		return nil, fmt.Errorf("create group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return &Group{
		ID:          name,
		Description: description,
	}, nil
}

// GetGroup retrieves a group by ID
func (c *Client) GetGroup(ctx context.Context, id string) (*Group, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/group/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}

	var entry Entry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	// Normalize member SPNs (e.g. "user@domain") to just the username part,
	// since kanidm returns SPNs but the provider accepts plain names.
	rawMembers := entry.GetStringSlice("member")
	members := make([]string, len(rawMembers))
	for i, m := range rawMembers {
		if before, _, found := strings.Cut(m, "@"); found {
			members[i] = before
		} else {
			members[i] = m
		}
	}

	g := &Group{
		ID:          entry.GetString("name"),
		Description: entry.GetString("description"),
		Members:     members,
	}

	if gidStr := entry.GetString("gidnumber"); gidStr != "" {
		if gid, err := strconv.ParseInt(gidStr, 10, 64); err == nil {
			g.UnixGID = &gid
		}
	}

	return g, nil
}

// UnixExtendGroup adds Unix attributes (gidnumber) to a group.
func (c *Client) UnixExtendGroup(ctx context.Context, id string, gid *int64) error {
	req := make(map[string]any)
	if gid != nil {
		req["gidnumber"] = *gid
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/group/%s/_unix", id), req)
	if err != nil {
		return fmt.Errorf("unix extend group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// UpdateGroup updates a group
func (c *Client) UpdateGroup(ctx context.Context, id, description string, members []string) error {
	attrs := make(map[string]any)

	if description != "" {
		attrs["description"] = []string{description}
	}

	if members != nil {
		attrs["member"] = members
	}

	req := NewUpdateRequest(attrs)

	resp, err := c.doRequest(ctx, "PATCH", "/v1/group/"+id, req)
	if err != nil {
		return fmt.Errorf("update group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/group/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// AddGroupMembers adds members to a group
func (c *Client) AddGroupMembers(ctx context.Context, groupID string, memberIDs []string) error {
	// Use the attribute endpoint to add members
	req := map[string]any{
		"attrs": memberIDs,
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/group/%s/_attr/member", groupID), req)
	if err != nil {
		return fmt.Errorf("add group members: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// RemoveGroupMembers removes members from a group
func (c *Client) RemoveGroupMembers(ctx context.Context, groupID string, memberIDs []string) error {
	// Use the attribute endpoint to remove members
	req := map[string]any{
		"attrs": memberIDs,
	}

	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/group/%s/_attr/member", groupID), req)
	if err != nil {
		return fmt.Errorf("remove group members: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}
