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

var _ resource.Resource = &CustomFieldResource{}
var _ resource.ResourceWithImportState = &CustomFieldResource{}

type CustomFieldResource struct {
	client *client.Client
}

type CustomFieldResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	SearchKey   types.String `tfsdk:"search_key"`
}

func NewCustomFieldResource() resource.Resource {
	return &CustomFieldResource{}
}

func (r *CustomFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (r *CustomFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA custom field.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The custom field ID (e.g. customfield_10001).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The custom field name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The custom field description.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "The custom field type (e.g. com.atlassian.jira.plugin.system.customfieldtypes:textfield, :textarea, :select, :multiselect, :float, :datepicker, :datetime, :radiobuttons, :cascadingselect, :url, :labels, :userpicker, :multiuserpicker).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"search_key": schema.StringAttribute{
				Description: "The searcher key for the custom field (e.g. com.atlassian.jira.plugin.system.customfieldtypes:textsearcher, :exacttextsearcher, :multiselectsearcher, :daterange, :numberrange).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CustomFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CustomFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"type":        plan.Type.ValueString(),
		"searcherKey": plan.SearchKey.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	var result map[string]interface{}
	err := r.client.Post("/rest/api/3/field", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating custom field", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", result["id"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// JIRA doesn't have a direct GET by ID for fields; list all and find ours
	var fields []map[string]interface{}
	err := r.client.Get("/rest/api/3/field", &fields)
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom fields", err.Error())
		return
	}

	var found bool
	for _, f := range fields {
		if fmt.Sprintf("%v", f["id"]) == state.ID.ValueString() {
			state.Name = types.StringValue(fmt.Sprintf("%v", f["name"]))
			if desc, ok := f["description"].(string); ok && desc != "" {
				state.Description = types.StringValue(desc)
			}
			if schema, ok := f["schema"].(map[string]interface{}); ok {
				if ct, ok := schema["custom"].(string); ok {
					state.Type = types.StringValue(ct)
				}
			}
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *CustomFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CustomFieldResourceModel
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

	err := r.client.Put(fmt.Sprintf("/rest/api/3/field/%s", plan.ID.ValueString()), body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating custom field", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *CustomFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CustomFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(fmt.Sprintf("/rest/api/3/field/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting custom field", err.Error())
		return
	}
}

func (r *CustomFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by field ID (e.g. customfield_10001)
	var fields []map[string]interface{}
	err := r.client.Get("/rest/api/3/field", &fields)
	if err != nil {
		resp.Diagnostics.AddError("Error importing custom field", err.Error())
		return
	}

	for _, f := range fields {
		if fmt.Sprintf("%v", f["id"]) == req.ID {
			state := CustomFieldResourceModel{
				ID:   types.StringValue(req.ID),
				Name: types.StringValue(fmt.Sprintf("%v", f["name"])),
			}
			if desc, ok := f["description"].(string); ok && desc != "" {
				state.Description = types.StringValue(desc)
			}
			if schema, ok := f["schema"].(map[string]interface{}); ok {
				if ct, ok := schema["custom"].(string); ok {
					state.Type = types.StringValue(ct)
				}
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Custom field not found", fmt.Sprintf("No custom field with ID %s found.", req.ID))
}
