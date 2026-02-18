package resources

import (
	"context"
	"fmt"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PermissionSchemeResource{}
var _ resource.ResourceWithImportState = &PermissionSchemeResource{}

type PermissionSchemeResource struct {
	client *client.Client
}

type PermissionSchemeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
}

var permissionGrantObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"permission":       types.StringType,
		"holder_type":      types.StringType,
		"holder_parameter": types.StringType,
	},
}

func NewPermissionSchemeResource() resource.Resource {
	return &PermissionSchemeResource{}
}

func (r *PermissionSchemeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_scheme"
}

func (r *PermissionSchemeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA permission scheme.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The permission scheme ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The permission scheme name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The permission scheme description.",
				Optional:    true,
			},
			"permissions": schema.ListNestedAttribute{
				Description: "List of permission grants.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							Description: "Permission key (e.g. BROWSE_PROJECTS, CREATE_ISSUES, ADMINISTER_PROJECTS).",
							Required:    true,
						},
						"holder_type": schema.StringAttribute{
							Description: "Holder type: group, projectRole, user, anyone, applicationRole, etc.",
							Required:    true,
						},
						"holder_parameter": schema.StringAttribute{
							Description: "Holder parameter: group name, role ID, account ID, etc. Leave empty for 'anyone'.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *PermissionSchemeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PermissionSchemeResource) buildPermissions(ctx context.Context, plan PermissionSchemeResourceModel) ([]map[string]interface{}, error) {
	if plan.Permissions.IsNull() || plan.Permissions.IsUnknown() {
		return nil, nil
	}

	var perms []struct {
		Permission      string `tfsdk:"permission"`
		HolderType      string `tfsdk:"holder_type"`
		HolderParameter string `tfsdk:"holder_parameter"`
	}

	diags := plan.Permissions.ElementsAs(ctx, &perms, false)
	if diags.HasError() {
		return nil, fmt.Errorf("error reading permissions")
	}

	var result []map[string]interface{}
	for _, p := range perms {
		grant := map[string]interface{}{
			"permission": p.Permission,
			"holder": map[string]interface{}{
				"type":      p.HolderType,
				"parameter": p.HolderParameter,
			},
		}
		result = append(result, grant)
	}
	return result, nil
}

func (r *PermissionSchemeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PermissionSchemeResourceModel
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

	perms, err := r.buildPermissions(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building permissions", err.Error())
		return
	}
	if perms != nil {
		body["permissions"] = perms
	}

	var result map[string]interface{}
	err = r.client.Post("/rest/api/3/permissionscheme", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating permission scheme", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PermissionSchemeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/permissionscheme/%s?expand=permissions", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading permission scheme", err.Error())
		return
	}

	state.Name = types.StringValue(fmt.Sprintf("%v", result["name"]))
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}

	// Parse permissions from response
	if permsRaw, ok := result["permissions"].([]interface{}); ok && len(permsRaw) > 0 {
		var permValues []attr.Value
		for _, p := range permsRaw {
			pMap := p.(map[string]interface{})
			permission := fmt.Sprintf("%v", pMap["permission"])

			holderType := ""
			holderParam := ""
			if holder, ok := pMap["holder"].(map[string]interface{}); ok {
				holderType = fmt.Sprintf("%v", holder["type"])
				if param, ok := holder["parameter"]; ok && param != nil {
					holderParam = fmt.Sprintf("%v", param)
				}
			}

			objVal, diags := types.ObjectValue(
				permissionGrantObjectType.AttrTypes,
				map[string]attr.Value{
					"permission":       types.StringValue(permission),
					"holder_type":      types.StringValue(holderType),
					"holder_parameter": types.StringValue(holderParam),
				},
			)
			resp.Diagnostics.Append(diags...)
			permValues = append(permValues, objVal)
		}

		listVal, diags := types.ListValue(permissionGrantObjectType, permValues)
		resp.Diagnostics.Append(diags...)
		state.Permissions = listVal
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PermissionSchemeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PermissionSchemeResourceModel
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

	perms, err := r.buildPermissions(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error building permissions", err.Error())
		return
	}
	if perms != nil {
		body["permissions"] = perms
	}

	err = r.client.Put(fmt.Sprintf("/rest/api/3/permissionscheme/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating permission scheme", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PermissionSchemeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PermissionSchemeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/permissionscheme/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting permission scheme", err.Error())
		return
	}
}

func (r *PermissionSchemeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/api/3/permissionscheme/%s?expand=permissions", req.ID), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error importing permission scheme", err.Error())
		return
	}

	state := PermissionSchemeResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%v", result["id"])),
		Name: types.StringValue(fmt.Sprintf("%v", result["name"])),
	}
	if desc, ok := result["description"].(string); ok && desc != "" {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
