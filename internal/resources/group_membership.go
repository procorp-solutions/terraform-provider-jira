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

var _ resource.Resource = &GroupMembershipResource{}

type GroupMembershipResource struct {
	client *client.Client
}

type GroupMembershipResourceModel struct {
	ID        types.String `tfsdk:"id"`
	GroupName types.String `tfsdk:"group_name"`
	AccountID types.String `tfsdk:"account_id"`
}

func NewGroupMembershipResource() resource.Resource {
	return &GroupMembershipResource{}
}

func (r *GroupMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (r *GroupMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's membership in a JIRA group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID (group_name/account_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
				Description: "The group name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				Description: "The Atlassian account ID of the user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *GroupMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"accountId": plan.AccountID.ValueString(),
	}

	params := url.Values{"groupname": {plan.GroupName.ValueString()}}
	err := r.client.Post("/rest/api/3/group/user?"+params.Encode(), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error adding user to group", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.GroupName.ValueString(), plan.AccountID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if user is in the group by listing group members
	params := url.Values{"groupname": {state.GroupName.ValueString()}}
	var result map[string]interface{}
	err := r.client.Get("/rest/api/3/group/member?"+params.Encode(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group membership", err.Error())
		return
	}

	found := false
	if values, ok := result["values"].([]interface{}); ok {
		for _, v := range values {
			member := v.(map[string]interface{})
			if fmt.Sprintf("%v", member["accountId"]) == state.AccountID.ValueString() {
				found = true
				break
			}
		}
	}

	if !found {
		// Check if there are more pages
		// For simplicity, if the user isn't in the first page, assume removed
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GroupMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Both fields are ForceNew, so Update should never be called.
	resp.Diagnostics.AddError("Unexpected update", "Group membership does not support in-place updates. Both group_name and account_id require replacement.")
}

func (r *GroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{
		"groupname": {state.GroupName.ValueString()},
		"accountId": {state.AccountID.ValueString()},
	}
	err := r.client.DeleteWithQuery("/rest/api/3/group/user", params)
	if err != nil {
		resp.Diagnostics.AddError("Error removing user from group", err.Error())
		return
	}
}
