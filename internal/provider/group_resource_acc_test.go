package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGroupResource(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig("acc-test-group", "Acceptance Test Group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_group.test", "id", "acc-test-group"),
					resource.TestCheckResourceAttr("kanidm_group.test", "description", "Acceptance Test Group"),
				),
			},
			{
				Config: testAccGroupConfig("acc-test-group", "Updated Group"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_group.test", "description", "Updated Group"),
				),
			},
			{
				ResourceName:      "kanidm_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGroupResource_WithMembers(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupWithMembersConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kanidm_group.test", "id", "acc-test-group-members"),
					resource.TestCheckTypeSetElemAttr("kanidm_group.test", "members.*", "acc-test-member"),
				),
			},
		},
	})
}

func testAccGroupConfig(id, description string) string {
	return testProviderConfig() + fmt.Sprintf(`
resource "kanidm_group" "test" {
  id          = %q
  description = %q
}
`, id, description)
}

func testAccGroupWithMembersConfig() string {
	return testProviderConfig() + `
resource "kanidm_person" "member" {
  id          = "acc-test-member"
  displayname = "Acc Test Member"
}

resource "kanidm_group" "test" {
  id          = "acc-test-group-members"
  description = "Group with members"
  members     = [kanidm_person.member.id]
}
`
}
