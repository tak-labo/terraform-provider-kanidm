package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderSchema(t *testing.T) {
	p := New("test")()

	factory := providerserver.NewProtocol6WithError(p)

	server, err := factory()
	require.NoError(t, err)

	resp, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Diagnostics)

	// url and token attributes must be present
	attrNames := make([]string, 0, len(resp.Provider.Block.Attributes))
	for _, a := range resp.Provider.Block.Attributes {
		attrNames = append(attrNames, a.Name)
	}
	assert.Contains(t, attrNames, "url")
	assert.Contains(t, attrNames, "token")
}

func TestProviderResources(t *testing.T) {
	p := New("test")()
	resources := p.Resources(context.Background())
	assert.Len(t, resources, 4, "expected 4 resources: person, service_account, group, oauth2_basic")
}

func TestProviderDataSources(t *testing.T) {
	p := New("test")()
	dataSources := p.DataSources(context.Background())
	assert.Len(t, dataSources, 4, "expected 4 data sources: person, group, service_account, oauth2_basic")
}
