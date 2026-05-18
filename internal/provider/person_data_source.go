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

var _ datasource.DataSource = (*personDataSource)(nil)

func NewPersonDataSource() datasource.DataSource {
	return &personDataSource{}
}

type personDataSource struct {
	resourceWithClient
}

type personDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"displayname"`
	Mail        types.List   `tfsdk:"mail"`
	LegalName   types.String `tfsdk:"legalname"`
	UnixGID     types.Int64  `tfsdk:"unix_gid"`
	UnixShell   types.String `tfsdk:"unix_shell"`
}

func (d *personDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

func (d *personDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kanidm person account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Username of the person to look up.",
				Required:            true,
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name of the person.",
				Computed:            true,
			},
			"mail": schema.ListAttribute{
				MarkdownDescription: "Email addresses.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"legalname": schema.StringAttribute{
				MarkdownDescription: "Legal name.",
				Computed:            true,
			},
			"unix_gid": schema.Int64Attribute{
				MarkdownDescription: "Unix GID number.",
				Computed:            true,
			},
			"unix_shell": schema.StringAttribute{
				MarkdownDescription: "Login shell.",
				Computed:            true,
			},
		},
	}
}

func (d *personDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state personDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading person data source", map[string]any{"id": state.ID.ValueString()})

	person, err := d.client.GetPerson(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Person Not Found", "No person found with id: "+state.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError("Error Reading Person", err.Error())
		return
	}

	state.DisplayName = types.StringValue(person.DisplayName)

	if len(person.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, person.Mail)
		resp.Diagnostics.Append(diags...)
		state.Mail = mailList
	} else {
		state.Mail = types.ListNull(types.StringType)
	}

	if person.LegalName != "" {
		state.LegalName = types.StringValue(person.LegalName)
	} else {
		state.LegalName = types.StringNull()
	}

	if person.UnixGID != nil {
		state.UnixGID = types.Int64Value(*person.UnixGID)
	} else {
		state.UnixGID = types.Int64Null()
	}

	if person.UnixShell != "" {
		state.UnixShell = types.StringValue(person.UnixShell)
	} else {
		state.UnixShell = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
