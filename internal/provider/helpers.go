package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

// applySupScopeMaps sets all supplemental scope maps for an OAuth2 client.
func applySupScopeMaps(ctx context.Context, c *client.Client, rsName string, supScopeMaps types.Set) diag.Diagnostics {
	var diags diag.Diagnostics
	var maps []supScopeMapModel
	diags.Append(supScopeMaps.ElementsAs(ctx, &maps, false)...)
	if diags.HasError() {
		return diags
	}
	for _, m := range maps {
		var scopes []string
		diags.Append(m.Scopes.ElementsAs(ctx, &scopes, false)...)
		if diags.HasError() {
			return diags
		}
		if err := c.SetOAuth2SupScopeMap(ctx, rsName, m.Group.ValueString(), scopes); err != nil {
			diags.AddError("Error Setting Supplemental Scope Map", err.Error())
			return diags
		}
	}
	return diags
}

// diffAndApplySupScopeMaps reconciles supplemental scope maps between old and new state.
func diffAndApplySupScopeMaps(ctx context.Context, c *client.Client, rsName string, oldSet, newSet types.Set) diag.Diagnostics {
	var diags diag.Diagnostics
	var oldMaps, newMaps []supScopeMapModel
	diags.Append(oldSet.ElementsAs(ctx, &oldMaps, false)...)
	diags.Append(newSet.ElementsAs(ctx, &newMaps, false)...)
	if diags.HasError() {
		return diags
	}

	newByGroup := make(map[string][]string, len(newMaps))
	for _, m := range newMaps {
		var scopes []string
		diags.Append(m.Scopes.ElementsAs(ctx, &scopes, false)...)
		newByGroup[m.Group.ValueString()] = scopes
	}

	for _, m := range oldMaps {
		if _, exists := newByGroup[m.Group.ValueString()]; !exists {
			if err := c.DeleteOAuth2SupScopeMap(ctx, rsName, m.Group.ValueString()); err != nil {
				diags.AddError("Error Deleting Supplemental Scope Map", err.Error())
				return diags
			}
		}
	}
	for group, scopes := range newByGroup {
		if err := c.SetOAuth2SupScopeMap(ctx, rsName, group, scopes); err != nil {
			diags.AddError("Error Setting Supplemental Scope Map", err.Error())
			return diags
		}
	}
	return diags
}

// applyClaimMaps sets all claim maps for an OAuth2 client.
func applyClaimMaps(ctx context.Context, c *client.Client, rsName string, claimMaps types.Set) diag.Diagnostics {
	var diags diag.Diagnostics
	var maps []claimMapModel
	diags.Append(claimMaps.ElementsAs(ctx, &maps, false)...)
	if diags.HasError() {
		return diags
	}

	joinByName := make(map[string]string)
	for _, m := range maps {
		var values []string
		diags.Append(m.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return diags
		}
		if err := c.SetOAuth2ClaimMap(ctx, rsName, m.Name.ValueString(), m.Group.ValueString(), values); err != nil {
			diags.AddError("Error Setting Claim Map", err.Error())
			return diags
		}
		joinByName[m.Name.ValueString()] = m.Join.ValueString()
	}

	for name, join := range joinByName {
		if err := c.SetOAuth2ClaimMapJoin(ctx, rsName, name, join); err != nil {
			diags.AddError("Error Setting Claim Map Join", err.Error())
			return diags
		}
	}
	return diags
}

// diffAndApplyClaimMaps reconciles claim maps between old and new state.
func diffAndApplyClaimMaps(ctx context.Context, c *client.Client, rsName string, oldSet, newSet types.Set) diag.Diagnostics {
	var diags diag.Diagnostics
	var oldMaps, newMaps []claimMapModel
	diags.Append(oldSet.ElementsAs(ctx, &oldMaps, false)...)
	diags.Append(newSet.ElementsAs(ctx, &newMaps, false)...)
	if diags.HasError() {
		return diags
	}

	type claimKey struct{ name, group string }

	newByKey := make(map[claimKey]claimMapModel, len(newMaps))
	for _, m := range newMaps {
		newByKey[claimKey{m.Name.ValueString(), m.Group.ValueString()}] = m
	}

	for _, m := range oldMaps {
		k := claimKey{m.Name.ValueString(), m.Group.ValueString()}
		if _, exists := newByKey[k]; !exists {
			if err := c.DeleteOAuth2ClaimMap(ctx, rsName, m.Name.ValueString(), m.Group.ValueString()); err != nil {
				diags.AddError("Error Deleting Claim Map", err.Error())
				return diags
			}
		}
	}

	joinByName := make(map[string]string)
	for _, m := range newMaps {
		var values []string
		diags.Append(m.Values.ElementsAs(ctx, &values, false)...)
		if diags.HasError() {
			return diags
		}
		if err := c.SetOAuth2ClaimMap(ctx, rsName, m.Name.ValueString(), m.Group.ValueString(), values); err != nil {
			diags.AddError("Error Setting Claim Map", err.Error())
			return diags
		}
		joinByName[m.Name.ValueString()] = m.Join.ValueString()
	}

	for name, join := range joinByName {
		if err := c.SetOAuth2ClaimMapJoin(ctx, rsName, name, join); err != nil {
			diags.AddError("Error Setting Claim Map Join", err.Error())
			return diags
		}
	}
	return diags
}

// applyImageChange uploads or deletes the OAuth2 client image based on plan vs state SHA256.
func applyImageChange(ctx context.Context, c *client.Client, rsName string, stateSHA256, planImagePath types.String) error {
	planPath := planImagePath.ValueString()
	if planPath == "" {
		// Image removed: delete if there was one before
		if !stateSHA256.IsNull() && stateSHA256.ValueString() != "" {
			return c.DeleteOAuth2Image(ctx, rsName)
		}
		return nil
	}
	// Image specified: upload if hash changed
	newHash, err := client.HashFileSHA256(planPath)
	if err != nil {
		return err
	}
	if stateSHA256.IsNull() || stateSHA256.ValueString() != newHash {
		return c.UploadOAuth2Image(ctx, rsName, planPath)
	}
	return nil
}

// computeImageSHA256 calculates the SHA256 of the image file at imagePath, returning null if empty/error.
func computeImageSHA256(imagePath types.String) types.String {
	if imagePath.IsNull() || imagePath.IsUnknown() || imagePath.ValueString() == "" {
		return types.StringNull()
	}
	hash, err := client.HashFileSHA256(imagePath.ValueString())
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(hash)
}

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

// stringSetDiff returns elements in left that are not in right.
func stringSetDiff(left, right []string) []string {
	if len(left) == 0 {
		return []string{}
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, v := range right {
		rightSet[v] = struct{}{}
	}
	diff := make([]string, 0, len(left))
	for _, v := range left {
		if _, exists := rightSet[v]; !exists {
			diff = append(diff, v)
		}
	}
	return diff
}

// stringsToSet converts a string slice to a types.Set, returning a null Set for empty slices.
func stringsToSet(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if len(values) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{}), nil
	}
	return types.SetValueFrom(ctx, types.StringType, values)
}
