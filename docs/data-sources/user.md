---
page_title: "jira_user Data Source - jira"
subcategory: ""
description: |-
  Fetches a user from JIRA.
---

# jira_user (Data Source)

Fetches a user from JIRA. Use this data source to look up users by email address or account ID.

~> **Note:** Users are managed through Atlassian Admin, not the JIRA REST API. This is a read-only data source.

## Example Usage

```terraform
# Look up a user by email
data "jira_user" "lead" {
  email_address = "lead@example.com"
}

# Look up a user by account ID
data "jira_user" "specific" {
  account_id = "5b10a2844c20165700ede21g"
}

# Use the user as a project lead
resource "jira_project" "example" {
  key              = "EXAM"
  name             = "Example Project"
  project_type_key = "software"
  lead_account_id  = data.jira_user.lead.account_id
}

# Output user information
output "lead_display_name" {
  value = data.jira_user.lead.display_name
}
```

## Schema

### Optional

- `email_address` (String) The email address of the user to look up.
- `account_id` (String) The account ID of the user to look up.

~> **Note:** Exactly one of `email_address` or `account_id` must be specified.

### Read-Only

- `display_name` (String) The display name of the user.
- `active` (Boolean) Whether the user is active.
