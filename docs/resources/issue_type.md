---
page_title: "jira_issue_type Resource - jira"
subcategory: ""
description: |-
  Manages an issue type in JIRA.
---

# jira_issue_type (Resource)

Manages an issue type in JIRA. Issue types distinguish different types of work (e.g., Bug, Story, Task).

## Example Usage

```terraform
# Create a standard issue type
resource "jira_issue_type" "story" {
  name        = "Story"
  description = "A user story"
  type        = "standard"
}

# Create a bug issue type
resource "jira_issue_type" "bug" {
  name        = "Bug"
  description = "A software defect"
  type        = "standard"
}

# Create a subtask issue type
resource "jira_issue_type" "subtask" {
  name        = "Sub-task"
  description = "A sub-task of an issue"
  type        = "subtask"
}

# Create an epic issue type
resource "jira_issue_type" "epic" {
  name        = "Epic"
  description = "A collection of related stories"
  type        = "standard"
}
```

## Schema

### Required

- `name` (String) The name of the issue type.
- `description` (String) A description of the issue type.
- `type` (String) The type category. Valid values: `standard`, `subtask`.

### Read-Only

- `id` (String) The ID of the issue type.

## Import

Issue types can be imported using the issue type ID:

```shell
terraform import jira_issue_type.story 10001
```
