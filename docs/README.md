# JIRA Terraform Provider — Full documentation

This document is the full resource and data source reference for the JIRA Terraform provider. For installation and quick start, see the [project README](../README.md).

## Configuration

```hcl
provider "jira" {
  url       = "https://your-org.atlassian.net"  # or JIRA_URL env var
  email     = "your-email@example.com"          # or JIRA_EMAIL env var
  api_token = "your-api-token"                  # or JIRA_API_TOKEN env var
}
```

## Resources

| Resource | Description |
|---|---|
| `jira_project` | Manages a JIRA project |
| `jira_project_component` | Manages a project component |
| `jira_workflow_scheme` | Manages a workflow scheme |
| `jira_permission_scheme` | Manages a permission scheme with grants |
| `jira_issue_type` | Manages an issue type |
| `jira_issue_type_scheme` | Manages an issue type scheme |
| `jira_custom_field` | Manages a custom field |
| `jira_automation_rule` | Manages an automation rule (disables on destroy) |
| `jira_group` | Manages a user group |
| `jira_group_membership` | Manages user-to-group membership |

## Data Sources

| Data Source | Description |
|---|---|
| `jira_workflow` | Looks up a workflow by name |
| `jira_user` | Looks up a user by account ID or email |
| `jira_issue_type` | Looks up an issue type by ID or name |
| `jira_permission_scheme` | Looks up a permission scheme by ID or name |
| `jira_issue_type_scheme` | Looks up an issue type scheme by ID or name |
| `jira_group` | Looks up a group by ID or name |

## Resource Reference

### jira_project

```hcl
resource "jira_project" "example" {
  key                   = "EXAM"
  name                  = "Example Project"
  description           = "Managed by Terraform"
  project_type_key      = "software"
  lead_account_id       = "5b10a2844c20165700ede21g"
  assignee_type         = "UNASSIGNED"
  # Optional: assign schemes (use IDs from jira_issue_type_scheme, jira_permission_scheme, jira_workflow_scheme)
  # issue_type_scheme_id   = jira_issue_type_scheme.default.id
  # permission_scheme_id   = jira_permission_scheme.standard.id
  # workflow_scheme_id     = jira_workflow_scheme.default.id
}
```

### jira_project_component

```hcl
resource "jira_project_component" "frontend" {
  project_key     = jira_project.example.key
  name            = "Frontend"
  description     = "Frontend components"
  lead_account_id = "5b10a2844c20165700ede21g"
  assignee_type   = "COMPONENT_LEAD"
}
```

### jira_workflow_scheme

```hcl
resource "jira_workflow_scheme" "default" {
  name             = "My Workflow Scheme"
  description      = "Custom workflow scheme"
  default_workflow = "jira"

  issue_type_mappings = {
    "10001" = "custom-workflow"
  }
}
```

### jira_permission_scheme

```hcl
resource "jira_permission_scheme" "standard" {
  name        = "Standard Permissions"
  description = "Standard permission scheme"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = "jira-software-users"
    },
    {
      permission       = "ADMINISTER_PROJECTS"
      holder_type      = "projectRole"
      holder_parameter = "10002"
    },
  ]
}
```

### jira_issue_type

```hcl
resource "jira_issue_type" "bug" {
  name        = "Bug"
  description = "A software defect"
  type        = "standard"  # or "subtask"
}
```

### jira_issue_type_scheme

```hcl
resource "jira_issue_type_scheme" "default" {
  name                  = "Default Scheme"
  default_issue_type_id = jira_issue_type.task.id
  issue_type_ids        = [jira_issue_type.bug.id, jira_issue_type.task.id]
}
```

### jira_custom_field

```hcl
resource "jira_custom_field" "story_points" {
  name        = "Story Points"
  description = "Estimated effort"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:numberrange"
}
```

### jira_automation_rule

```hcl
resource "jira_automation_rule" "auto_assign" {
  name  = "Auto-assign"
  state = "ENABLED"
  rule_json = jsonencode({
    trigger    = { type = "issue.created" }
    components = [{ type = "action", value = "assign" }]
  })
}
```

### jira_group

```hcl
resource "jira_group" "developers" {
  name = "developers"
}
```

### jira_group_membership

```hcl
resource "jira_group_membership" "dev_member" {
  group_name = jira_group.developers.name
  account_id = "5b10a2844c20165700ede21g"
}
```

### Data: jira_workflow

```hcl
data "jira_workflow" "default" {
  name = "jira"
}
```

### Data: jira_user

```hcl
data "jira_user" "lead" {
  email_address = "lead@example.com"
}

# Or by account ID:
data "jira_user" "specific" {
  account_id = "5b10a2844c20165700ede21g"
}
```

### Data: jira_issue_type

```hcl
data "jira_issue_type" "task" {
  name = "Task"
}

# Or by issue type ID:
data "jira_issue_type" "by_id" {
  id = "10001"
}
```

### Data: jira_permission_scheme

```hcl
data "jira_permission_scheme" "by_name" {
  name = "Default Permission Scheme"
}

# Or by permission scheme ID:
data "jira_permission_scheme" "by_id" {
  id = "10000"
}
```

### Data: jira_issue_type_scheme

```hcl
data "jira_issue_type_scheme" "by_name" {
  name = "Default Issue Type Scheme"
}

# Or by issue type scheme ID:
data "jira_issue_type_scheme" "by_id" {
  id = "10000"
}
```

### Data: jira_group

```hcl
data "jira_group" "by_name" {
  name = "jira-software-users"
}

# Or by group ID:
data "jira_group" "by_id" {
  id = "10000"
}
```

## Known Limitations

- **Automation rules** cannot be deleted via the JIRA Cloud API. On `terraform destroy`, rules are disabled instead.
- **Users** are managed through Atlassian Admin, not the JIRA REST API. Only a read-only data source is provided.
- **Workflows** are complex graph structures. Only a read-only data source is provided; use workflow _schemes_ to assign workflows to projects.
