package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPersonResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPersonConfig("acc-test-person", "Acceptance Test Person"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_person.test", "id", "acc-test-person"),
					resource.TestCheckResourceAttr("kanidm_person.test", "displayname", "Acceptance Test Person"),
				),
			},
			{
				Config: testAccPersonConfig("acc-test-person", "Updated Person"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_person.test", "displayname", "Updated Person"),
				),
			},
			{
				ResourceName:      "kanidm_person.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"generate_credential_reset_token",
					"credential_reset_token",
					"credential_reset_token_ttl",
				},
			},
		},
	})
}

func TestAccPersonResource_WithMail(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPersonWithMailConfig("acc-test-mail-person"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_person.test", "id", "acc-test-mail-person"),
					resource.TestCheckResourceAttr("kanidm_person.test", "mail.0", "test@example.com"),
				),
			},
		},
	})
}

func testAccPersonConfig(id, displayName string) string {
	return testProviderConfig() + fmt.Sprintf(`
resource "kanidm_person" "test" {
  id          = %q
  displayname = %q
}
`, id, displayName)
}

func testAccPersonWithMailConfig(id string) string {
	return testProviderConfig() + fmt.Sprintf(`
resource "kanidm_person" "test" {
  id          = %q
  displayname = "Mail Test Person"
  mail        = ["test@example.com"]
}
`, id)
}
