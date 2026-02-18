package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &AutomationRuleResource{}

type AutomationRuleResource struct {
	client *client.Client
}

type AutomationRuleResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	State    types.String `tfsdk:"state"`
	RuleJSON types.String `tfsdk:"rule_json"`
}

func NewAutomationRuleResource() resource.Resource {
	return &AutomationRuleResource{}
}

func (r *AutomationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_automation_rule"
}

func (r *AutomationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JIRA automation rule. Note: JIRA Cloud does not support deleting automation rules via API. On destroy, the rule will be disabled instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The automation rule UUID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The automation rule name.",
				Required:    true,
			},
			"state": schema.StringAttribute{
				Description: "The rule state: ENABLED or DISABLED.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ENABLED"),
			},
			"rule_json": schema.StringAttribute{
				Description: "The full rule definition as a JSON string. Tip: create a rule in the JIRA UI, retrieve it via the API, then use its JSON as a template.",
				Required:    true,
			},
		},
	}
}

func (r *AutomationRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AutomationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AutomationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the rule JSON and set the name
	var ruleBody map[string]interface{}
	if err := json.Unmarshal([]byte(plan.RuleJSON.ValueString()), &ruleBody); err != nil {
		resp.Diagnostics.AddError("Error parsing rule_json", err.Error())
		return
	}
	ruleBody["name"] = plan.Name.ValueString()

	var result map[string]interface{}
	err := r.client.Post("/rest/v1/rule", ruleBody, &result)
	if err != nil {
		resp.Diagnostics.AddError("Error creating automation rule", err.Error())
		return
	}

	if id, ok := result["id"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", id))
	} else if ruleUuid, ok := result["ruleUuid"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", ruleUuid))
	}

	// Set the desired state
	if plan.State.ValueString() == "ENABLED" {
		stateBody := map[string]interface{}{"state": "ENABLED"}
		_ = r.client.Put(fmt.Sprintf("/rest/v1/rule/%s/state", plan.ID.ValueString()), stateBody, nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AutomationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AutomationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	err := r.client.Get(fmt.Sprintf("/rest/v1/rule/%s", state.ID.ValueString()), &result)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading automation rule", err.Error())
		return
	}

	if name, ok := result["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if ruleState, ok := result["state"].(string); ok {
		state.State = types.StringValue(ruleState)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *AutomationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AutomationRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the rule definition
	var ruleBody map[string]interface{}
	if err := json.Unmarshal([]byte(plan.RuleJSON.ValueString()), &ruleBody); err != nil {
		resp.Diagnostics.AddError("Error parsing rule_json", err.Error())
		return
	}
	ruleBody["name"] = plan.Name.ValueString()

	err := r.client.Put(fmt.Sprintf("/rest/v1/rule/%s", plan.ID.ValueString()), ruleBody, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating automation rule", err.Error())
		return
	}

	// Update state if needed
	stateBody := map[string]interface{}{"state": plan.State.ValueString()}
	err = r.client.Put(fmt.Sprintf("/rest/v1/rule/%s/state", plan.ID.ValueString()), stateBody, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating automation rule state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *AutomationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AutomationRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// JIRA Cloud does not support deleting automation rules via API.
	// Instead, disable the rule.
	tflog.Warn(ctx, "JIRA Cloud does not support deleting automation rules via API. Disabling rule instead.",
		map[string]interface{}{"rule_id": state.ID.ValueString()})

	stateBody := map[string]interface{}{"state": "DISABLED"}
	err := r.client.Put(fmt.Sprintf("/rest/v1/rule/%s/state", state.ID.ValueString()), stateBody, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error disabling automation rule", err.Error())
		return
	}
}
