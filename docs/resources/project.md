---
page_title: "jira_project Resource - jira"
subcategory: ""
description: |-
  Manages a JIRA project.
---

# jira_project (Resource)

Manages a JIRA project. Projects are the primary container for issues in JIRA.

## Example Usage

```terraform
# Look up an existing user to be the project lead
data "jira_user" "lead" {
  email_address = "lead@example.com"
}

# Create a basic software project
resource "jira_project" "basic" {
  key              = "BASIC"
  name             = "Basic Project"
  description      = "A basic software project"
  project_type_key = "software"
  lead_account_id  = data.jira_user.lead.account_id
}

# Create a project with custom schemes
resource "jira_project" "advanced" {
  key              = "ADV"
  name             = "Advanced Project"
  description      = "Project with custom schemes"
  project_type_key = "software"
  lead_account_id  = data.jira_user.lead.account_id
  assignee_type    = "PROJECT_LEAD"

  issue_type_scheme_id = jira_issue_type_scheme.custom.id
  permission_scheme_id = jira_permission_scheme.custom.id
  workflow_scheme_id   = jira_workflow_scheme.custom.id
}
```

## Schema

### Required

- `key` (String) The project key. Must be unique and contain only uppercase letters (2-10 characters).
- `name` (String) The name of the project.
- `project_type_key` (String) The project type. Valid values: `software`, `service_desk`, `business`.
- `lead_account_id` (String) The account ID of the project lead.

### Optional

- `description` (String) A description of the project.
- `assignee_type` (String) The default assignee type. Valid values: `PROJECT_LEAD`, `UNASSIGNED`.
- `issue_type_scheme_id` (String) The ID of the issue type scheme to use.
- `permission_scheme_id` (String) The ID of the permission scheme to use.
- `workflow_scheme_id` (String) The ID of the workflow scheme to use.

### Read-Only

- `id` (String) The ID of the project.

## Import

Projects can be imported using the project key:

```shell
terraform import jira_project.example EXAM
```
