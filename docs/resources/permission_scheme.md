---
page_title: "jira_permission_scheme Resource - jira"
subcategory: ""
description: |-
  Manages a permission scheme in JIRA.
---

# jira_permission_scheme (Resource)

Manages a permission scheme in JIRA. Permission schemes control who can do what in a project.

## Example Usage

```terraform
# Look up existing groups
data "jira_group" "developers" {
  name = "jira-software-users"
}

data "jira_group" "admins" {
  name = "jira-administrators"
}

# Create a permission scheme
resource "jira_permission_scheme" "standard" {
  name        = "Standard Permission Scheme"
  description = "Standard permissions for development projects"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = data.jira_group.developers.name
    },
    {
      permission       = "CREATE_ISSUES"
      holder_type      = "group"
      holder_parameter = data.jira_group.developers.name
    },
    {
      permission       = "EDIT_ISSUES"
      holder_type      = "group"
      holder_parameter = data.jira_group.developers.name
    },
    {
      permission       = "ADMINISTER_PROJECTS"
      holder_type      = "group"
      holder_parameter = data.jira_group.admins.name
    },
    {
      permission       = "ASSIGN_ISSUES"
      holder_type      = "projectRole"
      holder_parameter = "10002"
    },
  ]
}

# Use the scheme in a project
resource "jira_project" "example" {
  key                  = "EXAM"
  name                 = "Example Project"
  project_type_key     = "software"
  lead_account_id      = data.jira_user.lead.account_id
  permission_scheme_id = jira_permission_scheme.standard.id
}
```

## Schema

### Required

- `name` (String) The name of the permission scheme.

### Optional

- `description` (String) A description of the permission scheme.
- `permissions` (List of Object) List of permission grants. Each grant has:
  - `permission` (String) The permission key (e.g., `BROWSE_PROJECTS`, `CREATE_ISSUES`, `EDIT_ISSUES`, `ADMINISTER_PROJECTS`).
  - `holder_type` (String) The type of holder. Valid values: `group`, `projectRole`, `user`, `applicationRole`.
  - `holder_parameter` (String) The holder identifier (group name, role ID, user account ID, etc.).

### Read-Only

- `id` (String) The ID of the permission scheme.

## Common Permissions

| Permission | Description |
|------------|-------------|
| `BROWSE_PROJECTS` | View the project and its issues |
| `CREATE_ISSUES` | Create issues in the project |
| `EDIT_ISSUES` | Edit issues |
| `ASSIGN_ISSUES` | Assign issues to users |
| `RESOLVE_ISSUES` | Resolve and reopen issues |
| `CLOSE_ISSUES` | Close issues |
| `DELETE_ISSUES` | Delete issues |
| `ADMINISTER_PROJECTS` | Administer the project |

## Import

Permission schemes can be imported using the scheme ID:

```shell
terraform import jira_permission_scheme.standard 10001
```
