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

var _ datasource.DataSource = &GroupDataSource{}

type GroupDataSource struct {
	client *client.Client
}

type GroupDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

func (d *GroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a JIRA group by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The group ID (groupId). Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The group name. Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (d *GroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing input", "Either id or name must be specified.")
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError("Ambiguous input", "Specify only one of id or name.")
		return
	}

	params := url.Values{}
	if hasID {
		params.Set("groupId", config.ID.ValueString())
	} else {
		params.Set("groupName", config.Name.ValueString())
	}

	var result map[string]interface{}
	err := d.client.Get("/rest/api/3/group/bulk?"+params.Encode(), &result)
	if err != nil {
		if client.IsNotFound(err) {
			if hasID {
				resp.Diagnostics.AddError("Group not found", fmt.Sprintf("No group with id '%s' found.", config.ID.ValueString()))
			} else {
				resp.Diagnostics.AddError("Group not found", fmt.Sprintf("No group with name '%s' found.", config.Name.ValueString()))
			}
			return
		}
		resp.Diagnostics.AddError("Error reading group", err.Error())
		return
	}

	values, ok := result["values"].([]interface{})
	if !ok || len(values) == 0 {
		if hasID {
			resp.Diagnostics.AddError("Group not found", fmt.Sprintf("No group with id '%s' found.", config.ID.ValueString()))
		} else {
			resp.Diagnostics.AddError("Group not found", fmt.Sprintf("No group with name '%s' found.", config.Name.ValueString()))
		}
		return
	}

	group := values[0].(map[string]interface{})
	config.Name = types.StringValue(fmt.Sprintf("%v", group["name"]))
	if groupId, ok := group["groupId"].(string); ok && groupId != "" {
		config.ID = types.StringValue(groupId)
	} else {
		config.ID = types.StringValue(fmt.Sprintf("%v", group["groupId"]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
