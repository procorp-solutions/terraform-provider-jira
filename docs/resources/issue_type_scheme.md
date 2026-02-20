---
page_title: "jira_issue_type_scheme Resource - jira"
subcategory: ""
description: |-
  Manages an issue type scheme in JIRA.
---

# jira_issue_type_scheme (Resource)

Manages an issue type scheme in JIRA. Issue type schemes define which issue types are available in a project.

## Example Usage

```terraform
# Look up existing issue types
data "jira_issue_type" "story" {
  name = "Story"
}

data "jira_issue_type" "bug" {
  name = "Bug"
}

data "jira_issue_type" "task" {
  name = "Task"
}

data "jira_issue_type" "epic" {
  name = "Epic"
}

data "jira_issue_type" "subtask" {
  name = "Sub-task"
}

# Create an issue type scheme
resource "jira_issue_type_scheme" "software" {
  name                  = "Software Development Scheme"
  description           = "Issue types for software projects"
  default_issue_type_id = data.jira_issue_type.task.id

  issue_type_ids = [
    data.jira_issue_type.epic.id,
    data.jira_issue_type.story.id,
    data.jira_issue_type.bug.id,
    data.jira_issue_type.task.id,
    data.jira_issue_type.subtask.id,
  ]
}

# Use the scheme in a project
resource "jira_project" "example" {
  key                  = "EXAM"
  name                 = "Example Project"
  project_type_key     = "software"
  lead_account_id      = data.jira_user.lead.account_id
  issue_type_scheme_id = jira_issue_type_scheme.software.id
}
```

## Schema

### Required

- `name` (String) The name of the issue type scheme.
- `issue_type_ids` (List of String) List of issue type IDs included in this scheme.

### Optional

- `description` (String) A description of the issue type scheme.
- `default_issue_type_id` (String) The ID of the default issue type for this scheme.

### Read-Only

- `id` (String) The ID of the issue type scheme.

## Import

Issue type schemes can be imported using the scheme ID:

```shell
terraform import jira_issue_type_scheme.software 10001
```
