# ============================================================================
# JIRA Terraform Provider - Example Configuration
# ============================================================================
# This example demonstrates all available resources and data sources.
#
# Prerequisites:
#   export JIRA_URL="https://your-org.atlassian.net"
#   export JIRA_EMAIL="your-email@example.com"
#   export JIRA_API_TOKEN="your-api-token"
# ============================================================================

terraform {
  required_providers {
    jira = {
      source  = "procorp-solutions/jira"
      version = "~> 0.1"
    }
  }
}

# Provider configuration (values come from environment variables)
provider "jira" {}

# ============================================================================
# Data Sources
# ============================================================================

# Look up an existing user by email
data "jira_user" "project_lead" {
  email_address = "lead@example.com"
}

# Look up an existing workflow
data "jira_workflow" "default" {
  name = "jira"
}

# Look up an existing issue type by name
data "jira_issue_type" "existing_task" {
  name = "Task"
}

# ============================================================================
# Issue Types
# ============================================================================

resource "jira_issue_type" "story" {
  name        = "Story"
  description = "A user story"
  type        = "standard"
}

resource "jira_issue_type" "bug" {
  name        = "Bug"
  description = "A software defect"
  type        = "standard"
}

resource "jira_issue_type" "subtask" {
  name        = "Sub-task"
  description = "A sub-task of an issue"
  type        = "subtask"
}

resource "jira_issue_type" "task" {
  name        = "Task"
  description = "A task that needs to be done"
  type        = "standard"
}

resource "jira_issue_type" "epic" {
  name        = "Epic"
  description = "A collection of related stories"
  type        = "standard"
}

# ============================================================================
# Issue Type Scheme
# ============================================================================

resource "jira_issue_type_scheme" "default" {
  name                  = "Default Issue Type Scheme"
  description           = "Standard issue type scheme for all projects"
  default_issue_type_id = jira_issue_type.task.id

  issue_type_ids = [
    jira_issue_type.epic.id,
    jira_issue_type.story.id,
    jira_issue_type.bug.id,
    jira_issue_type.task.id,
    jira_issue_type.subtask.id,
  ]
}

# ============================================================================
# Custom Fields
# ============================================================================

resource "jira_custom_field" "sprint_goal" {
  name        = "Sprint Goal"
  description = "The goal for the sprint"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:textarea"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:textsearcher"
}

resource "jira_custom_field" "story_points" {
  name        = "Story Points"
  description = "Estimated effort in story points"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:float"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:numberrange"
}

resource "jira_custom_field" "team" {
  name        = "Team"
  description = "The team responsible for this issue"
  type        = "com.atlassian.jira.plugin.system.customfieldtypes:select"
  search_key  = "com.atlassian.jira.plugin.system.customfieldtypes:multiselectsearcher"
}

# ============================================================================
# Workflow Scheme
# ============================================================================

resource "jira_workflow_scheme" "software" {
  name             = "Software Development Workflow Scheme"
  description      = "Workflow scheme for software development projects"
  default_workflow = "jira"

  issue_type_mappings = {
    (jira_issue_type.bug.id) = "jira"
  }
}

# ============================================================================
# Permission Scheme
# ============================================================================

resource "jira_permission_scheme" "standard" {
  name        = "Standard Permission Scheme"
  description = "Standard permissions for development projects"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = "jira-software-users"
    },
    {
      permission       = "CREATE_ISSUES"
      holder_type      = "group"
      holder_parameter = "jira-software-users"
    },
    {
      permission       = "EDIT_ISSUES"
      holder_type      = "group"
      holder_parameter = "jira-software-users"
    },
    {
      permission       = "ADMINISTER_PROJECTS"
      holder_type      = "group"
      holder_parameter = jira_group.admins.name
    },
    {
      permission       = "ASSIGN_ISSUES"
      holder_type      = "projectRole"
      holder_parameter = "10002"
    },
  ]
}

# ============================================================================
# Groups
# ============================================================================

resource "jira_group" "developers" {
  name = "developers"
}

resource "jira_group" "admins" {
  name = "project-administrators"
}

# Add the lead user to the admins group
resource "jira_group_membership" "lead_admin" {
  group_name = jira_group.admins.name
  account_id = data.jira_user.project_lead.account_id
}

# ============================================================================
# Project
# ============================================================================

resource "jira_project" "example" {
  key                   = "EXAM"
  name                  = "Example Project"
  description           = "An example project managed by Terraform"
  project_type_key      = "software"
  lead_account_id       = data.jira_user.project_lead.account_id
  assignee_type         = "UNASSIGNED"
  issue_type_scheme_id  = jira_issue_type_scheme.default.id
  permission_scheme_id  = jira_permission_scheme.standard.id
  workflow_scheme_id    = jira_workflow_scheme.software.id
}

# ============================================================================
# Project Components
# ============================================================================

resource "jira_project_component" "frontend" {
  project_key    = jira_project.example.key
  name           = "Frontend"
  description    = "Frontend components and UI"
  lead_account_id = data.jira_user.project_lead.account_id
  assignee_type  = "COMPONENT_LEAD"
}

resource "jira_project_component" "backend" {
  project_key = jira_project.example.key
  name        = "Backend"
  description = "Backend services and APIs"
}

resource "jira_project_component" "infrastructure" {
  project_key = jira_project.example.key
  name        = "Infrastructure"
  description = "Infrastructure and DevOps"
}

# ============================================================================
# Automation Rule
# ============================================================================

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
