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

var _ resource.Resource = &IssueTypeResource{}
var _ resource.ResourceWithImportState = &IssueTypeResource{}

type IssueTypeResource struct {
	client *client.Client
}

type IssueTypeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func NewIssueTypeResource() resource.Resource {
	return &IssueTypeResource{}
}

func (r *IssueTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue_type"
}

func (r *IssueTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA issue type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The issue type ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The issue type name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The issue type description.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The hierarchy level: 'standard' or 'subtask'.",
				Required:    true,
			},
		},
	}
}

func (r *IssueTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IssueTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name":  plan.Name.ValueString(),
		"type":  plan.Type.ValueString(),
		"scope": map[string]interface{}{"type": "GLOBAL"},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/issuetype", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating issue type", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IssueTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IssueTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/issuetype/%s", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading issue type", err.Error())
		return
	}

	state.Name = types.StringValue(fmt.Sprintf("%v", result["name"]))
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	// Determine type from subtask field
	if subtask, ok := result["subtask"].(bool); ok && subtask {
		state.Type = types.StringValue("subtask")
	} else {
		state.Type = types.StringValue("standard")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *IssueTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IssueTypeResourceModel
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

	err := r.client.Put(fmt.Sprintf("/rest/api/3/issuetype/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating issue type", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IssueTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IssueTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/issuetype/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting issue type", err.Error())
		return
	}
}

func (r *IssueTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/issuetype/%s", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing issue type", err.Error())
		return
	}

	state := IssueTypeResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%v", result["id"])),
		Name: types.StringValue(fmt.Sprintf("%v", result["name"])),
	}
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	if subtask, ok := result["subtask"].(bool); ok && subtask {
		state.Type = types.StringValue("subtask")
	} else {
		state.Type = types.StringValue("standard")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
