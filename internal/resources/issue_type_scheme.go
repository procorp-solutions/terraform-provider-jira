package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/david/terraform-provider-jira/internal/issuetype"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &IssueTypeSchemeResource{}
var _ resource.ResourceWithImportState = &IssueTypeSchemeResource{}

type IssueTypeSchemeResource struct {
	client *client.Client
}

type IssueTypeSchemeResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	DefaultIssueTypeID types.String `tfsdk:"default_issue_type_id"`
	IssueTypeIDs       types.List   `tfsdk:"issue_type_ids"`
}

func NewIssueTypeSchemeResource() resource.Resource {
	return &IssueTypeSchemeResource{}
}

func (r *IssueTypeSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue_type_scheme"
}

func (r *IssueTypeSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA issue type scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The issue type scheme ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The issue type scheme name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The issue type scheme description.",
				Optional:    true,
			},
			"default_issue_type_id": schema.StringAttribute{
				Description: "The default issue type ID for this scheme.",
				Optional:    true,
			},
			"issue_type_ids": schema.ListAttribute{
				Description: "List of issue type IDs in this scheme.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *IssueTypeSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IssueTypeSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IssueTypeSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var issueTypeIDs []string
	resp.Diagnostics.Append(plan.IssueTypeIDs.ElementsAs(ctx, &issueTypeIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invalidIssueTypeIDs, err := r.findProjectScopedIssueTypeIDs(issueTypeIDs, plan.DefaultIssueTypeID)
	if err != nil {
		resp.Diagnostics.AddError("Error validating issue types for issue type scheme", err.Error())
		return
	}
	if len(invalidIssueTypeIDs) > 0 {
		resp.Diagnostics.AddError(
			"Invalid issue types for classic issue type scheme",
			fmt.Sprintf(
				"Issue type IDs %s are project-scoped (next-gen/team-managed) and cannot be used in jira_issue_type_scheme. Use global issue type IDs instead.",
				strings.Join(invalidIssueTypeIDs, ", "),
			),
		)
		return
	}

	body := map[string]interface{}{
		"name":         plan.Name.ValueString(),
		"issueTypeIds": issueTypeIDs,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.DefaultIssueTypeID.IsNull() && !plan.DefaultIssueTypeID.IsUnknown() {
		body["defaultIssueTypeId"] = plan.DefaultIssueTypeID.ValueString()
	}

	var result map[string]interface{}
	err = r.client.Post("/rest/api/3/issuetypescheme", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating issue type scheme", err.Error())
		return
	}

	// The response may contain issueTypeSchemeId
	if id, ok := result["issueTypeSchemeId"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", id))
	} else if id, ok := result["id"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", id))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IssueTypeSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IssueTypeSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get all issue type schemes and find ours
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/issuetypescheme?id=%s", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading issue type scheme", err.Error())
		return
	}

	if values, ok := result["values"].([]interface{}); ok && len(values) > 0 {
		scheme := values[0].(map[string]interface{})
		state.Name = types.StringValue(fmt.Sprintf("%v", scheme["name"]))
		if desc, ok := scheme["description"].(string); ok && desc != "" {
			state.Description = types.StringValue(desc)
		}
		if ditID, ok := scheme["defaultIssueTypeId"].(string); ok && ditID != "" {
			state.DefaultIssueTypeID = types.StringValue(ditID)
		}
	}

	// Get issue type IDs for this scheme
	var itemsResult map[string]interface{}
	err = r.client.Get(fmt.Sprintf("/rest/api/3/issuetypescheme/mapping?issueTypeSchemeId=%s", state.ID.ValueString()), &itemsResult)
	if err == nil {
		if values, ok := itemsResult["values"].([]interface{}); ok {
			var ids []string
			for _, v := range values {
				item := v.(map[string]interface{})
				ids = append(ids, fmt.Sprintf("%v", item["issueTypeId"]))
			}
			listVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
			resp.Diagnostics.Append(diags...)
			state.IssueTypeIDs = listVal
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *IssueTypeSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IssueTypeSchemeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the scheme details
	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.DefaultIssueTypeID.IsNull() && !plan.DefaultIssueTypeID.IsUnknown() {
		body["defaultIssueTypeId"] = plan.DefaultIssueTypeID.ValueString()
	}

	err := r.client.Put(fmt.Sprintf("/rest/api/3/issuetypescheme/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating issue type scheme", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *IssueTypeSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IssueTypeSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/issuetypescheme/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting issue type scheme", err.Error())
		return
	}
}

func (r *IssueTypeSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	state := IssueTypeSchemeResourceModel{
		ID: types.StringValue(req.ID),
	}

	// Read the scheme details
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/issuetypescheme?id=%s", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing issue type scheme", err.Error())
		return
	}

	if values, ok := result["values"].([]interface{}); ok && len(values) > 0 {
		scheme := values[0].(map[string]interface{})
		state.Name = types.StringValue(fmt.Sprintf("%v", scheme["name"]))
		if desc, ok := scheme["description"].(string); ok && desc != "" {
			state.Description = types.StringValue(desc)
		}
		if ditID, ok := scheme["defaultIssueTypeId"].(string); ok && ditID != "" {
			state.DefaultIssueTypeID = types.StringValue(ditID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *IssueTypeSchemeResource) findProjectScopedIssueTypeIDs(issueTypeIDs []string, defaultIssueTypeID types.String) ([]string, error) {
	idsToValidate := make([]string, 0, len(issueTypeIDs)+1)
	seen := make(map[string]struct{}, len(issueTypeIDs)+1)

	for _, id := range issueTypeIDs {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		idsToValidate = append(idsToValidate, id)
	}

	if !defaultIssueTypeID.IsNull() && !defaultIssueTypeID.IsUnknown() {
		id := defaultIssueTypeID.ValueString()
		if _, exists := seen[id]; !exists && id != "" {
			idsToValidate = append(idsToValidate, id)
		}
	}

	var invalid []string
	for _, id := range idsToValidate {
		var issueType map[string]interface{}
		if err := r.client.Get(fmt.Sprintf("/rest/api/3/issuetype/%s", id), &issueType); err != nil {
			return nil, err
		}

		if issuetype.IsProjectScoped(issueType) {
			invalid = append(invalid, id)
		}
	}

	return invalid, nil
}
