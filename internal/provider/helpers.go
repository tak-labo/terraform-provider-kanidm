package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

// resourceWithClient is embedded in all resource structs to share Configure.
type resourceWithClient struct {
	client *client.Client
}

func (r *resourceWithClient) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *client.Client. Please report this issue to the provider developers.",
		)
		return
	}
	r.client = c
}

// boolPtrFromTypes converts a types.Bool to *bool, returning nil when null or unknown.
func boolPtrFromTypes(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// scopeMapsToGroupMap converts a slice of scopeMapModel to a map[group][]scopes.
func scopeMapsToGroupMap(ctx context.Context, scopeMaps []scopeMapModel, diags *diag.Diagnostics) map[string][]string {
	result := make(map[string][]string, len(scopeMaps))
	for _, sm := range scopeMaps {
		var scopes []string
		diags.Append(sm.Scopes.ElementsAs(ctx, &scopes, false)...)
		result[sm.Group.ValueString()] = scopes
	}
	return result
}
