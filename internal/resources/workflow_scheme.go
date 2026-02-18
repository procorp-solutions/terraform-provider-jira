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

var _ resource.Resource = &WorkflowSchemeResource{}
var _ resource.ResourceWithImportState = &WorkflowSchemeResource{}

type WorkflowSchemeResource struct {
	client *client.Client
}

type WorkflowSchemeResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	DefaultWorkflow   types.String `tfsdk:"default_workflow"`
	IssueTypeMappings types.Map    `tfsdk:"issue_type_mappings"`
}

func NewWorkflowSchemeResource() resource.Resource {
	return &WorkflowSchemeResource{}
}

func (r *WorkflowSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_scheme"
}

func (r *WorkflowSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA workflow scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The workflow scheme ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The workflow scheme name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The workflow scheme description.",
				Optional:    true,
			},
			"default_workflow": schema.StringAttribute{
				Description: "The name of the default workflow.",
				Optional:    true,
			},
			"issue_type_mappings": schema.MapAttribute{
				Description: "Map of issue type ID to workflow name. Overrides the default workflow for specific issue types.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *WorkflowSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WorkflowSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowSchemeResourceModel
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
	if !plan.DefaultWorkflow.IsNull() && !plan.DefaultWorkflow.IsUnknown() {
		body["defaultWorkflow"] = plan.DefaultWorkflow.ValueString()
	}
	if !plan.IssueTypeMappings.IsNull() && !plan.IssueTypeMappings.IsUnknown() {
		mappings := make(map[string]string)
		resp.Diagnostics.Append(plan.IssueTypeMappings.ElementsAs(ctx, &mappings, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Convert to JIRA format: issueTypeMappings
		itMappings := make(map[string]string)
		for k, v := range mappings {
			itMappings[k] = v
		}
		body["issueTypeMappings"] = itMappings
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/workflowscheme", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow scheme", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WorkflowSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading workflow scheme", err.Error())
		return
	}

	state.Name = types.StringValue(fmt.Sprintf("%v", result["name"]))
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	if dw, ok := result["defaultWorkflow"].(string); ok && dw != "" {
		state.DefaultWorkflow = types.StringValue(dw)
	}
	if itm, ok := result["issueTypeMappings"].(map[string]interface{}); ok && len(itm) > 0 {
		mappings := make(map[string]string)
		for k, v := range itm {
			mappings[k] = fmt.Sprintf("%v", v)
		}
		mapVal, diags := types.MapValueFrom(ctx, types.StringType, mappings)
		resp.Diagnostics.Append(diags...)
		state.IssueTypeMappings = mapVal
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *WorkflowSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowSchemeResourceModel
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
	if !plan.DefaultWorkflow.IsNull() && !plan.DefaultWorkflow.IsUnknown() {
		body["defaultWorkflow"] = plan.DefaultWorkflow.ValueString()
	}
	if !plan.IssueTypeMappings.IsNull() && !plan.IssueTypeMappings.IsUnknown() {
		mappings := make(map[string]string)
		resp.Diagnostics.Append(plan.IssueTypeMappings.ElementsAs(ctx, &mappings, false)...)
		body["issueTypeMappings"] = mappings
	}

	err := r.client.Put(fmt.Sprintf("/rest/api/3/workflowscheme/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workflow scheme", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *WorkflowSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/workflowscheme/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow scheme", err.Error())
		return
	}
}

func (r *WorkflowSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/workflowscheme/%s", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing workflow scheme", err.Error())
		return
	}

	state := WorkflowSchemeResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%v", result["id"])),
		Name: types.StringValue(fmt.Sprintf("%v", result["name"])),
	}
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	if dw, ok := result["defaultWorkflow"].(string); ok && dw != "" {
		state.DefaultWorkflow = types.StringValue(dw)
	}
	if itm, ok := result["issueTypeMappings"].(map[string]interface{}); ok && len(itm) > 0 {
		mappings := make(map[string]string)
		for k, v := range itm {
			mappings[k] = fmt.Sprintf("%v", v)
		}
		mapVal, diags := types.MapValueFrom(ctx, types.StringType, mappings)
		resp.Diagnostics.Append(diags...)
		state.IssueTypeMappings = mapVal
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
