package datasources

import (
	"context"
	"fmt"
	"net/url"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WorkflowDataSource{}

type WorkflowDataSource struct {
	client *client.Client
}

type WorkflowDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Steps       types.Int64  `tfsdk:"steps"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func NewWorkflowDataSource() datasource.DataSource {
	return &WorkflowDataSource{}
}

func (d *WorkflowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *WorkflowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a JIRA workflow by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The workflow name to search for.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The workflow description.",
				Computed:    true,
			},
			"steps": schema.Int64Attribute{
				Description: "The number of steps (statuses) in the workflow.",
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default workflow.",
				Computed:    true,
			},
		},
	}
}

func (d *WorkflowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.Client.")
		return
	}
	d.client = c
}

func (d *WorkflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WorkflowDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := url.Values{"workflowName": {config.Name.ValueString()}}
	var result map[string]interface{}
	err := d.client.Get("/rest/api/3/workflow/search?"+params.Encode(), &result)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow", err.Error())
		return
	}

	if values, ok := result["values"].([]interface{}); ok && len(values) > 0 {
		wf := values[0].(map[string]interface{})

		config.Name = types.StringValue(fmt.Sprintf("%v", wf["id"].(map[string]interface{})["name"]))
		if desc, ok := wf["description"].(string); ok {
			config.Description = types.StringValue(desc)
		} else {
			config.Description = types.StringValue("")
		}
		if statuses, ok := wf["statuses"].([]interface{}); ok {
			config.Steps = types.Int64Value(int64(len(statuses)))
		} else {
			config.Steps = types.Int64Value(0)
		}
		if isDefault, ok := wf["isDefault"].(bool); ok {
			config.IsDefault = types.BoolValue(isDefault)
		} else {
			config.IsDefault = types.BoolValue(false)
		}
	} else {
		resp.Diagnostics.AddError("Workflow not found",
			fmt.Sprintf("No workflow with name '%s' found.", config.Name.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
