---
page_title: "jira_group Data Source - jira"
subcategory: ""
description: |-
  Fetches a group from JIRA.
---

# jira_group (Data Source)

Fetches a group from JIRA. Use this data source to look up existing groups by name or ID.

## Example Usage

```terraform
# Look up by name
data "jira_group" "developers" {
  name = "jira-software-users"
}

data "jira_group" "admins" {
  name = "jira-administrators"
}

# Look up by ID
data "jira_group" "by_id" {
  id = "10000"
}

# Use in a permission scheme
resource "jira_permission_scheme" "standard" {
  name        = "Standard Permissions"
  description = "Standard permission scheme"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = data.jira_group.developers.name
    },
    {
      permission       = "ADMINISTER_PROJECTS"
      holder_type      = "group"
      holder_parameter = data.jira_group.admins.name
    },
  ]
}

# Output group details
output "developers_group_id" {
  value = data.jira_group.developers.id
}
```

## Schema

### Optional

- `name` (String) The name of the group to look up.
- `id` (String) The ID of the group to look up.

~> **Note:** Exactly one of `name` or `id` must be specified.

### Read-Only

- `group_id` (String) The group ID (same as `id` when looked up by name).
