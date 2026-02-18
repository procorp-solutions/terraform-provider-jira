# Security

## Reporting a vulnerability

If you believe you have found a security vulnerability in this project, please report it responsibly.

- **Preferred:** Open a [GitHub Security Advisory](https://github.com/procorp-solutions/terraform-provider-jira/security/advisories/new) so maintainers can review and coordinate a fix before public disclosure.
- **Alternative:** Open a private issue or contact the maintainers directly (e.g. via the repository’s listed maintainers) with a description of the issue and steps to reproduce.

Please do not open a public issue for security-sensitive bugs. We will acknowledge your report and work with you on a fix and disclosure timeline.

## Scope

- This provider talks to the JIRA Cloud REST API using credentials you supply (URL, email, API token). Keep credentials out of Terraform config and use environment variables or a secret manager.
- Vulnerabilities in JIRA Cloud itself or in Atlassian’s APIs are out of scope; report those to [Atlassian](https://www.atlassian.com/trust/security/security-advisories).
