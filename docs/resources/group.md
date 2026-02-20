---
page_title: "jira_group Resource - jira"
subcategory: ""
description: |-
  Manages a user group in JIRA.
---

# jira_group (Resource)

Manages a user group in JIRA. Groups are used to organize users and assign permissions.

## Example Usage

```terraform
# Create a developers group
resource "jira_group" "developers" {
  name = "developers"
}

# Create an administrators group
resource "jira_group" "admins" {
  name = "project-administrators"
}

# Create a QA team group
resource "jira_group" "qa" {
  name = "qa-team"
}

# Use the group in a permission scheme
resource "jira_permission_scheme" "standard" {
  name = "Standard Permissions"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = jira_group.developers.name
    },
    {
      permission       = "ADMINISTER_PROJECTS"
      holder_type      = "group"
      holder_parameter = jira_group.admins.name
    },
  ]
}
```

## Schema

### Required

- `name` (String) The name of the group.

### Read-Only

- `id` (String) The ID of the group.

## Import

Groups can be imported using the group name:

```shell
terraform import jira_group.developers developers
```
