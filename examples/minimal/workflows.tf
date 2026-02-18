# Use the default Jira workflow (available in all Jira Cloud instances)
data "jira_workflow" "default" {
  name = "jira"
}

resource "jira_workflow_scheme" "demo" {
  name                  = "Demo Workflow Scheme"
  description           = "Workflow scheme for the minimal example"
  default_workflow      = data.jira_workflow.default.name
  issue_type_mappings   = {
    (data.jira_issue_type.story.id) = data.jira_workflow.default.name
    (data.jira_issue_type.task.id)  = data.jira_workflow.default.name
  }
}
