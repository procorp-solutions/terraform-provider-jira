package issuetype

// IsProjectScoped returns true when the issue type has a "scope" field in the API response.
// Global (classic) issue types have no scope; project-scoped (next-gen) ones do.
// This is the single source of truth so datasource and resource behave consistently.
func IsProjectScoped(issueType map[string]interface{}) bool {
	_, hasScope := issueType["scope"]
	return hasScope
}
