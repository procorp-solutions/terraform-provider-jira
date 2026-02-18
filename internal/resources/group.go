package resources

import (
	"context"
	"fmt"
	"net/url"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

type GroupResource struct {
	client *client.Client
}

type GroupResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA user group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The group ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The group name.",
				Required:    true,
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/group", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	if groupId, ok := result["groupId"].(string); ok {
		plan.ID = types.StringValue(groupId)
	} else {
		// Fallback: use name as ID
		plan.ID = types.StringValue(plan.Name.ValueString())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use bulk get to find the group
	var result map[string]interface{}
	params := url.Values{"groupName": {state.Name.ValueString()}}
	err := r.client.Get("/rest/api/3/group/bulk?"+params.Encode(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group", err.Error())
		return
	}

	if values, ok := result["values"].([]interface{}); ok && len(values) > 0 {
		group := values[0].(map[string]interface{})
		state.Name = types.StringValue(fmt.Sprintf("%v", group["name"]))
		if groupId, ok := group["groupId"].(string); ok {
			state.ID = types.StringValue(groupId)
		}
	} else {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GroupResourceModel
	var state GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// JIRA doesn't support renaming groups directly.
	// We need to delete the old group and create a new one.
	params := url.Values{"groupname": {state.Name.ValueString()}}
	err := r.client.DeleteWithQuery("/rest/api/3/group", params)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting old group during rename", err.Error())
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	var result map[string]interface{}
	err = r.client.Post("/rest/api/3/group", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating new group during rename", err.Error())
		return
	}

	if groupId, ok := result["groupId"].(string); ok {
		plan.ID = types.StringValue(groupId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{"groupname": {state.Name.ValueString()}}
	err := r.client.DeleteWithQuery("/rest/api/3/group", params)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting group", err.Error())
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by group name
	var result map[string]interface{}
	params := url.Values{"groupName": {req.ID}}
	err := r.client.Get("/rest/api/3/group/bulk?"+params.Encode(), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing group", err.Error())
		return
	}

	if values, ok := result["values"].([]interface{}); ok && len(values) > 0 {
		group := values[0].(map[string]interface{})
		state := GroupResourceModel{
			Name: types.StringValue(fmt.Sprintf("%v", group["name"])),
		}
		if groupId, ok := group["groupId"].(string); ok {
			state.ID = types.StringValue(groupId)
		} else {
			state.ID = types.StringValue(req.ID)
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	} else {
		resp.Diagnostics.AddError("Group not found", fmt.Sprintf("No group with name %s found.", req.ID))
	}
}
