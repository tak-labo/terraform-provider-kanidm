package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccPreCheck verifies required environment variables are set.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 to run acceptance tests")
	}
	if os.Getenv("KANIDM_URL") == "" {
		t.Fatal("KANIDM_URL must be set for acceptance tests")
	}
	if os.Getenv("KANIDM_TOKEN") == "" {
		t.Fatal("KANIDM_TOKEN must be set for acceptance tests")
	}
}

// testAccProviderFactories returns provider factories for acceptance tests.
// TLS verification is skipped to support local Docker instances with self-signed certs.
var testAccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kanidm": providerserver.NewProtocol6WithError(New("test")()),
}

// testProviderConfig returns the provider config block with insecure TLS for acceptance tests.
func testProviderConfig() string {
	return `
provider "kanidm" {
  insecure = true
}
`
}
