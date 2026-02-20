---
page_title: "jira Provider"
description: |-
  Manage JIRA Cloud resources using Terraform.
---

# jira Provider

The JIRA provider allows you to manage [JIRA Cloud](https://www.atlassian.com/software/jira) resources using Terraform. Use it to manage projects, workflows, permission schemes, issue types, custom fields, groups, and more as infrastructure as code.

## Features

- **Projects** — Create and manage JIRA projects with custom schemes
- **Issue Types** — Define custom issue types and issue type schemes
- **Workflows** — Assign workflow schemes to projects
- **Permissions** — Configure permission schemes with fine-grained access control
- **Custom Fields** — Create custom fields for your issues
- **Groups** — Manage user groups and group memberships
- **Automation** — Configure automation rules

## Example Usage

```terraform
terraform {
  required_version = ">= 1.0"

  required_providers {
    jira = {
      source  = "procorp-solutions/jira"
      version = "~> 0.1"
    }
  }
}

# Configure the JIRA provider
# Credentials can be provided via:
# 1. Provider configuration (shown below)
# 2. Environment variables (JIRA_URL, JIRA_EMAIL, JIRA_API_TOKEN)
provider "jira" {
  # Required: JIRA Cloud instance URL
  url = var.jira_url

  # Required: Atlassian account email
  email = var.jira_email

  # Required: Atlassian API token
  api_token = var.jira_api_token
}

# Input variables for provider configuration
variable "jira_url" {
  type        = string
  description = "JIRA Cloud instance URL (e.g., 'https://your-org.atlassian.net')"
}

variable "jira_email" {
  type        = string
  description = "Atlassian account email"
}

variable "jira_api_token" {
  type        = string
  description = "Atlassian API token"
  sensitive   = true
}

# Look up an existing user
data "jira_user" "lead" {
  email_address = "lead@example.com"
}

# Create a project with custom schemes
resource "jira_project" "example" {
  key              = "EXAM"
  name             = "Example Project"
  description      = "Managed by Terraform"
  project_type_key = "software"
  lead_account_id  = data.jira_user.lead.account_id
}
```

## Authentication

The provider requires authentication to your JIRA Cloud instance. You need:

1. **JIRA Cloud URL** — Your Atlassian instance URL (e.g., `https://your-org.atlassian.net`)
2. **Email** — The email address of your Atlassian account
3. **API Token** — An [Atlassian API token](https://id.atlassian.com/manage-profile/security/api-tokens)

### Environment Variables

You can configure the provider using environment variables:

```bash
export JIRA_URL="https://your-org.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-api-token"
```

Then use an empty provider block:

```terraform
provider "jira" {}
```

## Schema

### Optional

- `url` (String) JIRA Cloud instance URL (e.g., `https://your-org.atlassian.net`). Can also be set via the `JIRA_URL` environment variable.
- `email` (String) Atlassian account email. Can also be set via the `JIRA_EMAIL` environment variable.
- `api_token` (String, Sensitive) Atlassian API token. Can also be set via the `JIRA_API_TOKEN` environment variable.
