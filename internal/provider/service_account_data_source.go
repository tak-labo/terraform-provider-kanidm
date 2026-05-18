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

var _ datasource.DataSource = (*serviceAccountDataSource)(nil)

func NewServiceAccountDataSource() datasource.DataSource {
	return &serviceAccountDataSource{}
}

type serviceAccountDataSource struct {
	resourceWithClient
}

type serviceAccountDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	DisplayName    types.String `tfsdk:"displayname"`
	EntryManagedBy types.Set    `tfsdk:"entry_managed_by"`
}

func (d *serviceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *serviceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kanidm service account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the service account to look up.",
				Required:            true,
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name.",
				Computed:            true,
			},
			"entry_managed_by": schema.SetAttribute{
				MarkdownDescription: "Set of account or group IDs that manage this service account.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *serviceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serviceAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service account data source", map[string]any{"id": state.ID.ValueString()})

	sa, err := d.client.GetServiceAccount(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Service Account Not Found", "No service account found with id: "+state.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError("Error Reading Service Account", err.Error())
		return
	}

	state.DisplayName = types.StringValue(sa.DisplayName)

	if len(sa.EntryManagedBy) > 0 {
		managedBySet, diags := types.SetValueFrom(ctx, types.StringType, sa.EntryManagedBy)
		resp.Diagnostics.Append(diags...)
		state.EntryManagedBy = managedBySet
	} else {
		state.EntryManagedBy = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
