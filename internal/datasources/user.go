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

var _ datasource.DataSource = &UserDataSource{}

type UserDataSource struct {
	client *client.Client
}

type UserDataSourceModel struct {
	AccountID    types.String `tfsdk:"account_id"`
	EmailAddress types.String `tfsdk:"email_address"`
	DisplayName  types.String `tfsdk:"display_name"`
	Active       types.Bool   `tfsdk:"active"`
	TimeZone     types.String `tfsdk:"timezone"`
}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a JIRA user by account ID or email address.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Description: "The user's Atlassian account ID. Provide either account_id or email_address.",
				Optional:    true,
				Computed:    true,
			},
			"email_address": schema.StringAttribute{
				Description: "The user's email address. Provide either account_id or email_address.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The user's display name.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the user account is active.",
				Computed:    true,
			},
			"timezone": schema.StringAttribute{
				Description: "The user's timezone.",
				Computed:    true,
			},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user map[string]interface{}

	if !config.AccountID.IsNull() && !config.AccountID.IsUnknown() && config.AccountID.ValueString() != "" {
		// Look up by account ID
		params := url.Values{"accountId": {config.AccountID.ValueString()}}
		err := d.client.Get("/rest/api/3/user?"+params.Encode(), &user)
		if err != nil {
			resp.Diagnostics.AddError("Error reading user by account ID", err.Error())
			return
		}
	} else if !config.EmailAddress.IsNull() && !config.EmailAddress.IsUnknown() && config.EmailAddress.ValueString() != "" {
		// Search by email
		params := url.Values{"query": {config.EmailAddress.ValueString()}}
		var users []map[string]interface{}
		err := d.client.Get("/rest/api/3/user/search?"+params.Encode(), &users)
		if err != nil {
			resp.Diagnostics.AddError("Error searching for user by email", err.Error())
			return
		}
		if len(users) == 0 {
			resp.Diagnostics.AddError("User not found",
				fmt.Sprintf("No user found with email '%s'.", config.EmailAddress.ValueString()))
			return
		}
		user = users[0]
	} else {
		resp.Diagnostics.AddError("Missing input",
			"Either account_id or email_address must be specified.")
		return
	}

	config.AccountID = types.StringValue(fmt.Sprintf("%v", user["accountId"]))
	if email, ok := user["emailAddress"].(string); ok && email != "" {
		config.EmailAddress = types.StringValue(email)
	} else {
		config.EmailAddress = types.StringValue("")
	}
	config.DisplayName = types.StringValue(fmt.Sprintf("%v", user["displayName"]))
	if active, ok := user["active"].(bool); ok {
		config.Active = types.BoolValue(active)
	}
	if tz, ok := user["timeZone"].(string); ok {
		config.TimeZone = types.StringValue(tz)
	} else {
		config.TimeZone = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
