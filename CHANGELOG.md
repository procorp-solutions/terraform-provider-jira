# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

- Initial open-source release.

## [0.1.0] - TBD

### Added

- **Resources**
  - `jira_project` — Manage JIRA projects
  - `jira_project_component` — Manage project components
  - `jira_workflow_scheme` — Manage workflow schemes
  - `jira_permission_scheme` — Manage permission schemes
  - `jira_issue_type` — Manage issue types
  - `jira_issue_type_scheme` — Manage issue type schemes
  - `jira_custom_field` — Manage custom fields
  - `jira_automation_rule` — Manage automation rules (disabled on destroy; deletion not supported by JIRA API)
  - `jira_group` — Manage user groups
  - `jira_group_membership` — Manage group membership
- **Data sources**
  - `jira_workflow` — Look up workflow by name
  - `jira_user` — Look up user by account ID or email
  - `jira_issue_type` — Look up issue type by ID or name
  - `jira_permission_scheme` — Look up permission scheme by ID or name
  - `jira_issue_type_scheme` — Look up issue type scheme by ID or name
  - `jira_group` — Look up group by ID or name
- Provider configuration via `url`, `email`, `api_token` or environment variables `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`

[Unreleased]: https://github.com/procorp-solutions/terraform-provider-jira/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/procorp-solutions/terraform-provider-jira/releases/tag/v0.1.0
