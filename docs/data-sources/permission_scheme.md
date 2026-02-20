---
page_title: "jira_permission_scheme Data Source - jira"
subcategory: ""
description: |-
  Fetches a permission scheme from JIRA.
---

# jira_permission_scheme (Data Source)

Fetches a permission scheme from JIRA. Use this data source to look up existing permission schemes by name or ID.

## Example Usage

```terraform
# Look up by name
data "jira_permission_scheme" "default" {
  name = "Default Permission Scheme"
}

# Look up by ID
data "jira_permission_scheme" "by_id" {
  id = "10000"
}

# Use in a project
resource "jira_project" "example" {
  key                  = "EXAM"
  name                 = "Example Project"
  project_type_key     = "software"
  lead_account_id      = data.jira_user.lead.account_id
  permission_scheme_id = data.jira_permission_scheme.default.id
}

# Output scheme details
output "default_permission_scheme_id" {
  value = data.jira_permission_scheme.default.id
}
```

## Schema

### Optional

- `name` (String) The name of the permission scheme to look up.
- `id` (String) The ID of the permission scheme to look up.

~> **Note:** Exactly one of `name` or `id` must be specified.

### Read-Only

- `description` (String) The description of the permission scheme.
