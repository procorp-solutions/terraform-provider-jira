package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/david/terraform-provider-jira/internal/issuetype"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IssueTypeDataSource{}

type IssueTypeDataSource struct {
	client *client.Client
}

type IssueTypeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func NewIssueTypeDataSource() datasource.DataSource {
	return &IssueTypeDataSource{}
}

func (d *IssueTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue_type"
}

func (d *IssueTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a JIRA issue type by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The issue type ID. Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The issue type name. Provide either id or name.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The issue type description.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The hierarchy level: standard or subtask.",
				Computed:    true,
			},
		},
	}
}

func (d *IssueTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IssueTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IssueTypeDataSourceModel
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

	var issueType map[string]interface{}

	if hasID {
		err := d.client.Get(fmt.Sprintf("/rest/api/3/issuetype/%s", config.ID.ValueString()), &issueType)
		if err != nil {
			if client.IsNotFound(err) {
				resp.Diagnostics.AddError("Issue type not found",
					fmt.Sprintf("No issue type with id '%s' found.", config.ID.ValueString()))
				return
			}
			resp.Diagnostics.AddError("Error reading issue type", err.Error())
			return
		}
	} else {
		var issueTypes []map[string]interface{}
		err := d.client.Get("/rest/api/3/issuetype", &issueTypes)
		if err != nil {
			resp.Diagnostics.AddError("Error listing issue types", err.Error())
			return
		}

		var matches []map[string]interface{}
		for _, it := range issueTypes {
			name := fmt.Sprintf("%v", it["name"])
			if strings.EqualFold(name, config.Name.ValueString()) {
				matches = append(matches, it)
			}
		}
		if len(matches) == 0 {
			resp.Diagnostics.AddError("Issue type not found",
				fmt.Sprintf("No issue type with name '%s' found.", config.Name.ValueString()))
			return
		}
		// Select only issue types without scope (global/classic); skip project-scoped (next-gen)
		issueType = d.selectGlobalIssueType(matches)
		if issueType == nil {
			resp.Diagnostics.AddError("Issue type not found",
				fmt.Sprintf("No global (classic) issue type with name '%s' found; only project-scoped (next-gen) types exist.", config.Name.ValueString()))
			return
		}
	}

	config.ID = types.StringValue(fmt.Sprintf("%v", issueType["id"]))
	config.Name = types.StringValue(fmt.Sprintf("%v", issueType["name"]))

	if desc, ok := issueType["description"].(string); ok && desc != "" {
		config.Description = types.StringValue(desc)
	} else {
		config.Description = types.StringValue("")
	}

	if subtask, ok := issueType["subtask"].(bool); ok && subtask {
		config.Type = types.StringValue("subtask")
	} else {
		config.Type = types.StringValue("standard")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// selectGlobalIssueType returns a global (classic) issue type from the name matches.
// Global types have no "scope" field in the API response; project-scoped (next-gen) ones do.
// We check both the list item and the single-item GET — if either has "scope", we skip it.
func (d *IssueTypeDataSource) selectGlobalIssueType(matches []map[string]interface{}) map[string]interface{} {
	for _, it := range matches {
		// Skip if the list response already contains scope
		if issuetype.IsProjectScoped(it) {
			continue
		}
		// Fetch full details by ID; the single-item GET may include scope the list omitted
		id := fmt.Sprintf("%v", it["id"])
		var full map[string]interface{}
		if err := d.client.Get(fmt.Sprintf("/rest/api/3/issuetype/%s", id), &full); err != nil {
			continue
		}
		if !issuetype.IsProjectScoped(full) {
			return full
		}
	}
	return nil
}

