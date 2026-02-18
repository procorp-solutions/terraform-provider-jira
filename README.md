# JIRA Terraform Provider

A Terraform provider for managing [JIRA Cloud](https://www.atlassian.com/software/jira) resources via the REST API. Use it to manage projects, workflows, permission schemes, issue types, custom fields, groups, and more as code.

## Features

- **Resources:** Projects, components, workflow schemes, permission schemes, issue types and schemes, custom fields, automation rules, groups, and group membership.
- **Data sources:** Look up workflows, users, issue types, permission schemes, issue type schemes, and groups by ID or name.
- **Credentials** via provider block or environment variables.

## Requirements

- [Terraform](https://www.terraform.io/downloads) >= 1.0
- JIRA Cloud instance (e.g. `https://your-org.atlassian.net`)
- [Atlassian API token](https://id.atlassian.com/manage-profile/security/api-tokens) for the account used to manage JIRA

## Installation

Add the provider to your configuration and run `terraform init`. Terraform will download it from the [Terraform Registry](https://registry.terraform.io):

```hcl
terraform {
  required_providers {
    jira = {
      source  = "procorp-solutions/jira"
      version = "~> 0.1"
    }
  }
}
```

## Configuration

Configure the provider with a block or with environment variables (useful to avoid storing credentials in code):

```hcl
provider "jira" {
  url       = "https://your-org.atlassian.net"  # or set JIRA_URL
  email     = "your-email@example.com"          # or set JIRA_EMAIL
  api_token = "your-api-token"                  # or set JIRA_API_TOKEN (sensitive)
}
```

Environment variables: `JIRA_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`. Provider attributes override env vars when both are set.

## Quick start

Set credentials (do not commit these), then init and plan:

```bash
export JIRA_URL="https://your-org.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-api-token"
cd examples/
terraform init && terraform plan
```

Or run the minimal example: `cd examples/minimal/` then `terraform init && terraform plan`.

## Supported resources and data sources

| Resource | Description |
|----------|-------------|
| `jira_project` | JIRA project |
| `jira_project_component` | Project component |
| `jira_workflow_scheme` | Workflow scheme |
| `jira_permission_scheme` | Permission scheme |
| `jira_issue_type` | Issue type |
| `jira_issue_type_scheme` | Issue type scheme |
| `jira_custom_field` | Custom field |
| `jira_automation_rule` | Automation rule |
| `jira_group` | User group |
| `jira_group_membership` | Group membership |

| Data source | Description |
|-------------|-------------|
| `jira_workflow` | Workflow by name |
| `jira_user` | User by account ID or email |
| `jira_issue_type` | Issue type by ID or name |
| `jira_permission_scheme` | Permission scheme by ID or name |
| `jira_issue_type_scheme` | Issue type scheme by ID or name |
| `jira_group` | Group by ID or name |

## Examples

- **[examples/](examples/)** — Full example: project, components, issue types, custom fields, workflow scheme, permission scheme, groups, automation rule.
- **[examples/minimal/](examples/minimal/)** — Minimal example: one project using custom issue type, permission, and workflow schemes (data sources only for existing workflows and issue types).

## Documentation

Full resource and data source reference: [docs/README.md](docs/README.md).

## Publishing to the Terraform Registry

To publish this provider to the [Terraform Registry](https://registry.terraform.io):

1. **Repository:** Use a **public** GitHub repo named `terraform-provider-jira` (lowercase). The registry only discovers repos matching `terraform-provider-{name}`.

2. **GPG key:** Generate an RSA GPG key for signing releases (registry does not support ECC):
   ```bash
   gpg --full-generate-key
   ```
   Choose “RSA and RSA”, 4096 bits. Export the public key and add it in [Registry → User Settings → Signing Keys](https://registry.terraform.io/settings/gpg-keys). Export the **private** key (ASCII-armored) for the next step.

3. **GitHub Actions secrets:** In the repo go to **Settings → Secrets and variables → Actions** and add:
   - `GPG_PRIVATE_KEY` — ASCII-armored private key (including `-----BEGIN...` and `-----END...`).
   - `PASSPHRASE` — passphrase for the GPG key.

4. **Release:** Push a version tag (semver with `v` prefix). The [Release](.github/workflows/release.yml) workflow will build binaries and create a GitHub Release:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

5. **Publish on the registry:** In the [Terraform Registry](https://registry.terraform.io), go to **Publish → Provider**, sign in with GitHub, select the repository, choose a category, accept the terms, and click **Publish**. The registry will ingest the release and future tags will be picked up via webhook.

See [Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing) for full details.

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) for how to build, test, and submit changes.
