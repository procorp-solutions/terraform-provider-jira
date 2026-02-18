locals {
  groups = {
    dev    = "jira-software-users"
    lead   = "project-administrators"
    editor = "jira-software-users"
  }
}

data "jira_group" "dev" {
  name = local.groups.dev
}

data "jira_group" "lead" {
  name = local.groups.lead
}

data "jira_group" "editor" {
  name = local.groups.editor
}

resource "jira_permission_scheme" "demo" {
  name        = "Demo Permission Scheme"
  description = "Permission scheme for the minimal example"

  permissions = [
    {
      permission       = "BROWSE_PROJECTS"
      holder_type      = "group"
      holder_parameter = data.jira_group.dev.name
    },
    {
      permission       = "CREATE_ISSUES"
      holder_type      = "group"
      holder_parameter = data.jira_group.dev.name
    },
    {
      permission       = "EDIT_ISSUES"
      holder_type      = "group"
      holder_parameter = data.jira_group.lead.name
    },
  ]
}
