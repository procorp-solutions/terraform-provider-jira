---
page_title: "jira_automation_rule Resource - jira"
subcategory: ""
description: |-
  Manages an automation rule in JIRA.
---

# jira_automation_rule (Resource)

Manages an automation rule in JIRA. Automation rules automatically perform actions when specified triggers occur.

~> **Note:** Automation rules cannot be deleted via the JIRA Cloud API. When you run `terraform destroy`, the rule will be **disabled** instead of deleted.

## Example Usage

```terraform
# Auto-assign issues to component lead
resource "jira_automation_rule" "auto_assign" {
  name  = "Auto-assign to component lead"
  state = "ENABLED"

  rule_json = jsonencode({
    trigger = {
      type = "issue.created"
    }
    components = [
      {
        type  = "action"
        value = "assign_to_component_lead"
      }
    ]
  })
}

# Notify on high priority issues
resource "jira_automation_rule" "notify_high_priority" {
  name  = "Notify on high priority"
  state = "ENABLED"

  rule_json = jsonencode({
    trigger = {
      type = "issue.created"
      conditions = {
        priority = "High"
      }
    }
    components = [
      {
        type  = "action"
        value = "send_notification"
        config = {
          to = "project-lead"
        }
      }
    ]
  })
}

# Disabled rule (for testing)
resource "jira_automation_rule" "draft" {
  name  = "Draft rule"
  state = "DISABLED"

  rule_json = jsonencode({
    trigger = {
      type = "manual"
    }
    components = []
  })
}
```

## Schema

### Required

- `name` (String) The name of the automation rule.
- `state` (String) The state of the rule. Valid values: `ENABLED`, `DISABLED`.
- `rule_json` (String) The rule configuration as a JSON string.

### Read-Only

- `id` (String) The ID of the automation rule.

## Import

Automation rules can be imported using the rule ID:

```shell
terraform import jira_automation_rule.auto_assign 10001
```
