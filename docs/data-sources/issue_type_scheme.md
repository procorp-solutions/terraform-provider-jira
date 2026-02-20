---
page_title: "jira_issue_type_scheme Data Source - jira"
subcategory: ""
description: |-
  Fetches an issue type scheme from JIRA.
---

# jira_issue_type_scheme (Data Source)

Fetches an issue type scheme from JIRA. Use this data source to look up existing issue type schemes by name or ID.

## Example Usage

```terraform
# Look up by name
data "jira_issue_type_scheme" "default" {
  name = "Default Issue Type Scheme"
}

# Look up by ID
data "jira_issue_type_scheme" "by_id" {
  id = "10000"
}

# Use in a project
resource "jira_project" "example" {
  key                  = "EXAM"
  name                 = "Example Project"
  project_type_key     = "software"
  lead_account_id      = data.jira_user.lead.account_id
  issue_type_scheme_id = data.jira_issue_type_scheme.default.id
}

# Output scheme details
output "default_scheme_id" {
  value = data.jira_issue_type_scheme.default.id
}
```

## Schema

### Optional

- `name` (String) The name of the issue type scheme to look up.
- `id` (String) The ID of the issue type scheme to look up.

~> **Note:** Exactly one of `name` or `id` must be specified.

### Read-Only

- `description` (String) The description of the issue type scheme.
- `default_issue_type_id` (String) The default issue type ID for this scheme.
- `issue_type_ids` (List of String) List of issue type IDs in this scheme.
