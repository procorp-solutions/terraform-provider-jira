package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

type ProjectResource struct {
	client *client.Client
}

type ProjectResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Key                 types.String `tfsdk:"key"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	ProjectTypeKey      types.String `tfsdk:"project_type_key"`
	LeadAccountID       types.String `tfsdk:"lead_account_id"`
	AssigneeType        types.String `tfsdk:"assignee_type"`
	IssueTypeSchemeID   types.String `tfsdk:"issue_type_scheme_id"`
	PermissionSchemeID  types.String `tfsdk:"permission_scheme_id"`
	WorkflowSchemeID    types.String `tfsdk:"workflow_scheme_id"`
}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The project ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "The project key (e.g. PROJ). Must be unique and uppercase.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The project name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The project description.",
				Optional:    true,
			},
			"project_type_key": schema.StringAttribute{
				Description: "The project type: software, business, or service_desk.",
				Required:    true,
			},
			"lead_account_id": schema.StringAttribute{
				Description: "The Atlassian account ID of the project lead.",
				Required:    true,
			},
			"assignee_type": schema.StringAttribute{
				Description: "Default assignee type: PROJECT_LEAD or UNASSIGNED.",
				Optional:    true,
			},
			"issue_type_scheme_id": schema.StringAttribute{
				Description: "Issue type scheme ID. Use the ID from jira_issue_type_scheme.",
				Optional:    true,
			},
			"permission_scheme_id": schema.StringAttribute{
				Description: "Permission scheme ID. Use the ID from jira_permission_scheme.",
				Optional:    true,
			},
			"workflow_scheme_id": schema.StringAttribute{
				Description: "Workflow scheme ID. Use the ID from jira_workflow_scheme.",
				Optional:    true,
			},
		},
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			"Expected *client.Client, got unexpected type.")
		return
	}
	r.client = c
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"key":            plan.Key.ValueString(),
		"name":           plan.Name.ValueString(),
		"projectTypeKey": plan.ProjectTypeKey.ValueString(),
		"leadAccountId":  plan.LeadAccountID.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.AssigneeType.IsNull() && !plan.AssigneeType.IsUnknown() {
		body["assigneeType"] = plan.AssigneeType.ValueString()
	}
	if !plan.IssueTypeSchemeID.IsNull() && !plan.IssueTypeSchemeID.IsUnknown() {
		id, err := strconv.ParseInt(plan.IssueTypeSchemeID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("issue_type_scheme_id"), "Invalid issue type scheme ID",
				"Scheme ID must be a numeric string (e.g. from jira_issue_type_scheme.id).")
			return
		}
		body["issueTypeScheme"] = id
	}
	if !plan.PermissionSchemeID.IsNull() && !plan.PermissionSchemeID.IsUnknown() {
		id, err := strconv.ParseInt(plan.PermissionSchemeID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("permission_scheme_id"), "Invalid permission scheme ID",
				"Scheme ID must be a numeric string (e.g. from jira_permission_scheme.id).")
			return
		}
		body["permissionScheme"] = id
	}
	if !plan.WorkflowSchemeID.IsNull() && !plan.WorkflowSchemeID.IsUnknown() {
		id, err := strconv.ParseInt(plan.WorkflowSchemeID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("workflow_scheme_id"), "Invalid workflow scheme ID",
				"Scheme ID must be a numeric string (e.g. from jira_workflow_scheme.id).")
			return
		}
		body["workflowScheme"] = id
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/project", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/project/%s", state.Key.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	state.Key = types.StringValue(fmt.Sprintf("%v", result["key"]))
	state.Name = types.StringValue(fmt.Sprintf("%v", result["name"]))
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}
	state.ProjectTypeKey = types.StringValue(fmt.Sprintf("%v", result["projectTypeKey"]))
	if lead, ok := result["lead"].(map[string]interface{}); ok {
		state.LeadAccountID = types.StringValue(fmt.Sprintf("%v", lead["accountId"]))
	}
	if at, ok := result["assigneeType"].(string); ok && at != "" {
		state.AssigneeType = types.StringValue(at)
	}
	if id := schemeIDFromResponse(result, "issueTypeScheme"); id != "" {
		state.IssueTypeSchemeID = types.StringValue(id)
	}
	if id := schemeIDFromResponse(result, "permissionScheme"); id != "" {
		state.PermissionSchemeID = types.StringValue(id)
	}
	if id := schemeIDFromResponse(result, "workflowScheme"); id != "" {
		state.WorkflowSchemeID = types.StringValue(id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// schemeIDFromResponse extracts a scheme ID from the GET project response.
// The API may return an object like {"id": "10011"} or a direct value.
func schemeIDFromResponse(result map[string]interface{}, key string) string {
	v, ok := result[key]
	if !ok || v == nil {
		return ""
	}
	if obj, ok := v.(map[string]interface{}); ok {
		if id, ok := obj["id"]; ok && id != nil {
			return fmt.Sprintf("%v", id)
		}
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"key":            plan.Key.ValueString(),
		"name":           plan.Name.ValueString(),
		"projectTypeKey": plan.ProjectTypeKey.ValueString(),
		"leadAccountId":  plan.LeadAccountID.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.AssigneeType.IsNull() && !plan.AssigneeType.IsUnknown() {
		body["assigneeType"] = plan.AssigneeType.ValueString()
	}
	// Jira Cloud PUT project does not accept issueTypeScheme, permissionScheme, or workflowScheme.
	var result map[string]interface{}
	err := r.client.Put(fmt.Sprintf("/rest/api/3/project/%s", plan.Key.ValueString()), body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	// Assign issue type scheme via dedicated endpoint when set or changed.
	if !plan.IssueTypeSchemeID.IsNull() && !plan.IssueTypeSchemeID.IsUnknown() {
		assignBody := map[string]interface{}{
			"issueTypeSchemeId": plan.IssueTypeSchemeID.ValueString(),
			"projectId":         plan.ID.ValueString(),
		}
		if err := r.client.Put("/rest/api/3/issuetypescheme/project", assignBody, nil); err != nil {
			resp.Diagnostics.AddError("Error assigning issue type scheme to project", err.Error())
			return
		}
	}

	// Assign permission scheme via dedicated endpoint when set.
	if !plan.PermissionSchemeID.IsNull() && !plan.PermissionSchemeID.IsUnknown() {
		id, err := strconv.ParseInt(plan.PermissionSchemeID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("permission_scheme_id"), "Invalid permission scheme ID",
				"Scheme ID must be a numeric string (e.g. from jira_permission_scheme.id).")
			return
		}
		permBody := map[string]interface{}{"id": id}
		if err := r.client.Put(fmt.Sprintf("/rest/api/3/project/%s/permissionscheme", plan.Key.ValueString()), permBody, nil); err != nil {
			resp.Diagnostics.AddError("Error assigning permission scheme to project", err.Error())
			return
		}
	}

	// Assign workflow scheme via dedicated endpoint when set.
	if !plan.WorkflowSchemeID.IsNull() && !plan.WorkflowSchemeID.IsUnknown() {
		workflowBody := map[string]interface{}{
			"projectId":         plan.ID.ValueString(),
			"workflowSchemeId": plan.WorkflowSchemeID.ValueString(),
		}
		if err := r.client.Put("/rest/api/3/workflowscheme/project", workflowBody, nil); err != nil {
			resp.Diagnostics.AddError("Error assigning workflow scheme to project", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/project/%s", state.Key.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting project", err.Error())
		return
	}
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by project key
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), req.ID)...)

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/project/%s", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing project", err.Error())
		return
	}

	state := ProjectResourceModel{
		ID:             types.StringValue(fmt.Sprintf("%v", result["id"])),
		Key:            types.StringValue(fmt.Sprintf("%v", result["key"])),
		Name:           types.StringValue(fmt.Sprintf("%v", result["name"])),
		ProjectTypeKey: types.StringValue(fmt.Sprintf("%v", result["projectTypeKey"])),
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
	if id := schemeIDFromResponse(result, "issueTypeScheme"); id != "" {
		state.IssueTypeSchemeID = types.StringValue(id)
	}
	if id := schemeIDFromResponse(result, "permissionScheme"); id != "" {
		state.PermissionSchemeID = types.StringValue(id)
	}
	if id := schemeIDFromResponse(result, "workflowScheme"); id != "" {
		state.WorkflowSchemeID = types.StringValue(id)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
