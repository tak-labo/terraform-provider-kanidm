package client

import (
	"context"
	"fmt"
	"strings"
)

// OAuth2Client represents a Kanidm OAuth2 resource server
type OAuth2Client struct {
	Name                           string
	DisplayName                    string
	Origin                         string
	RedirectURIs                   []string
	ScopeMaps                      map[string][]string
	ClientID                       string // Computed
	ClientSecret                   string // Only for basic/confidential clients, populated on creation
	IsPublic                       bool
	AllowInsecureClientDisablePKCE bool
}

// CreateOAuth2BasicClient creates a new OAuth2 basic (confidential) client
func (c *Client) CreateOAuth2BasicClient(ctx context.Context, name, displayName, origin string) (*OAuth2Client, error) {
	req := NewCreateRequest(map[string]any{
		"name":                     []string{name},
		"displayname":              []string{displayName},
		"oauth2_rs_origin_landing": []string{origin},
	})

	resp, err := c.doRequest(ctx, "POST", "/v1/oauth2/_basic", req)
	if err != nil {
		return nil, fmt.Errorf("create oauth2 basic client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// The create response doesn't include the client secret
	// We need to retrieve it using the show secret endpoint
	clientSecret, err := c.GetOAuth2BasicSecret(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("retrieve client secret: %w", err)
	}

	return &OAuth2Client{
		Name:         name,
		DisplayName:  displayName,
		Origin:       origin,
		ClientID:     name, // Client ID is typically the name
		ClientSecret: clientSecret,
		IsPublic:     false,
	}, nil
}

// CreateOAuth2PublicClient creates a new OAuth2 public client
func (c *Client) CreateOAuth2PublicClient(ctx context.Context, name, displayName, origin string) (*OAuth2Client, error) {
	req := NewCreateRequest(map[string]any{
		"name":                     []string{name},
		"displayname":              []string{displayName},
		"oauth2_rs_origin_landing": []string{origin},
	})

	resp, err := c.doRequest(ctx, "POST", "/v1/oauth2/_public", req)
	if err != nil {
		return nil, fmt.Errorf("create oauth2 public client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return &OAuth2Client{
		Name:        name,
		DisplayName: displayName,
		Origin:      origin,
		ClientID:    name,
		IsPublic:    true,
	}, nil
}

// GetOAuth2Client retrieves an OAuth2 client by name
func (c *Client) GetOAuth2Client(ctx context.Context, name string) (*OAuth2Client, error) {
	resp, err := c.doRequest(ctx, "GET", "/v1/oauth2/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("get oauth2 client: %w", err)
	}

	var entry Entry
	if err := decodeResponse(resp, &entry); err != nil {
		return nil, err
	}

	// Determine if public based on oauth2_rs_basic_secret attribute presence
	// Note: The value is hidden for basic clients, so we check if the key exists in attrs
	_, hasBasicSecret := entry.Attrs["oauth2_rs_basic_secret"]
	isPublic := !hasBasicSecret

	// Use 'name' attribute for the name (not oauth2_rs_name which is for internal use)
	clientName := entry.GetString("name")
	if clientName == "" {
		clientName = entry.GetString("oauth2_rs_name")
	}

	// Read portal landing URL (single-value) as "origin"
	// oauth2_rs_origin_landing = portal link URL shown in Kanidm application portal
	origin := entry.GetString("oauth2_rs_origin_landing")
	if len(origin) > 0 && origin[len(origin)-1] == '/' {
		origin = origin[:len(origin)-1]
	}

	allowInsecureDisablePKCE := false
	if vals := entry.GetStringSlice("oauth2_allow_insecure_client_disable_pkce"); len(vals) > 0 && vals[0] == "true" {
		allowInsecureDisablePKCE = true
	}

	// Read allowed redirect URIs from oauth2_rs_origin (multi-value).
	// Strip trailing slash only for root URLs (e.g. https://example.com/) to stay
	// consistent with the write path, which adds a trailing slash to root URLs.
	// Path URIs (e.g. https://example.com/callback/) are preserved as-is.
	rawRedirectURIs := entry.GetStringSlice("oauth2_rs_origin")
	redirectURIs := make([]string, len(rawRedirectURIs))
	for i, u := range rawRedirectURIs {
		if len(u) > 0 && u[len(u)-1] == '/' && strings.Count(u, "/") == 3 {
			u = u[:len(u)-1]
		}
		redirectURIs[i] = u
	}

	return &OAuth2Client{
		Name:                           clientName,
		DisplayName:                    entry.GetString("displayname"),
		Origin:                         origin,
		RedirectURIs:                   redirectURIs,
		ClientID:                       clientName,
		IsPublic:                       isPublic,
		AllowInsecureClientDisablePKCE: allowInsecureDisablePKCE,
		// Note: Client secret is never returned in GET responses
	}, nil
}

// UpdateOAuth2Client updates an OAuth2 client
func (c *Client) UpdateOAuth2Client(ctx context.Context, name string, displayName, origin string, redirectURIs []string, allowInsecureDisablePKCE *bool) error {
	attrs := make(map[string]any)

	if displayName != "" {
		attrs["displayname"] = []string{displayName}
	}

	if origin != "" {
		// oauth2_rs_origin_landing = portal landing URL (single-value)
		landingURL := origin
		if strings.Count(landingURL, "/") == 2 {
			landingURL = landingURL + "/"
		}
		attrs["oauth2_rs_origin_landing"] = []string{landingURL}
	}

	if redirectURIs != nil {
		// oauth2_rs_origin = allowed redirect URIs (multi-value)
		// Root URLs like https://example.com must be sent as https://example.com/
		normalized := make([]string, len(redirectURIs))
		for i, u := range redirectURIs {
			if strings.Count(u, "/") == 2 {
				u = u + "/"
			}
			normalized[i] = u
		}
		attrs["oauth2_rs_origin"] = normalized
	}

	if allowInsecureDisablePKCE != nil {
		if *allowInsecureDisablePKCE {
			attrs["oauth2_allow_insecure_client_disable_pkce"] = []string{"true"}
		} else {
			attrs["oauth2_allow_insecure_client_disable_pkce"] = []string{"false"}
		}
	}

	req := NewUpdateRequest(attrs)

	resp, err := c.doRequest(ctx, "PATCH", "/v1/oauth2/"+name, req)
	if err != nil {
		return fmt.Errorf("update oauth2 client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeleteOAuth2Client deletes an OAuth2 client
func (c *Client) DeleteOAuth2Client(ctx context.Context, name string) error {
	resp, err := c.doRequest(ctx, "DELETE", "/v1/oauth2/"+name, nil)
	if err != nil {
		return fmt.Errorf("delete oauth2 client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// SetOAuth2ScopeMap sets the scope mapping for an OAuth2 client
func (c *Client) SetOAuth2ScopeMap(ctx context.Context, rsName, groupName string, scopes []string) error {
	// Send scopes array directly (not wrapped in an object)
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/oauth2/%s/_scopemap/%s", rsName, groupName), scopes)
	if err != nil {
		return fmt.Errorf("set oauth2 scope map: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// DeleteOAuth2ScopeMap removes a scope mapping for an OAuth2 client
func (c *Client) DeleteOAuth2ScopeMap(ctx context.Context, rsName, groupName string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/oauth2/%s/_scopemap/%s", rsName, groupName), nil)
	if err != nil {
		return fmt.Errorf("delete oauth2 scope map: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return nil
}

// GetOAuth2BasicSecret retrieves the client secret for a basic OAuth2 client
func (c *Client) GetOAuth2BasicSecret(ctx context.Context, name string) (string, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/oauth2/%s/_basic_secret", name), nil)
	if err != nil {
		return "", fmt.Errorf("get oauth2 basic secret: %w", err)
	}

	// The API returns the secret as a plain JSON string
	var secret string
	if err := decodeResponse(resp, &secret); err != nil {
		return "", err
	}

	return secret, nil
}

// RegenerateOAuth2BasicSecret regenerates the client secret for a basic OAuth2 client
// This invalidates the old secret and generates a new one
func (c *Client) RegenerateOAuth2BasicSecret(ctx context.Context, name string) (string, error) {
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/oauth2/%s/_basic_secret", name), nil)
	if err != nil {
		return "", fmt.Errorf("regenerate oauth2 basic secret: %w", err)
	}

	// The API returns the new secret as a plain JSON string
	var secret string
	if err := decodeResponse(resp, &secret); err != nil {
		return "", err
	}

	return secret, nil
}
