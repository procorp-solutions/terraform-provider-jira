---
page_title: "jira_issue_type Data Source - jira"
subcategory: ""
description: |-
  Fetches an issue type from JIRA.
---

# jira_issue_type (Data Source)

Fetches an issue type from JIRA. Use this data source to look up existing issue types by name or ID.

## Example Usage

```terraform
# Look up issue types by name
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

# Look up by ID
data "jira_issue_type" "by_id" {
  id = "10001"
}

# Use in an issue type scheme
resource "jira_issue_type_scheme" "software" {
  name                  = "Software Development Scheme"
  default_issue_type_id = data.jira_issue_type.task.id

  issue_type_ids = [
    data.jira_issue_type.epic.id,
    data.jira_issue_type.story.id,
    data.jira_issue_type.bug.id,
    data.jira_issue_type.task.id,
    data.jira_issue_type.subtask.id,
  ]
}

# Output issue type details
output "story_id" {
  value = data.jira_issue_type.story.id
}
```

## Schema

### Optional

- `name` (String) The name of the issue type to look up.
- `id` (String) The ID of the issue type to look up.

~> **Note:** Exactly one of `name` or `id` must be specified.

### Read-Only

- `description` (String) The description of the issue type.
- `subtask` (Boolean) Whether this is a subtask issue type.
