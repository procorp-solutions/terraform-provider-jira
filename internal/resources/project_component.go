package resources

import (
	"context"
	"fmt"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectComponentResource{}
var _ resource.ResourceWithImportState = &ProjectComponentResource{}

type ProjectComponentResource struct {
	client *client.Client
}

type ProjectComponentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectKey    types.String `tfsdk:"project_key"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	LeadAccountID types.String `tfsdk:"lead_account_id"`
	AssigneeType  types.String `tfsdk:"assignee_type"`
}

func NewProjectComponentResource() resource.Resource {
	return &ProjectComponentResource{}
}

func (r *ProjectComponentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_component"
}

func (r *ProjectComponentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA project component.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The component ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "The project key this component belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The component name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The component description.",
				Optional:    true,
			},
			"lead_account_id": schema.StringAttribute{
				Description: "The Atlassian account ID of the component lead.",
				Optional:    true,
			},
			"assignee_type": schema.StringAttribute{
				Description: "Default assignee type: PROJECT_DEFAULT, COMPONENT_LEAD, PROJECT_LEAD, or UNASSIGNED.",
				Optional:    true,
			},
		},
	}
}

func (r *ProjectComponentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client.")
		return
	}
	r.client = c
}

func (r *ProjectComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"project": plan.ProjectKey.ValueString(),
		"name":    plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.LeadAccountID.IsNull() && !plan.LeadAccountID.IsUnknown() {
		body["leadAccountId"] = plan.LeadAccountID.ValueString()
	}
	if !plan.AssigneeType.IsNull() && !plan.AssigneeType.IsUnknown() {
		body["assigneeType"] = plan.AssigneeType.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/component", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project component", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProjectComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectComponentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/component/%s", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project component", err.Error())
		return
	}

	state.Name = types.StringValue(fmt.Sprintf("%v", result["name"]))
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	if lead, ok := result["lead"].(map[string]interface{}); ok {
		state.LeadAccountID = types.StringValue(fmt.Sprintf("%v", lead["accountId"]))
	}
	if at, ok := result["assigneeType"].(string); ok && at != "" {
		state.AssigneeType = types.StringValue(at)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ProjectComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.LeadAccountID.IsNull() && !plan.LeadAccountID.IsUnknown() {
		body["leadAccountId"] = plan.LeadAccountID.ValueString()
	}
	if !plan.AssigneeType.IsNull() && !plan.AssigneeType.IsUnknown() {
		body["assigneeType"] = plan.AssigneeType.ValueString()
	}

	err := r.client.Put(fmt.Sprintf("/rest/api/3/component/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project component", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProjectComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectComponentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/component/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting project component", err.Error())
		return
	}
}

func (r *ProjectComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/component/%s", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing project component", err.Error())
		return
	}

	state := ProjectComponentResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%v", result["id"])),
		Name: types.StringValue(fmt.Sprintf("%v", result["name"])),
	}
	if proj, ok := result["project"].(string); ok {
		state.ProjectKey = types.StringValue(proj)
	}
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	if lead, ok := result["lead"].(map[string]interface{}); ok {
		state.LeadAccountID = types.StringValue(fmt.Sprintf("%v", lead["accountId"]))
	}
	if at, ok := result["assigneeType"].(string); ok && at != "" {
		state.AssigneeType = types.StringValue(at)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
