---
page_title: "jira_workflow_scheme Resource - jira"
subcategory: ""
description: |-
  Manages a workflow scheme in JIRA.
---

# jira_workflow_scheme (Resource)

Manages a workflow scheme in JIRA. Workflow schemes map issue types to workflows, defining the lifecycle of issues.

## Example Usage

```terraform
# Look up existing workflows
data "jira_workflow" "default" {
  name = "jira"
}

data "jira_workflow" "bug" {
  name = "Bug Workflow"
}

# Look up issue types
data "jira_issue_type" "bug" {
  name = "Bug"
}

data "jira_issue_type" "story" {
  name = "Story"
}

# Create a workflow scheme
resource "jira_workflow_scheme" "software" {
  name             = "Software Workflow Scheme"
  description      = "Workflow scheme for software projects"
  default_workflow = data.jira_workflow.default.name

  issue_type_mappings = {
    (data.jira_issue_type.bug.id)   = data.jira_workflow.bug.name
    (data.jira_issue_type.story.id) = data.jira_workflow.default.name
  }
}

# Use the scheme in a project
resource "jira_project" "example" {
  key                = "EXAM"
  name               = "Example Project"
  project_type_key   = "software"
  lead_account_id    = data.jira_user.lead.account_id
  workflow_scheme_id = jira_workflow_scheme.software.id
}
```

## Schema

### Required

- `name` (String) The name of the workflow scheme.
- `default_workflow` (String) The name of the default workflow for issue types not explicitly mapped.

### Optional

- `description` (String) A description of the workflow scheme.
- `issue_type_mappings` (Map of String) A map of issue type IDs to workflow names.

### Read-Only

- `id` (String) The ID of the workflow scheme.

## Import

Workflow schemes can be imported using the scheme ID:

```shell
terraform import jira_workflow_scheme.software 10001
```
