resource "jira_issue_type_scheme" "demo" {
  name                  = "Demo Issue Type Scheme"
  description           = "Issue type scheme for the minimal example"
  default_issue_type_id = data.jira_issue_type.story.id
  issue_type_ids = [
    data.jira_issue_type.story.id,
    data.jira_issue_type.bug.id,
    data.jira_issue_type.subtask.id,
    data.jira_issue_type.epic.id,
    data.jira_issue_type.task.id,
  ]
}

data "jira_issue_type" "story" {
  name = "Story"
}

data "jira_issue_type" "bug" {
  name = "Bug"
}

data "jira_issue_type" "subtask" {
  name = "Sub-task"
}

data "jira_issue_type" "task" {
  name = "Task"
}

data "jira_issue_type" "epic" {
  name = "Epic"
}
