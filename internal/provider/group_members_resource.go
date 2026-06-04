package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

var _ resource.Resource = (*groupMembersResource)(nil)

func NewGroupMembersResource() resource.Resource {
	return &groupMembersResource{}
}

type groupMembersResource struct {
	resourceWithClient
}

type groupMembersResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Group   types.String `tfsdk:"group"`
	Members types.Set    `tfsdk:"members"`
}

func (r *groupMembersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_members"
}

func (r *groupMembersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a subset of group memberships without taking ownership of the group's full member list.\n\n" +
			"Use this for built-in or externally managed groups where Terraform should only add/remove the listed members. " +
			"Do not use this resource against the same target group as `kanidm_group.members`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stable Kanidm UUID for the target group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "Target group identifier. May be a built-in group name, SPN, UUID, or a referenced `kanidm_group` ID.",
				Required:            true,
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "Set of members this resource should ensure are present in the target group. Only this managed subset is added/removed.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *groupMembersResource) resolveGroupUUID(ctx context.Context, identifier string) (string, string, error) {
	group, err := r.client.GetGroup(ctx, identifier)
	if err != nil {
		return "", "", err
	}
	if group.UUID == "" {
		return "", "", errors.New("group did not return a UUID")
	}
	return group.UUID, group.ID, nil
}

func (r *groupMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupMembersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupUUID, _, err := r.resolveGroupUUID(ctx, plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Resolving Group", err.Error())
		return
	}

	var desiredMembers []string
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &desiredMembers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(desiredMembers) > 0 {
		if err := r.client.AddGroupMembers(ctx, groupUUID, desiredMembers); err != nil {
			resp.Diagnostics.AddError("Error Adding Group Members", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(groupUUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupMembersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error Reading Group", err.Error())
		return
	}

	// Track only the members we manage (intersection of state.Members and group.Members)
	var trackedMembers []string
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &trackedMembers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentMemberSet := make(map[string]struct{}, len(group.Members))
	for _, m := range group.Members {
		currentMemberSet[m] = struct{}{}
	}

	managedMembers := make([]string, 0, len(trackedMembers))
	for _, m := range trackedMembers {
		if _, exists := currentMemberSet[m]; exists {
			managedMembers = append(managedMembers, m)
		}
	}

	state.ID = types.StringValue(group.UUID)
	membersSet, diags := stringsToSet(ctx, managedMembers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Members = membersSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupMembersResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldGroupUUID := state.ID.ValueString()
	newGroupUUID, _, err := r.resolveGroupUUID(ctx, plan.Group.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Resolving Group", err.Error())
		return
	}

	var previousMembers, desiredMembers []string
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &previousMembers, false)...)
	resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &desiredMembers, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if oldGroupUUID != newGroupUUID {
		if len(previousMembers) > 0 {
			if err := r.client.RemoveGroupMembers(ctx, oldGroupUUID, previousMembers); err != nil && !errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("Error Removing Group Members", err.Error())
				return
			}
		}
		if len(desiredMembers) > 0 {
			if err := r.client.AddGroupMembers(ctx, newGroupUUID, desiredMembers); err != nil {
				resp.Diagnostics.AddError("Error Adding Group Members", err.Error())
				return
			}
		}
	} else {
		toAdd := stringSetDiff(desiredMembers, previousMembers)
		toRemove := stringSetDiff(previousMembers, desiredMembers)

		if len(toAdd) > 0 {
			if err := r.client.AddGroupMembers(ctx, newGroupUUID, toAdd); err != nil {
				resp.Diagnostics.AddError("Error Adding Group Members", err.Error())
				return
			}
		}
		if len(toRemove) > 0 {
			if err := r.client.RemoveGroupMembers(ctx, newGroupUUID, toRemove); err != nil && !errors.Is(err, client.ErrNotFound) {
				resp.Diagnostics.AddError("Error Removing Group Members", err.Error())
				return
			}
		}
	}

	plan.ID = types.StringValue(newGroupUUID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupMembersResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var members []string
	resp.Diagnostics.Append(state.Members.ElementsAs(ctx, &members, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(members) == 0 {
		return
	}

	if err := r.client.RemoveGroupMembers(ctx, state.ID.ValueString(), members); err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error Removing Group Members", err.Error())
		return
	}

	tflog.Debug(ctx, "Removed managed group membership subset", map[string]any{"group_id": state.ID.ValueString()})
}
