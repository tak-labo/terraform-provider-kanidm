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
	_ resource.Resource                = (*groupResource)(nil)
	_ resource.ResourceWithImportState = (*groupResource)(nil)
)

func NewGroupResource() resource.Resource {
	return &groupResource{}
}

type groupResource struct {
	resourceWithClient
}

type groupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Members     types.Set    `tfsdk:"members"`
	UnixGID     types.Int64  `tfsdk:"unix_gid"`
}

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kanidm group.\n\nGroups are used to organize users and service accounts, and control access to resources.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the group (group name). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the group.",
				Optional:            true,
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Set of member IDs (persons or service accounts). " +
					"Members are managed as a complete set - any changes will replace all members.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"unix_gid": schema.Int64Attribute{
				MarkdownDescription: "Unix GID number for Linux/PAM group authentication.",
				Optional:            true,
			},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating group", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Create the group
	description := ""
	if !plan.Description.IsNull() {
		description = plan.Description.ValueString()
	}

	group, err := r.client.CreateGroup(ctx, plan.ID.ValueString(), description)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Group",
			"Could not create group: "+err.Error(),
		)
		return
	}

	// Add members if provided
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		var memberIDs []string
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &memberIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(memberIDs) > 0 {
			tflog.Debug(ctx, "Adding members to group", map[string]any{
				"count": len(memberIDs),
			})
			if err := r.client.UpdateGroup(ctx, group.ID, "", memberIDs); err != nil {
				resp.Diagnostics.AddError(
					"Error Adding Members",
					"Group was created but members could not be added: "+err.Error(),
				)
				return
			}
		}
	}

	// Read back the created group
	createdGroup, err := r.client.GetGroup(ctx, group.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group",
			"Group was created but could not be read back: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(createdGroup.ID)
	plan.Description = types.StringValue(createdGroup.Description)

	// Preserve null when no members were specified and none were returned.
	if !plan.Members.IsNull() || len(createdGroup.Members) > 0 {
		membersSet, diags := types.SetValueFrom(ctx, types.StringType, createdGroup.Members)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Members = membersSet
	}

	// Apply Unix extension if unix_gid is set
	if !plan.UnixGID.IsNull() && !plan.UnixGID.IsUnknown() {
		gid := plan.UnixGID.ValueInt64()
		if err := r.client.UnixExtendGroup(ctx, group.ID, &gid); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Unix GID",
				"Group was created but Unix GID could not be set: "+err.Error(),
			)
			return
		}
	}

	if createdGroup.UnixGID != nil {
		plan.UnixGID = types.Int64Value(*createdGroup.UnixGID)
	}

	tflog.Debug(ctx, "Group created successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading group", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get current group from API
	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Group not found, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Group",
			"Could not read group: "+err.Error(),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(group.ID)
	state.Description = types.StringValue(group.Description)

	// Preserve null when no members are configured and none were returned.
	if !state.Members.IsNull() || len(group.Members) > 0 {
		membersSet, diags := types.SetValueFrom(ctx, types.StringType, group.Members)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Members = membersSet
	}

	if group.UnixGID != nil {
		state.UnixGID = types.Int64Value(*group.UnixGID)
	} else {
		state.UnixGID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating group", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Prepare members list
	var memberIDs []string
	if !plan.Members.IsNull() && !plan.Members.IsUnknown() {
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &memberIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update group (description and members)
	description := ""
	if !plan.Description.IsNull() {
		description = plan.Description.ValueString()
	}

	if err := r.client.UpdateGroup(ctx, plan.ID.ValueString(), description, memberIDs); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Group",
			"Could not update group: "+err.Error(),
		)
		return
	}

	// Apply Unix extension if unix_gid changed
	if !plan.UnixGID.Equal(state.UnixGID) && !plan.UnixGID.IsNull() && !plan.UnixGID.IsUnknown() {
		gid := plan.UnixGID.ValueInt64()
		if err := r.client.UnixExtendGroup(ctx, plan.ID.ValueString(), &gid); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Unix GID",
				"Could not update Unix GID: "+err.Error(),
			)
			return
		}
	}

	// Read back the updated group
	updatedGroup, err := r.client.GetGroup(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Group",
			"Group was updated but could not be read back: "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = types.StringValue(updatedGroup.ID)
	plan.Description = types.StringValue(updatedGroup.Description)

	// Preserve null when no members are configured and none were returned.
	if !plan.Members.IsNull() || len(updatedGroup.Members) > 0 {
		membersSet, diags := types.SetValueFrom(ctx, types.StringType, updatedGroup.Members)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Members = membersSet
	}

	if updatedGroup.UnixGID != nil {
		plan.UnixGID = types.Int64Value(*updatedGroup.UnixGID)
	} else {
		plan.UnixGID = types.Int64Null()
	}

	tflog.Debug(ctx, "Group updated successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting group", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the group
	if err := r.client.DeleteGroup(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Group not found during delete, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Group",
			"Could not delete group: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Group deleted successfully", map[string]any{
		"id": state.ID.ValueString(),
	})
}

func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID (group name) directly as the import identifier
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Imported group", map[string]any{
		"id": req.ID,
	})
}
