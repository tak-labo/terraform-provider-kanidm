package provider

import (
	"context"
	"errors"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

var (
	_ resource.Resource                = (*accountPolicyResource)(nil)
	_ resource.ResourceWithImportState = (*accountPolicyResource)(nil)
)

func NewAccountPolicyResource() resource.Resource {
	return &accountPolicyResource{}
}

type accountPolicyResource struct {
	resourceWithClient
}

type accountPolicyResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Group                     types.String `tfsdk:"group"`
	AuthSessionExpiry         types.Int64  `tfsdk:"authsession_expiry"`
	PrivilegeExpiry           types.Int64  `tfsdk:"privilege_expiry"`
	AuthPasswordMinimumLength types.Int64  `tfsdk:"auth_password_minimum_length"`
	CredentialTypeMinimum     types.String `tfsdk:"credential_type_minimum"`
	LimitSearchMaxResults     types.Int64  `tfsdk:"limit_search_max_results"`
	LimitSearchMaxFilterTest  types.Int64  `tfsdk:"limit_search_max_filter_test"`
	AllowPrimaryCredFallback  types.Bool   `tfsdk:"allow_primary_cred_fallback"`
}

func (r *accountPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_policy"
}

func (r *accountPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages account policy on a Kanidm group.\n\n" +
			"Account policy defines security requirements for accounts that are members of the group. " +
			"Policy is resolved across all groups a user belongs to, with the strictest setting winning.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "Group name, SPN, or UUID. The account_policy class will be enabled on this group.",
				Required:            true,
			},
			"authsession_expiry": schema.Int64Attribute{
				MarkdownDescription: "Maximum authentication session duration in seconds.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"privilege_expiry": schema.Int64Attribute{
				MarkdownDescription: "Maximum privileged session duration in seconds (maximum 3600).",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(3600),
				},
			},
			"auth_password_minimum_length": schema.Int64Attribute{
				MarkdownDescription: "Minimum character length for passwords.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"credential_type_minimum": schema.StringAttribute{
				MarkdownDescription: "Minimum credential strength. Valid values: any, mfa, passkey, attested_passkey.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("any", "mfa", "passkey", "attested_passkey"),
				},
			},
			"limit_search_max_results": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of query results returned in a single search operation.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"limit_search_max_filter_test": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of entries examined in a partially indexed query.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"allow_primary_cred_fallback": schema.BoolAttribute{
				MarkdownDescription: "Allow fallback to primary password for LDAP authentication when no POSIX password exists.",
				Optional:            true,
			},
		},
	}
}

func (r *accountPolicyResource) resolveGroupUUID(ctx context.Context, identifier string) (string, error) {
	group, err := r.client.GetGroup(ctx, identifier)
	if err != nil {
		return "", err
	}
	if group.UUID == "" {
		return "", errors.New("group did not return a UUID")
	}
	return group.UUID, nil
}

func (r *accountPolicyResource) readPolicyAttrs(ctx context.Context, groupUUID string, model *accountPolicyResourceModel) {
	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "authsession_expiry"); err == nil && len(vals) > 0 {
		if v, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			model.AuthSessionExpiry = types.Int64Value(v)
		}
	} else {
		model.AuthSessionExpiry = types.Int64Null()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "privilege_expiry"); err == nil && len(vals) > 0 {
		if v, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			model.PrivilegeExpiry = types.Int64Value(v)
		}
	} else {
		model.PrivilegeExpiry = types.Int64Null()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "auth_password_minimum_length"); err == nil && len(vals) > 0 {
		if v, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			model.AuthPasswordMinimumLength = types.Int64Value(v)
		}
	} else {
		model.AuthPasswordMinimumLength = types.Int64Null()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "credential_type_minimum"); err == nil && len(vals) > 0 {
		model.CredentialTypeMinimum = types.StringValue(vals[0])
	} else {
		model.CredentialTypeMinimum = types.StringNull()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "limit_search_max_results"); err == nil && len(vals) > 0 {
		if v, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			model.LimitSearchMaxResults = types.Int64Value(v)
		}
	} else {
		model.LimitSearchMaxResults = types.Int64Null()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "limit_search_max_filter_test"); err == nil && len(vals) > 0 {
		if v, err := strconv.ParseInt(vals[0], 10, 64); err == nil {
			model.LimitSearchMaxFilterTest = types.Int64Value(v)
		}
	} else {
		model.LimitSearchMaxFilterTest = types.Int64Null()
	}

	if vals, err := r.client.GetGroupAttr(ctx, groupUUID, "allow_primary_cred_fallback"); err == nil && len(vals) > 0 {
		model.AllowPrimaryCredFallback = types.BoolValue(vals[0] == "true")
	} else {
		model.AllowPrimaryCredFallback = types.BoolNull()
	}
}

func (r *accountPolicyResource) setPolicyAttrs(ctx context.Context, groupUUID string, plan accountPolicyResourceModel) error {
	type attrSetter struct {
		name  string
		value string
		set   bool
	}

	setters := []attrSetter{}

	if !plan.AuthSessionExpiry.IsNull() && !plan.AuthSessionExpiry.IsUnknown() {
		setters = append(setters, attrSetter{"authsession_expiry", strconv.FormatInt(plan.AuthSessionExpiry.ValueInt64(), 10), true})
	}
	if !plan.PrivilegeExpiry.IsNull() && !plan.PrivilegeExpiry.IsUnknown() {
		setters = append(setters, attrSetter{"privilege_expiry", strconv.FormatInt(plan.PrivilegeExpiry.ValueInt64(), 10), true})
	}
	if !plan.AuthPasswordMinimumLength.IsNull() && !plan.AuthPasswordMinimumLength.IsUnknown() {
		setters = append(setters, attrSetter{"auth_password_minimum_length", strconv.FormatInt(plan.AuthPasswordMinimumLength.ValueInt64(), 10), true})
	}
	if !plan.CredentialTypeMinimum.IsNull() && !plan.CredentialTypeMinimum.IsUnknown() {
		setters = append(setters, attrSetter{"credential_type_minimum", plan.CredentialTypeMinimum.ValueString(), true})
	}
	if !plan.LimitSearchMaxResults.IsNull() && !plan.LimitSearchMaxResults.IsUnknown() {
		setters = append(setters, attrSetter{"limit_search_max_results", strconv.FormatInt(plan.LimitSearchMaxResults.ValueInt64(), 10), true})
	}
	if !plan.LimitSearchMaxFilterTest.IsNull() && !plan.LimitSearchMaxFilterTest.IsUnknown() {
		setters = append(setters, attrSetter{"limit_search_max_filter_test", strconv.FormatInt(plan.LimitSearchMaxFilterTest.ValueInt64(), 10), true})
	}
	if !plan.AllowPrimaryCredFallback.IsNull() && !plan.AllowPrimaryCredFallback.IsUnknown() {
		setters = append(setters, attrSetter{"allow_primary_cred_fallback", strconv.FormatBool(plan.AllowPrimaryCredFallback.ValueBool()), true})
	}

	for _, s := range setters {
		if err := r.client.SetGroupAttr(ctx, groupUUID, s.name, []string{s.value}); err != nil {
			return err
		}
	}
	return nil
}

func (r *accountPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupUUID, err := r.resolveGroupUUID(ctx, plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Resolving Group", err.Error())
		return
	}

	tflog.Debug(ctx, "Enabling account policy on group", map[string]any{"group": groupUUID})

	classes, err := r.client.GetGroupAttr(ctx, groupUUID, "class")
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Group Classes", err.Error())
		return
	}
	hasPolicy := false
	for _, c := range classes {
		if c == "account_policy" {
			hasPolicy = true
			break
		}
	}
	if !hasPolicy {
		if err := r.client.AddGroupClass(ctx, groupUUID, "account_policy"); err != nil {
			resp.Diagnostics.AddError("Error Enabling Account Policy", err.Error())
			return
		}
	}

	if err := r.setPolicyAttrs(ctx, groupUUID, plan); err != nil {
		resp.Diagnostics.AddError("Error Setting Account Policy Attributes", err.Error())
		return
	}

	plan.ID = types.StringValue(groupUUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accountPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accountPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading account policy", map[string]any{"id": state.ID.ValueString()})

	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Group", err.Error())
		return
	}

	state.Group = types.StringValue(group.ID)
	state.ID = types.StringValue(group.UUID)

	r.readPolicyAttrs(ctx, group.UUID, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *accountPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state accountPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupUUID := state.ID.ValueString()

	if !plan.Group.Equal(state.Group) {
		newUUID, err := r.resolveGroupUUID(ctx, plan.Group.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error Resolving Group", err.Error())
			return
		}
		groupUUID = newUUID
	}

	tflog.Debug(ctx, "Updating account policy", map[string]any{"group": groupUUID})

	clearIfNull := func(attrName string, oldVal, newVal bool) {
		if !newVal && oldVal {
			_ = r.client.DeleteGroupAttr(ctx, groupUUID, attrName)
		}
	}

	if !plan.AuthSessionExpiry.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "authsession_expiry", []string{strconv.FormatInt(plan.AuthSessionExpiry.ValueInt64(), 10)}); err != nil {
			resp.Diagnostics.AddError("Error Setting authsession_expiry", err.Error())
			return
		}
	} else {
		clearIfNull("authsession_expiry", !state.AuthSessionExpiry.IsNull(), true)
	}

	if !plan.PrivilegeExpiry.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "privilege_expiry", []string{strconv.FormatInt(plan.PrivilegeExpiry.ValueInt64(), 10)}); err != nil {
			resp.Diagnostics.AddError("Error Setting privilege_expiry", err.Error())
			return
		}
	} else {
		clearIfNull("privilege_expiry", !state.PrivilegeExpiry.IsNull(), true)
	}

	if !plan.AuthPasswordMinimumLength.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "auth_password_minimum_length", []string{strconv.FormatInt(plan.AuthPasswordMinimumLength.ValueInt64(), 10)}); err != nil {
			resp.Diagnostics.AddError("Error Setting auth_password_minimum_length", err.Error())
			return
		}
	} else {
		clearIfNull("auth_password_minimum_length", !state.AuthPasswordMinimumLength.IsNull(), true)
	}

	if !plan.CredentialTypeMinimum.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "credential_type_minimum", []string{plan.CredentialTypeMinimum.ValueString()}); err != nil {
			resp.Diagnostics.AddError("Error Setting credential_type_minimum", err.Error())
			return
		}
	} else {
		clearIfNull("credential_type_minimum", !state.CredentialTypeMinimum.IsNull(), true)
	}

	if !plan.LimitSearchMaxResults.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "limit_search_max_results", []string{strconv.FormatInt(plan.LimitSearchMaxResults.ValueInt64(), 10)}); err != nil {
			resp.Diagnostics.AddError("Error Setting limit_search_max_results", err.Error())
			return
		}
	} else {
		clearIfNull("limit_search_max_results", !state.LimitSearchMaxResults.IsNull(), true)
	}

	if !plan.LimitSearchMaxFilterTest.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "limit_search_max_filter_test", []string{strconv.FormatInt(plan.LimitSearchMaxFilterTest.ValueInt64(), 10)}); err != nil {
			resp.Diagnostics.AddError("Error Setting limit_search_max_filter_test", err.Error())
			return
		}
	} else {
		clearIfNull("limit_search_max_filter_test", !state.LimitSearchMaxFilterTest.IsNull(), true)
	}

	if !plan.AllowPrimaryCredFallback.IsNull() {
		if err := r.client.SetGroupAttr(ctx, groupUUID, "allow_primary_cred_fallback", []string{strconv.FormatBool(plan.AllowPrimaryCredFallback.ValueBool())}); err != nil {
			resp.Diagnostics.AddError("Error Setting allow_primary_cred_fallback", err.Error())
			return
		}
	} else {
		clearIfNull("allow_primary_cred_fallback", !state.AllowPrimaryCredFallback.IsNull(), true)
	}

	plan.ID = types.StringValue(groupUUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accountPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accountPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupUUID := state.ID.ValueString()
	tflog.Debug(ctx, "Removing account policy from group", map[string]any{"group": groupUUID})

	for _, attr := range []string{
		"authsession_expiry", "privilege_expiry", "auth_password_minimum_length",
		"credential_type_minimum", "limit_search_max_results", "limit_search_max_filter_test",
		"allow_primary_cred_fallback",
	} {
		_ = r.client.DeleteGroupAttr(ctx, groupUUID, attr)
	}

	if err := r.client.RemoveGroupClass(ctx, groupUUID, "account_policy"); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error Removing Account Policy Class", err.Error())
	}
}

func (r *accountPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
