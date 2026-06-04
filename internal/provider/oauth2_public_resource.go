package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

var (
	_ resource.Resource                = (*oauth2PublicResource)(nil)
	_ resource.ResourceWithImportState = (*oauth2PublicResource)(nil)
)

func NewOAuth2PublicResource() resource.Resource {
	return &oauth2PublicResource{}
}

type oauth2PublicResource struct {
	resourceWithClient
}

type oauth2PublicResourceModel struct {
	Name                           types.String `tfsdk:"name"`
	DisplayName                    types.String `tfsdk:"displayname"`
	Origin                         types.String `tfsdk:"origin"`
	RedirectURIs                   types.List   `tfsdk:"redirect_uris"`
	ScopeMaps                      types.Set    `tfsdk:"scope_map"`
	SupScopeMaps                   types.Set    `tfsdk:"sup_scope_map"`
	ClaimMaps                      types.Set    `tfsdk:"claim_map"`
	ImagePath                      types.String `tfsdk:"image_path"`
	ImageSHA256                    types.String `tfsdk:"image_sha256"`
	AllowInsecureClientDisablePKCE types.Bool   `tfsdk:"allow_insecure_client_disable_pkce"`
	JwtLegacyCryptoEnable          types.Bool   `tfsdk:"jwt_legacy_crypto_enable"`
	PreferShortUsername            types.Bool   `tfsdk:"prefer_short_username"`
}

func (r *oauth2PublicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth2_public"
}

func (r *oauth2PublicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kanidm OAuth2 public client.\n\n" +
			"OAuth2 public clients are used for browser-based applications (SPAs) and native apps " +
			"that cannot securely store a client secret. PKCE is required and enforced.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the OAuth2 client (client ID). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name of the OAuth2 client.",
				Required:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "Origin URL where the OAuth2 client application is hosted (e.g., https://app.example.com).",
				Required:            true,
			},
			"redirect_uris": schema.ListAttribute{
				MarkdownDescription: "List of allowed redirect URIs for OAuth2 callbacks.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"allow_insecure_client_disable_pkce": schema.BoolAttribute{
				MarkdownDescription: "Allow this client to disable PKCE. Use only for legacy clients that cannot support PKCE.",
				Optional:            true,
				Computed:            true,
			},
			"jwt_legacy_crypto_enable": schema.BoolAttribute{
				MarkdownDescription: "Enable legacy RS256 JWT signing. Use for clients that do not support ES256.",
				Optional:            true,
				Computed:            true,
			},
			"prefer_short_username": schema.BoolAttribute{
				MarkdownDescription: "Return short username instead of SPN in the preferred_username OIDC claim.",
				Optional:            true,
				Computed:            true,
			},
			"image_path": schema.StringAttribute{
				MarkdownDescription: "Local file path to an image to upload as the OAuth2 client icon.",
				Optional:            true,
			},
			"image_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA-256 hash of the image file last uploaded. Used to detect changes.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"scope_map": schema.SetNestedBlock{
				MarkdownDescription: "Scope mappings that define which OAuth2 scopes are granted to members of specific groups.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							MarkdownDescription: "Name of the Kanidm group to map scopes to.",
							Required:            true,
						},
						"scopes": schema.ListAttribute{
							MarkdownDescription: "List of OAuth2 scopes to grant to group members.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"sup_scope_map": schema.SetNestedBlock{
				MarkdownDescription: "Supplemental scope mappings automatically granted to group members and cannot be declined by the user.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.StringAttribute{
							MarkdownDescription: "Name of the Kanidm group to map supplemental scopes to.",
							Required:            true,
						},
						"scopes": schema.SetAttribute{
							MarkdownDescription: "Set of OAuth2 scopes to automatically grant to group members.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"claim_map": schema.SetNestedBlock{
				MarkdownDescription: "Custom OIDC claim mappings for group members.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the custom OIDC claim.",
							Required:            true,
						},
						"group": schema.StringAttribute{
							MarkdownDescription: "Name of the Kanidm group to map claim values to.",
							Required:            true,
						},
						"values": schema.SetAttribute{
							MarkdownDescription: "Set of claim values to grant to group members.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"join": schema.StringAttribute{
							MarkdownDescription: "Join strategy for this claim. Valid values: csv, ssv, array.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *oauth2PublicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan oauth2PublicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating OAuth2 public client", map[string]any{"name": plan.Name.ValueString()})

	oauth2Client, err := r.client.CreateOAuth2PublicClient(
		ctx,
		plan.Name.ValueString(),
		plan.DisplayName.ValueString(),
		plan.Origin.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating OAuth2 Public Client", err.Error())
		return
	}

	var redirectURIs []string
	if !plan.RedirectURIs.IsNull() && !plan.RedirectURIs.IsUnknown() {
		resp.Diagnostics.Append(plan.RedirectURIs.ElementsAs(ctx, &redirectURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if err := r.client.UpdateOAuth2Client(
		ctx,
		oauth2Client.Name,
		plan.DisplayName.ValueString(),
		plan.Origin.ValueString(),
		redirectURIs,
		boolPtrFromTypes(plan.AllowInsecureClientDisablePKCE),
		boolPtrFromTypes(plan.JwtLegacyCryptoEnable),
		boolPtrFromTypes(plan.PreferShortUsername),
	); err != nil {
		resp.Diagnostics.AddError("Error Configuring OAuth2 Public Client", err.Error())
		return
	}

	if !plan.ScopeMaps.IsNull() && !plan.ScopeMaps.IsUnknown() {
		var scopeMaps []scopeMapModel
		resp.Diagnostics.Append(plan.ScopeMaps.ElementsAs(ctx, &scopeMaps, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, scopeMap := range scopeMaps {
			var scopes []string
			resp.Diagnostics.Append(scopeMap.Scopes.ElementsAs(ctx, &scopes, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			if err := r.client.SetOAuth2ScopeMap(ctx, oauth2Client.Name, scopeMap.Group.ValueString(), scopes); err != nil {
				resp.Diagnostics.AddError("Error Setting Scope Map", err.Error())
				return
			}
		}
	}

	if !plan.SupScopeMaps.IsNull() && !plan.SupScopeMaps.IsUnknown() {
		if diags := applySupScopeMaps(ctx, r.client, oauth2Client.Name, plan.SupScopeMaps); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	if !plan.ClaimMaps.IsNull() && !plan.ClaimMaps.IsUnknown() {
		if diags := applyClaimMaps(ctx, r.client, oauth2Client.Name, plan.ClaimMaps); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	if !plan.ImagePath.IsNull() && plan.ImagePath.ValueString() != "" {
		if err := r.client.UploadOAuth2Image(ctx, oauth2Client.Name, plan.ImagePath.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error Uploading OAuth2 Image", err.Error())
			return
		}
	}

	created, err := r.client.GetOAuth2Client(ctx, oauth2Client.Name)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OAuth2 Public Client", err.Error())
		return
	}

	plan.Name = types.StringValue(created.Name)
	plan.DisplayName = types.StringValue(created.DisplayName)
	plan.Origin = types.StringValue(created.Origin)

	if len(created.RedirectURIs) > 0 {
		list, diags := types.ListValueFrom(ctx, types.StringType, created.RedirectURIs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.RedirectURIs = list
	} else {
		plan.RedirectURIs = types.ListNull(types.StringType)
	}

	plan.AllowInsecureClientDisablePKCE = types.BoolValue(created.AllowInsecureClientDisablePKCE)
	plan.JwtLegacyCryptoEnable = types.BoolValue(created.JwtLegacyCryptoEnable)
	plan.PreferShortUsername = types.BoolValue(created.PreferShortUsername)
	plan.ImageSHA256 = computeImageSHA256(plan.ImagePath)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauth2PublicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state oauth2PublicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OAuth2 public client", map[string]any{"name": state.Name.ValueString()})

	oauth2Client, err := r.client.GetOAuth2Client(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading OAuth2 Public Client", err.Error())
		return
	}

	if !oauth2Client.IsPublic {
		resp.Diagnostics.AddError(
			"Invalid Client Type",
			"Expected OAuth2 public client but found confidential client. This resource manages public clients only.",
		)
		return
	}

	state.Name = types.StringValue(oauth2Client.Name)
	state.DisplayName = types.StringValue(oauth2Client.DisplayName)
	state.Origin = types.StringValue(oauth2Client.Origin)

	if len(oauth2Client.RedirectURIs) > 0 {
		list, diags := types.ListValueFrom(ctx, types.StringType, oauth2Client.RedirectURIs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.RedirectURIs = list
	} else {
		state.RedirectURIs = types.ListNull(types.StringType)
	}

	// scope_map, sup_scope_map, claim_map are preserved from state
	state.AllowInsecureClientDisablePKCE = types.BoolValue(oauth2Client.AllowInsecureClientDisablePKCE)
	state.JwtLegacyCryptoEnable = types.BoolValue(oauth2Client.JwtLegacyCryptoEnable)
	state.PreferShortUsername = types.BoolValue(oauth2Client.PreferShortUsername)
	state.ImageSHA256 = computeImageSHA256(state.ImagePath)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *oauth2PublicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state oauth2PublicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating OAuth2 public client", map[string]any{"name": plan.Name.ValueString()})

	var redirectURIs []string
	if !plan.RedirectURIs.IsNull() && !plan.RedirectURIs.IsUnknown() {
		resp.Diagnostics.Append(plan.RedirectURIs.ElementsAs(ctx, &redirectURIs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if err := r.client.UpdateOAuth2Client(
		ctx,
		plan.Name.ValueString(),
		plan.DisplayName.ValueString(),
		plan.Origin.ValueString(),
		redirectURIs,
		boolPtrFromTypes(plan.AllowInsecureClientDisablePKCE),
		boolPtrFromTypes(plan.JwtLegacyCryptoEnable),
		boolPtrFromTypes(plan.PreferShortUsername),
	); err != nil {
		resp.Diagnostics.AddError("Error Updating OAuth2 Public Client", err.Error())
		return
	}

	var oldScopeMaps, newScopeMaps []scopeMapModel
	resp.Diagnostics.Append(state.ScopeMaps.ElementsAs(ctx, &oldScopeMaps, false)...)
	resp.Diagnostics.Append(plan.ScopeMaps.ElementsAs(ctx, &newScopeMaps, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldByGroup := scopeMapsToGroupMap(ctx, oldScopeMaps, &resp.Diagnostics)
	newByGroup := scopeMapsToGroupMap(ctx, newScopeMaps, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	for group := range oldByGroup {
		if _, exists := newByGroup[group]; !exists {
			if err := r.client.DeleteOAuth2ScopeMap(ctx, plan.Name.ValueString(), group); err != nil && !errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("Error Deleting Scope Map", err.Error())
				return
			}
		}
	}
	for group, scopes := range newByGroup {
		if err := r.client.SetOAuth2ScopeMap(ctx, plan.Name.ValueString(), group, scopes); err != nil {
			resp.Diagnostics.AddError("Error Setting Scope Map", err.Error())
			return
		}
	}

	if diags := diffAndApplySupScopeMaps(ctx, r.client, plan.Name.ValueString(), state.SupScopeMaps, plan.SupScopeMaps); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if diags := diffAndApplyClaimMaps(ctx, r.client, plan.Name.ValueString(), state.ClaimMaps, plan.ClaimMaps); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if err := applyImageChange(ctx, r.client, plan.Name.ValueString(), state.ImageSHA256, plan.ImagePath); err != nil {
		resp.Diagnostics.AddError("Error Managing OAuth2 Image", err.Error())
		return
	}

	updated, err := r.client.GetOAuth2Client(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading OAuth2 Public Client", err.Error())
		return
	}

	plan.Name = types.StringValue(updated.Name)
	plan.DisplayName = types.StringValue(updated.DisplayName)
	plan.Origin = types.StringValue(updated.Origin)

	if len(updated.RedirectURIs) > 0 {
		list, diags := types.ListValueFrom(ctx, types.StringType, updated.RedirectURIs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.RedirectURIs = list
	} else {
		plan.RedirectURIs = types.ListNull(types.StringType)
	}

	plan.AllowInsecureClientDisablePKCE = types.BoolValue(updated.AllowInsecureClientDisablePKCE)
	plan.JwtLegacyCryptoEnable = types.BoolValue(updated.JwtLegacyCryptoEnable)
	plan.PreferShortUsername = types.BoolValue(updated.PreferShortUsername)
	plan.ImageSHA256 = computeImageSHA256(plan.ImagePath)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *oauth2PublicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state oauth2PublicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting OAuth2 public client", map[string]any{"name": state.Name.ValueString()})

	if err := r.client.DeleteOAuth2Client(ctx, state.Name.ValueString()); err != nil {
		if !errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Error Deleting OAuth2 Public Client", err.Error())
		}
	}
}

func (r *oauth2PublicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
