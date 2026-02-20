---
page_title: "jira_group_membership Resource - jira"
subcategory: ""
description: |-
  Manages a user's membership in a JIRA group.
---

# jira_group_membership (Resource)

Manages a user's membership in a JIRA group. Use this resource to add users to groups.

## Example Usage

```terraform
# Look up an existing user
data "jira_user" "developer" {
  email_address = "developer@example.com"
}

data "jira_user" "lead" {
  email_address = "lead@example.com"
}

# Create a group
resource "jira_group" "developers" {
  name = "developers"
}

resource "jira_group" "admins" {
  name = "project-administrators"
}

# Add users to groups
resource "jira_group_membership" "developer_member" {
  group_name = jira_group.developers.name
  account_id = data.jira_user.developer.account_id
}

resource "jira_group_membership" "lead_developer" {
  group_name = jira_group.developers.name
  account_id = data.jira_user.lead.account_id
}

resource "jira_group_membership" "lead_admin" {
  group_name = jira_group.admins.name
  account_id = data.jira_user.lead.account_id
}
```

## Schema

### Required

- `group_name` (String) The name of the group.
- `account_id` (String) The account ID of the user to add to the group.

### Read-Only

- `id` (String) The ID of the membership (composite of group name and account ID).

## Import

Group memberships can be imported using the format `group_name/account_id`:

```shell
terraform import jira_group_membership.developer_member developers/5b10a2844c20165700ede21g
```
