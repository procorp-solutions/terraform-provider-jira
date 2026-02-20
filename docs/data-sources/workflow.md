---
page_title: "jira_workflow Data Source - jira"
subcategory: ""
description: |-
  Fetches a workflow from JIRA.
---

# jira_workflow (Data Source)

Fetches a workflow from JIRA. Use this data source to look up existing workflows by name for use in workflow schemes.

~> **Note:** Workflows are complex graph structures. This is a read-only data source. Use workflow schemes to assign workflows to projects.

## Example Usage

```terraform
# Look up the default Jira workflow
data "jira_workflow" "default" {
  name = "jira"
}

# Look up a custom workflow
data "jira_workflow" "bug_workflow" {
  name = "Bug Workflow"
}

# Use workflows in a workflow scheme
resource "jira_workflow_scheme" "software" {
  name             = "Software Workflow Scheme"
  description      = "Workflow scheme for software projects"
  default_workflow = data.jira_workflow.default.name

  issue_type_mappings = {
    (data.jira_issue_type.bug.id) = data.jira_workflow.bug_workflow.name
  }
}

# Output workflow details
output "default_workflow_name" {
  value = data.jira_workflow.default.name
}
```

## Schema

### Required

- `name` (String) The name of the workflow to look up.

### Read-Only

- `id` (String) The ID of the workflow.
- `description` (String) The description of the workflow.
