# ============================================================================
# JIRA Terraform Provider - Minimal Example
# ============================================================================
# Minimal project using custom schemes (issue types, permissions, workflows).
# Uses environment variables for credentials; no secrets in code.
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

provider "jira" {}

# Look up an existing user by email (use your project lead's email)
data "jira_user" "project_lead" {
  email_address = "lead@example.com"
}

resource "jira_project" "demo" {
  key                   = "DEMO"
  name                  = "Demo Project"
  description           = "Minimal demo project with custom schemes"
  project_type_key      = "software"
  lead_account_id       = data.jira_user.project_lead.account_id
  issue_type_scheme_id  = jira_issue_type_scheme.demo.id
  permission_scheme_id  = jira_permission_scheme.demo.id
  workflow_scheme_id    = jira_workflow_scheme.demo.id
}

data "jira_permission_scheme" "default" {
  name = "Default Permission Scheme"
}

data "jira_issue_type_scheme" "default" {
  name = "Default Issue Type Scheme"
}
