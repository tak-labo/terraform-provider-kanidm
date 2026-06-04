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
	_ resource.Resource                = (*applicationResource)(nil)
	_ resource.ResourceWithImportState = (*applicationResource)(nil)
)

func NewApplicationResource() resource.Resource {
	return &applicationResource{}
}

type applicationResource struct {
	resourceWithClient
}

type applicationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"displayname"`
	LinkedGroup types.String `tfsdk:"linked_group"`
}

func (r *applicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *applicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kanidm application entry.\n\n" +
			"Applications are used for per-user application passwords in legacy LDAP/basic-auth style integrations. " +
			"Each application links to a single Kanidm group; only members of that group may mint application-specific credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable Kanidm UUID for this application.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Unique application name. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name for the application.",
				Required:            true,
			},
			"linked_group": schema.StringAttribute{
				MarkdownDescription: "UUID of the Kanidm group whose members may use this application.",
				Required:            true,
			},
		},
	}
}

func (r *applicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating application", map[string]any{"name": plan.Name.ValueString()})

	app, err := r.client.CreateApplication(
		ctx,
		plan.Name.ValueString(),
		plan.DisplayName.ValueString(),
		plan.LinkedGroup.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Application", err.Error())
		return
	}

	plan.ID = types.StringValue(app.ID)
	plan.Name = types.StringValue(app.Name)
	plan.DisplayName = types.StringValue(app.DisplayName)
	plan.LinkedGroup = types.StringValue(app.LinkedGroup)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading application", map[string]any{"id": state.ID.ValueString()})

	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Application", err.Error())
		return
	}

	state.ID = types.StringValue(app.ID)
	state.Name = types.StringValue(app.Name)
	state.DisplayName = types.StringValue(app.DisplayName)
	state.LinkedGroup = types.StringValue(app.LinkedGroup)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state applicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating application", map[string]any{"id": state.ID.ValueString()})

	app, err := r.client.UpdateApplication(
		ctx,
		state.ID.ValueString(),
		plan.DisplayName.ValueString(),
		plan.LinkedGroup.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Application", err.Error())
		return
	}

	plan.ID = types.StringValue(app.ID)
	plan.Name = types.StringValue(app.Name)
	plan.DisplayName = types.StringValue(app.DisplayName)
	plan.LinkedGroup = types.StringValue(app.LinkedGroup)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state applicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting application", map[string]any{"id": state.ID.ValueString()})

	if err := r.client.DeleteApplication(ctx, state.ID.ValueString()); err != nil {
		if !errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError("Error Deleting Application", err.Error())
		}
	}
}

func (r *applicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
