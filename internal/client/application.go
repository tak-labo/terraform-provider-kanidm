package client

import (
	"context"
	"fmt"
)

// Application represents a Kanidm application entry.
type Application struct {
	ID          string
	Name        string
	DisplayName string
	LinkedGroup string
}

type scimReference struct {
	UUID  *string `json:"uuid,omitempty"`
	Value *string `json:"value,omitempty"`
}

type scimReferenceResponse struct {
	UUID  string `json:"uuid"`
	Value string `json:"value"`
}

type scimApplicationCreateReq struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"displayname"`
	LinkedGroup scimReference `json:"linked_group"`
}

type scimApplicationEntry struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	DisplayName string                  `json:"displayname"`
	LinkedGroup []scimReferenceResponse `json:"linked_group"`
}

type scimEntryPutReq struct {
	ID          string           `json:"id"`
	DisplayName *string          `json:"displayname,omitempty"`
	LinkedGroup *[]scimReference `json:"linked_group,omitempty"`
}

// CreateApplication creates a new application entry linked to a Kanidm group.
func (c *Client) CreateApplication(ctx context.Context, name, displayName, linkedGroupID string) (*Application, error) {
	body := scimApplicationCreateReq{
		Name:        name,
		DisplayName: displayName,
		LinkedGroup: scimReference{UUID: &linkedGroupID},
	}

	resp, err := c.doRequest(ctx, "POST", "/v1/application", body)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	app, err := c.GetApplication(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("read application after create: %w", err)
	}
	return app, nil
}

// GetApplication retrieves an application entry by name.
func (c *Client) GetApplication(ctx context.Context, name string) (*Application, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/application/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}

	var entry scimApplicationEntry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	linkedGroup := ""
	if len(entry.LinkedGroup) > 0 {
		linkedGroup = entry.LinkedGroup[0].UUID
	}

	return &Application{
		ID:          entry.ID,
		Name:        entry.Name,
		DisplayName: entry.DisplayName,
		LinkedGroup: linkedGroup,
	}, nil
}

// UpdateApplication updates an application entry.
func (c *Client) UpdateApplication(ctx context.Context, id, displayName, linkedGroupID string) (*Application, error) {
	body := scimEntryPutReq{
		ID:          id,
		DisplayName: &displayName,
		LinkedGroup: &[]scimReference{{UUID: &linkedGroupID}},
	}

	resp, err := c.doRequest(ctx, "PUT", "/v1/application/"+id, body)
	if err != nil {
		return nil, fmt.Errorf("update application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	app, err := c.GetApplication(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("read application after update: %w", err)
	}
	return app, nil
}

// DeleteApplication deletes an application entry.
func (c *Client) DeleteApplication(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/application/"+id, nil)
	if err != nil {
		return fmt.Errorf("delete application: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
