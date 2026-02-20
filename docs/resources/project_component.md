---
page_title: "jira_project_component Resource - jira"
subcategory: ""
description: |-
  Manages a project component in JIRA.
---

# jira_project_component (Resource)

Manages a project component in JIRA. Components are subsections of a project used to group issues.

## Example Usage

```terraform
data "jira_user" "lead" {
  email_address = "lead@example.com"
}

resource "jira_project" "example" {
  key              = "EXAM"
  name             = "Example Project"
  project_type_key = "software"
  lead_account_id  = data.jira_user.lead.account_id
}

# Create a component with a lead
resource "jira_project_component" "frontend" {
  project_key     = jira_project.example.key
  name            = "Frontend"
  description     = "Frontend components and UI"
  lead_account_id = data.jira_user.lead.account_id
  assignee_type   = "COMPONENT_LEAD"
}

# Create a simple component
resource "jira_project_component" "backend" {
  project_key = jira_project.example.key
  name        = "Backend"
  description = "Backend services and APIs"
}
```

## Schema

### Required

- `project_key` (String) The key of the project this component belongs to.
- `name` (String) The name of the component.

### Optional

- `description` (String) A description of the component.
- `lead_account_id` (String) The account ID of the component lead.
- `assignee_type` (String) The default assignee type. Valid values: `PROJECT_DEFAULT`, `COMPONENT_LEAD`, `PROJECT_LEAD`, `UNASSIGNED`.

### Read-Only

- `id` (String) The ID of the component.

## Import

Components can be imported using the component ID:

```shell
terraform import jira_project_component.frontend 10001
```
