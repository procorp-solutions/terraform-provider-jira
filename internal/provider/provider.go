package provider

import (
	"context"
	"os"

	"github.com/david/terraform-provider-jira/internal/client"
	"github.com/david/terraform-provider-jira/internal/datasources"
	"github.com/david/terraform-provider-jira/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &JiraProvider{}

// JiraProvider defines the JIRA Terraform provider.
type JiraProvider struct {
	version string
}

// JiraProviderModel maps the provider schema to Go types.
type JiraProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Email    types.String `tfsdk:"email"`
	APIToken types.String `tfsdk:"api_token"`
}

// New returns a new provider instance.
func New() provider.Provider {
	return &JiraProvider{
		version: "0.1.0",
	}
}

func (p *JiraProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jira"
	resp.Version = p.version
}

func (p *JiraProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing JIRA Cloud resources.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "JIRA Cloud instance URL (e.g. https://myorg.atlassian.net). Can also be set via JIRA_URL environment variable.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Atlassian account email. Can also be set via JIRA_EMAIL environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "Atlassian API token. Can also be set via JIRA_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *JiraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config JiraProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve URL
	jiraURL := os.Getenv("JIRA_URL")
	if !config.URL.IsNull() {
		jiraURL = config.URL.ValueString()
	}
	if jiraURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing JIRA URL",
			"The provider cannot create the JIRA API client because the URL is missing. "+
				"Set the url attribute in the provider configuration or the JIRA_URL environment variable.",
		)
	}

	// Resolve Email
	email := os.Getenv("JIRA_EMAIL")
	if !config.Email.IsNull() {
		email = config.Email.ValueString()
	}
	if email == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Missing JIRA Email",
			"The provider cannot create the JIRA API client because the email is missing. "+
				"Set the email attribute in the provider configuration or the JIRA_EMAIL environment variable.",
		)
	}

	// Resolve API Token
	apiToken := os.Getenv("JIRA_API_TOKEN")
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing JIRA API Token",
			"The provider cannot create the JIRA API client because the API token is missing. "+
				"Set the api_token attribute in the provider configuration or the JIRA_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(jiraURL, email, apiToken)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *JiraProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewProjectResource,
		resources.NewProjectComponentResource,
		resources.NewWorkflowSchemeResource,
		resources.NewPermissionSchemeResource,
		resources.NewIssueTypeResource,
		resources.NewIssueTypeSchemeResource,
		resources.NewCustomFieldResource,
		resources.NewAutomationRuleResource,
		resources.NewGroupResource,
		resources.NewGroupMembershipResource,
	}
}

func (p *JiraProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewWorkflowDataSource,
		datasources.NewUserDataSource,
		datasources.NewIssueTypeDataSource,
		datasources.NewPermissionSchemeDataSource,
		datasources.NewIssueTypeSchemeDataSource,
		datasources.NewGroupDataSource,
	}
}
