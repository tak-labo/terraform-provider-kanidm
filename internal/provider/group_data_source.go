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

var _ datasource.DataSource = (*groupDataSource)(nil)

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

type groupDataSource struct {
	resourceWithClient
}

type groupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Members     types.Set    `tfsdk:"members"`
	UnixGID     types.Int64  `tfsdk:"unix_gid"`
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kanidm group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Name of the group to look up.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the group.",
				Computed:            true,
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Set of member IDs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"unix_gid": schema.Int64Attribute{
				MarkdownDescription: "Unix GID number.",
				Computed:            true,
			},
		},
	}
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading group data source", map[string]any{"id": state.ID.ValueString()})

	group, err := d.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Group Not Found", "No group found with id: "+state.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError("Error Reading Group", err.Error())
		return
	}

	state.Description = types.StringValue(group.Description)

	if len(group.Members) > 0 {
		membersSet, diags := types.SetValueFrom(ctx, types.StringType, group.Members)
		resp.Diagnostics.Append(diags...)
		state.Members = membersSet
	} else {
		state.Members = types.SetNull(types.StringType)
	}

	if group.UnixGID != nil {
		state.UnixGID = types.Int64Value(*group.UnixGID)
	} else {
		state.UnixGID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
