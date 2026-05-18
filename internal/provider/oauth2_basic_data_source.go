package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

var _ datasource.DataSource = (*oauth2BasicDataSource)(nil)

func NewOAuth2BasicDataSource() datasource.DataSource {
	return &oauth2BasicDataSource{}
}

type oauth2BasicDataSource struct {
	resourceWithClient
}

type oauth2BasicDataSourceModel struct {
	Name                           types.String `tfsdk:"name"`
	DisplayName                    types.String `tfsdk:"displayname"`
	Origin                         types.String `tfsdk:"origin"`
	RedirectURIs                   types.List   `tfsdk:"redirect_uris"`
	AllowInsecureClientDisablePKCE types.Bool   `tfsdk:"allow_insecure_client_disable_pkce"`
	JwtLegacyCryptoEnable          types.Bool   `tfsdk:"jwt_legacy_crypto_enable"`
	PreferShortUsername            types.Bool   `tfsdk:"prefer_short_username"`
}

func (d *oauth2BasicDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth2_basic"
}

func (d *oauth2BasicDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kanidm OAuth2 basic (confidential) client.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the OAuth2 client to look up.",
				Required:            true,
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name.",
				Computed:            true,
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "Origin URL.",
				Computed:            true,
			},
			"redirect_uris": schema.ListAttribute{
				MarkdownDescription: "Allowed redirect URIs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allow_insecure_client_disable_pkce": schema.BoolAttribute{
				MarkdownDescription: "Whether PKCE can be disabled.",
				Computed:            true,
			},
			"jwt_legacy_crypto_enable": schema.BoolAttribute{
				MarkdownDescription: "Whether legacy RS256 JWT signing is enabled.",
				Computed:            true,
			},
			"prefer_short_username": schema.BoolAttribute{
				MarkdownDescription: "Whether short username is returned in preferred_username claim.",
				Computed:            true,
			},
		},
	}
}

func (d *oauth2BasicDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oauth2BasicDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OAuth2 basic data source", map[string]any{"name": state.Name.ValueString()})

	oauth2, err := d.client.GetOAuth2Client(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("OAuth2 Client Not Found", "No OAuth2 client found with name: "+state.Name.ValueString())
			return
		}
		resp.Diagnostics.AddError("Error Reading OAuth2 Client", err.Error())
		return
	}

	state.DisplayName = types.StringValue(oauth2.DisplayName)
	state.Origin = types.StringValue(oauth2.Origin)
	state.AllowInsecureClientDisablePKCE = types.BoolValue(oauth2.AllowInsecureClientDisablePKCE)
	state.JwtLegacyCryptoEnable = types.BoolValue(oauth2.JwtLegacyCryptoEnable)
	state.PreferShortUsername = types.BoolValue(oauth2.PreferShortUsername)

	if len(oauth2.RedirectURIs) > 0 {
		uriList, diags := types.ListValueFrom(ctx, types.StringType, oauth2.RedirectURIs)
		resp.Diagnostics.Append(diags...)
		state.RedirectURIs = uriList
	} else {
		state.RedirectURIs = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
