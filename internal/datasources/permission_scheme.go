package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PermissionSchemeDataSource{}

type PermissionSchemeDataSource struct {
	client *client.Client
}

type PermissionSchemeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewPermissionSchemeDataSource() datasource.DataSource {
	return &PermissionSchemeDataSource{}
}

func (d *PermissionSchemeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission_scheme"
}

func (d *PermissionSchemeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a JIRA permission scheme by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The permission scheme ID. Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The permission scheme name. Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The permission scheme description.",
				Computed:    true,
			},
		},
	}
}

func (d *PermissionSchemeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PermissionSchemeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PermissionSchemeDataSourceModel
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

	var scheme map[string]interface{}

	if hasID {
		var result map[string]interface{}
		err := d.client.Get(fmt.Sprintf("/rest/api/3/permissionscheme/%s", config.ID.ValueString()), &result)
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError("Permission scheme not found",
					fmt.Sprintf("No permission scheme with id '%s' found.", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Error reading permission scheme", err.Error())
			return
		}
		scheme = result
	} else {
		var wrapper struct {
			PermissionSchemes []map[string]interface{} `json:"permissionSchemes"`
		}
		err := d.client.Get("/rest/api/3/permissionscheme", &wrapper)
		if err != nil {
			resp.Diagnostics.AddError("Error listing permission schemes", err.Error())
			return
		}
		wanted := config.Name.ValueString()
		for _, s := range wrapper.PermissionSchemes {
			name := fmt.Sprintf("%v", s["name"])
			if strings.EqualFold(name, wanted) {
				scheme = s
				break
			}
		}
		if scheme == nil {
			resp.Diagnostics.AddError("Permission scheme not found",
				fmt.Sprintf("No permission scheme with name '%s' found.", config.Name.ValueString()))
			return
		}
	}

	config.ID = types.StringValue(fmt.Sprintf("%v", scheme["id"]))
	config.Name = types.StringValue(fmt.Sprintf("%v", scheme["name"]))
	if desc, ok := scheme["description"].(string); ok && desc != "" {
		config.Description = types.StringValue(desc)
	} else {
		config.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
