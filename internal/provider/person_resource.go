package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tak-labo/terraform-provider-kanidm/internal/client"
)

// Ensure the implementation satisfies the required interfaces
var (
	_ resource.Resource                = (*personResource)(nil)
	_ resource.ResourceWithImportState = (*personResource)(nil)
)

// NewPersonResource creates a new person resource
func NewPersonResource() resource.Resource {
	return &personResource{}
}

// personResource is the resource implementation
type personResource struct {
	resourceWithClient
}

// personResourceModel describes the resource data model
type personResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	DisplayName                  types.String `tfsdk:"displayname"`
	Mail                         types.List   `tfsdk:"mail"`
	LegalName                    types.String `tfsdk:"legalname"`
	Password                     types.String `tfsdk:"password"`
	GenerateCredentialResetToken types.Bool   `tfsdk:"generate_credential_reset_token"`
	CredentialResetToken         types.String `tfsdk:"credential_reset_token"`
	CredentialResetTokenTTL      types.Int64  `tfsdk:"credential_reset_token_ttl"`
	UnixGID                      types.Int64  `tfsdk:"unix_gid"`
	UnixShell                    types.String `tfsdk:"unix_shell"`
}

// Metadata returns the resource type name
func (r *personResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

// Schema defines the schema for the resource
func (r *personResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Kanidm person account.

## Authentication Setup

Kanidm supports two credential setup workflows:

### Password-Based Authentication
Set the ` + "`password`" + ` attribute to create a password-based account:

` + "```hcl" + `
resource "kanidm_person" "example" {
  id          = "jdoe"
  displayname = "John Doe"
  password    = var.initial_password
}
` + "```" + `

### Passkey/Modern Authentication (Recommended)
Set ` + "`generate_credential_reset_token = true`" + ` to generate a one-time token for credential setup via the Kanidm web UI:

` + "```hcl" + `
resource "kanidm_person" "example" {
  id                            = "jdoe"
  displayname                   = "John Doe"
  generate_credential_reset_token = true
}

output "credential_reset_token" {
  value     = kanidm_person.example.credential_reset_token
  sensitive = true
}
` + "```" + `

The user can then visit the Kanidm web UI with the token to set up passkeys or passwords.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier for the person account (username). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"displayname": schema.StringAttribute{
				MarkdownDescription: "Display name of the person.",
				Required:            true,
			},
			"mail": schema.ListAttribute{
				MarkdownDescription: "Email addresses for the person.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for the person account. **Note:** This is write-only and will not be stored in state. " +
					"Mutually exclusive with `generate_credential_reset_token`. " +
					"Consider using `lifecycle { ignore_changes = [password] }` if the password is managed externally.",
				Optional:  true,
				Sensitive: true,
			},
			"generate_credential_reset_token": schema.BoolAttribute{
				MarkdownDescription: "Whether to generate a credential reset token for passkey/password setup via the web UI. " +
					"Mutually exclusive with `password`. Defaults to `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"credential_reset_token": schema.StringAttribute{
				MarkdownDescription: "The credential reset token (generated when `generate_credential_reset_token` is `true`). " +
					"This token can be used once to set up credentials via the Kanidm web UI. **Computed value only.**",
				Computed:  true,
				Sensitive: true,
			},
			"credential_reset_token_ttl": schema.Int64Attribute{
				MarkdownDescription: "Time-to-live for the credential reset token in seconds. Defaults to 3600 (1 hour).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
			},
			"legalname": schema.StringAttribute{
				MarkdownDescription: "Legal name of the person (full legal name as it appears on official documents).",
				Optional:            true,
			},
			"unix_gid": schema.Int64Attribute{
				MarkdownDescription: "Unix GID number for Linux/PAM authentication. Enables Unix account integration.",
				Optional:            true,
			},
			"unix_shell": schema.StringAttribute{
				MarkdownDescription: "Login shell for Unix/PAM authentication (e.g. `/bin/bash`). Requires `unix_gid` to be set.",
				Optional:            true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state
func (r *personResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan personResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate mutually exclusive options
	hasPassword := !plan.Password.IsNull() && !plan.Password.IsUnknown()
	generateToken := plan.GenerateCredentialResetToken.ValueBool()

	if hasPassword && generateToken {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both 'password' and 'generate_credential_reset_token'. Choose one authentication setup method.",
		)
		return
	}

	tflog.Debug(ctx, "Creating person", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Create the person account
	person, err := r.client.CreatePerson(ctx, plan.ID.ValueString(), plan.DisplayName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Person",
			"Could not create person: "+err.Error(),
		)
		return
	}

	// Set password if provided
	if hasPassword {
		tflog.Debug(ctx, "Setting initial password for person")
		if err := r.client.SetPersonPassword(ctx, person.ID, plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Password",
				"Person was created but password could not be set: "+err.Error(),
			)
			return
		}
	}

	// Generate credential reset token if requested
	if generateToken {
		tflog.Debug(ctx, "Generating credential reset token for person")
		ttl := int(plan.CredentialResetTokenTTL.ValueInt64())
		token, err := r.client.CreatePersonCredentialResetToken(ctx, person.ID, &ttl)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Generating Credential Reset Token",
				"Person was created but credential reset token could not be generated: "+err.Error(),
			)
			return
		}
		plan.CredentialResetToken = types.StringValue(token)
	}

	// Update mail and legalname if provided
	var mailAddrs []string
	if !plan.Mail.IsNull() && !plan.Mail.IsUnknown() {
		resp.Diagnostics.Append(plan.Mail.ElementsAs(ctx, &mailAddrs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var legalName *string
	if !plan.LegalName.IsNull() && !plan.LegalName.IsUnknown() {
		v := plan.LegalName.ValueString()
		legalName = &v
	}

	if len(mailAddrs) > 0 || legalName != nil {
		if err := r.client.UpdatePerson(ctx, person.ID, "", mailAddrs, legalName); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Person Attributes",
				"Person was created but attributes could not be set: "+err.Error(),
			)
			return
		}
	}

	// Apply Unix extension if unix_gid is set
	if !plan.UnixGID.IsNull() && !plan.UnixGID.IsUnknown() {
		gid := plan.UnixGID.ValueInt64()
		var shell *string
		if !plan.UnixShell.IsNull() && !plan.UnixShell.IsUnknown() {
			v := plan.UnixShell.ValueString()
			shell = &v
		}
		if err := r.client.UnixExtendPerson(ctx, person.ID, &gid, shell); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Unix Attributes",
				"Person was created but Unix attributes could not be set: "+err.Error(),
			)
			return
		}
	}

	// Read back the person to get the current state
	createdPerson, err := r.client.GetPerson(ctx, person.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Person was created but could not be read back: "+err.Error(),
		)
		return
	}

	// Map response to state
	plan.ID = types.StringValue(createdPerson.ID)
	plan.DisplayName = types.StringValue(createdPerson.DisplayName)

	if len(createdPerson.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, createdPerson.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mail = mailList
	}

	if createdPerson.LegalName != "" {
		plan.LegalName = types.StringValue(createdPerson.LegalName)
	}

	if createdPerson.UnixGID != nil {
		plan.UnixGID = types.Int64Value(*createdPerson.UnixGID)
	}
	if createdPerson.UnixShell != "" {
		plan.UnixShell = types.StringValue(createdPerson.UnixShell)
	}

	// Password is write-only, keep the planned value but don't try to read it back

	// Ensure credential_reset_token fields are properly set with defaults if not already set
	if plan.GenerateCredentialResetToken.IsNull() || plan.GenerateCredentialResetToken.IsUnknown() {
		plan.GenerateCredentialResetToken = types.BoolValue(false)
	}
	if plan.CredentialResetTokenTTL.IsNull() || plan.CredentialResetTokenTTL.IsUnknown() {
		plan.CredentialResetTokenTTL = types.Int64Value(3600)
	}
	// If credential_reset_token wasn't generated, ensure it's null not unknown
	if plan.CredentialResetToken.IsUnknown() {
		plan.CredentialResetToken = types.StringNull()
	}

	tflog.Debug(ctx, "Person created successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *personResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state personResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading person", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Get current person from API
	person, err := r.client.GetPerson(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			tflog.Warn(ctx, "Person not found, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Could not read person: "+err.Error(),
		)
		return
	}

	// Update state with current values
	state.ID = types.StringValue(person.ID)
	state.DisplayName = types.StringValue(person.DisplayName)

	if len(person.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, person.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
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

	// Password is write-only and not readable from API, preserve existing state value
	// credential_reset_token fields should use defaults when not explicitly set
	if state.GenerateCredentialResetToken.IsNull() || state.GenerateCredentialResetToken.IsUnknown() {
		state.GenerateCredentialResetToken = types.BoolValue(false)
	}
	if state.CredentialResetTokenTTL.IsNull() || state.CredentialResetTokenTTL.IsUnknown() {
		state.CredentialResetTokenTTL = types.Int64Value(3600)
	}
	// credential_reset_token is only set during Create/Update when generated, otherwise null
	if state.CredentialResetToken.IsUnknown() {
		state.CredentialResetToken = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state
func (r *personResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state personResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating person", map[string]any{
		"id": plan.ID.ValueString(),
	})

	// Prepare mail addresses
	var mailAddrs []string
	if !plan.Mail.IsNull() && !plan.Mail.IsUnknown() {
		resp.Diagnostics.Append(plan.Mail.ElementsAs(ctx, &mailAddrs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var legalName *string
	if !plan.LegalName.IsNull() && !plan.LegalName.IsUnknown() {
		v := plan.LegalName.ValueString()
		legalName = &v
	}

	// Update person attributes (displayname, mail, legalname)
	if err := r.client.UpdatePerson(ctx, plan.ID.ValueString(), plan.DisplayName.ValueString(), mailAddrs, legalName); err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Person",
			"Could not update person: "+err.Error(),
		)
		return
	}

	// Update password if changed
	if !plan.Password.Equal(state.Password) && !plan.Password.IsNull() {
		tflog.Debug(ctx, "Updating password for person")
		if err := r.client.SetPersonPassword(ctx, plan.ID.ValueString(), plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Password",
				"Person was updated but password could not be changed: "+err.Error(),
			)
			return
		}
	}

	// Apply Unix extension if unix_gid changed
	if !plan.UnixGID.Equal(state.UnixGID) || !plan.UnixShell.Equal(state.UnixShell) {
		if !plan.UnixGID.IsNull() && !plan.UnixGID.IsUnknown() {
			gid := plan.UnixGID.ValueInt64()
			var shell *string
			if !plan.UnixShell.IsNull() && !plan.UnixShell.IsUnknown() {
				v := plan.UnixShell.ValueString()
				shell = &v
			}
			if err := r.client.UnixExtendPerson(ctx, plan.ID.ValueString(), &gid, shell); err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Unix Attributes",
					"Could not update Unix attributes: "+err.Error(),
				)
				return
			}
		}
	}

	// Generate new credential reset token if requested and changed
	if plan.GenerateCredentialResetToken.ValueBool() && !plan.GenerateCredentialResetToken.Equal(state.GenerateCredentialResetToken) {
		tflog.Debug(ctx, "Generating new credential reset token for person")
		ttl := int(plan.CredentialResetTokenTTL.ValueInt64())
		token, err := r.client.CreatePersonCredentialResetToken(ctx, plan.ID.ValueString(), &ttl)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Generating Credential Reset Token",
				"Person was updated but credential reset token could not be generated: "+err.Error(),
			)
			return
		}
		plan.CredentialResetToken = types.StringValue(token)
	}

	// Read back the updated person
	updatedPerson, err := r.client.GetPerson(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Person",
			"Person was updated but could not be read back: "+err.Error(),
		)
		return
	}

	// Update state
	plan.ID = types.StringValue(updatedPerson.ID)
	plan.DisplayName = types.StringValue(updatedPerson.DisplayName)

	if len(updatedPerson.Mail) > 0 {
		mailList, diags := types.ListValueFrom(ctx, types.StringType, updatedPerson.Mail)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Mail = mailList
	} else {
		plan.Mail = types.ListNull(types.StringType)
	}

	if updatedPerson.LegalName != "" {
		plan.LegalName = types.StringValue(updatedPerson.LegalName)
	} else {
		plan.LegalName = types.StringNull()
	}

	if updatedPerson.UnixGID != nil {
		plan.UnixGID = types.Int64Value(*updatedPerson.UnixGID)
	} else {
		plan.UnixGID = types.Int64Null()
	}

	if updatedPerson.UnixShell != "" {
		plan.UnixShell = types.StringValue(updatedPerson.UnixShell)
	} else {
		plan.UnixShell = types.StringNull()
	}

	// Ensure credential_reset_token fields are properly set
	if plan.GenerateCredentialResetToken.IsNull() || plan.GenerateCredentialResetToken.IsUnknown() {
		plan.GenerateCredentialResetToken = types.BoolValue(false)
	}
	if plan.CredentialResetTokenTTL.IsNull() || plan.CredentialResetTokenTTL.IsUnknown() {
		plan.CredentialResetTokenTTL = types.Int64Value(3600)
	}
	// credential_reset_token is only set during Update when generated, otherwise null
	if plan.CredentialResetToken.IsUnknown() {
		plan.CredentialResetToken = types.StringNull()
	}

	tflog.Debug(ctx, "Person updated successfully", map[string]any{
		"id": plan.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state
func (r *personResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state personResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting person", map[string]any{
		"id": state.ID.ValueString(),
	})

	// Delete the person
	if err := r.client.DeletePerson(ctx, state.ID.ValueString()); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			// Person already deleted, just remove from state
			tflog.Warn(ctx, "Person not found during delete, removing from state", map[string]any{
				"id": state.ID.ValueString(),
			})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Person",
			"Could not delete person: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Person deleted successfully", map[string]any{
		"id": state.ID.ValueString(),
	})
}

// ImportState imports an existing person into Terraform state
func (r *personResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID (username) directly as the import identifier
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Imported person", map[string]any{
		"id": req.ID,
	})
}
