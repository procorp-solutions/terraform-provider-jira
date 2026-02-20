---
page_title: "jira_custom_field Resource - jira"
subcategory: ""
description: |-
  Manages a custom field in JIRA.
---

# jira_custom_field (Resource)

Manages a custom field in JIRA. Custom fields allow you to capture additional information on issues.

## Example Usage

```terraform
# Text area field
resource "jira_custom_field" "sprint_goal" {
  name        = "Sprint Goal"
  description = "The goal for the sprint"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:textarea"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:textsearcher"
}

# Number field for story points
resource "jira_custom_field" "story_points" {
  name        = "Story Points"
  description = "Estimated effort in story points"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:numberrange"
}

# Select list field
resource "jira_custom_field" "team" {
  name        = "Team"
  description = "The team responsible for this issue"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:select"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:multiselectsearcher"
}

# Date picker field
resource "jira_custom_field" "target_date" {
  name        = "Target Date"
  description = "The target completion date"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:datepicker"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:daterange"
}
```

## Schema

### Required

- `name` (String) The name of the custom field.
- `type` (String) The custom field type. See common types below.
- `search_key` (String) The searcher key for the field.

### Optional

- `description` (String) A description of the custom field.

### Read-Only

- `id` (String) The ID of the custom field.

## Common Field Types

| Type | Description |
|------|-------------|
| `com.atlassian.jira.plugin.system.customfieldtypes:textfield` | Single line text |
| `com.atlassian.jira.plugin.system.customfieldtypes:textarea` | Multi-line text |
| `com.atlassian.jira.plugin.system.customfieldtypes:float` | Number |
| `com.atlassian.jira.plugin.system.customfieldtypes:select` | Single select |
| `com.atlassian.jira.plugin.system.customfieldtypes:multiselect` | Multi-select |
| `com.atlassian.jira.plugin.system.customfieldtypes:datepicker` | Date |
| `com.atlassian.jira.plugin.system.customfieldtypes:datetime` | Date and time |
| `com.atlassian.jira.plugin.system.customfieldtypes:userpicker` | User picker |
| `com.atlassian.jira.plugin.system.customfieldtypes:url` | URL |
| `com.atlassian.jira.plugin.system.customfieldtypes:labels` | Labels |

## Import

Custom fields can be imported using the field ID:

```shell
terraform import jira_custom_field.story_points customfield_10001
```
