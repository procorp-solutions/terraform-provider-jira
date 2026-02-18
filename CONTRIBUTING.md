# Contributing to terraform-provider-jira

Thank you for your interest in contributing. This document explains how to build the provider, run tests, and submit changes.

## Development setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/procorp-solutions/terraform-provider-jira.git
   cd terraform-provider-jira
   ```

2. **Install Go**

   The project uses Go 1.22 or later. See [go.mod](go.mod) for the exact version.

3. **Build the provider**

   ```bash
   go build -o terraform-provider-jira .
   ```

## Running tests

- Run the Go test suite:

  ```bash
  go test ./...
  ```

- Optionally run Terraform against the examples (requires a JIRA Cloud instance, credentials, and the provider published to the Terraform Registry so `terraform init` can download it).

## Code style

- Format Go code with `gofmt` (or use your editor’s format-on-save).
- Follow standard Go conventions and [HashiCorp’s Terraform provider development practices](https://developer.hashicorp.com/terraform/plugin) where applicable.
- Keep resource and data source logic in the existing package layout under `internal/`.

## Submitting changes

1. Open an issue to discuss larger changes or new features.
2. Fork the repo and create a branch for your change.
3. Make your edits, ensure `go build ./...` and `go test ./...` pass.
4. Commit with clear messages; reference the issue number if applicable.
5. Open a pull request. Describe what changed and why; link any related issues.

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT).
