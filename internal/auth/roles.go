// Package auth owns everything that identifies and authorizes an operator:
// GitHub OAuth client, session cookies, request middleware, and the role
// model used by membership checks.
package auth

import "slices"

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

var allRoles = []string{RoleAdmin, RoleOperator, RoleViewer}

func ValidRole(role string) bool {
	return slices.Contains(allRoles, role)
}

// The three capability helpers below are the only place role semantics live.
// Handlers should call these by name rather than string-comparing roles, so
// changes to the role model touch one file instead of every endpoint.

// CanPromoteDeploy gates the override endpoints (promote/hold).
func CanPromoteDeploy(role string) bool {
	return role == RoleAdmin || role == RoleOperator
}

// CanInviteMembers gates invite creation and member removal.
func CanInviteMembers(role string) bool {
	return role == RoleAdmin
}

// CanViewExperiments gates the read-only experiment endpoints.
func CanViewExperiments(role string) bool {
	return role == RoleAdmin || role == RoleOperator || role == RoleViewer
}
