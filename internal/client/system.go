package client

import (
	"context"
	"fmt"
)

// GetSystemAttr retrieves a specific attribute from the Kanidm system configuration.
func (c *Client) GetSystemAttr(ctx context.Context, attrName string) ([]string, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/system/_attr/%s", attrName), nil)
	if err != nil {
		return nil, fmt.Errorf("get system attr %s: %w", attrName, err)
	}

	var vals []string
	if err := decodeResponse(resp, &vals); err != nil {
		return nil, err
	}
	return vals, nil
}

// AddSystemAttr appends values to a system attribute.
func (c *Client) AddSystemAttr(ctx context.Context, attrName string, values []string) error {
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/system/_attr/%s", attrName), values)
	if err != nil {
		return fmt.Errorf("add system attr %s: %w", attrName, err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

// RemoveSystemAttr removes values from a system attribute.
func (c *Client) RemoveSystemAttr(ctx context.Context, attrName string, values []string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/system/_attr/%s", attrName), values)
	if err != nil {
		return fmt.Errorf("remove system attr %s: %w", attrName, err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
