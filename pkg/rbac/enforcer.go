// Package rbac provides a simple, extensible RBAC (Role-Based Access Control) interface.
//
// Phase 1: Simple role checking from user.roles array.
// Future: Replace SimpleEnforcer with CasbinEnforcer without breaking the interface contract.
package rbac

// Enforcer defines the interface for authorization checks.
// All implementations must be safe for concurrent use.
type Enforcer interface {
	// IsAllowed checks if a user with the given roles can perform action on resource.
	// Returns true if the action is allowed.
	IsAllowed(roles []string, action string, resource string) bool

	// HasRole checks if the given roles contain a specific role.
	HasRole(roles []string, role string) bool

	// IsAdmin checks if the given roles contain the "admin" role.
	IsAdmin(roles []string) bool
}
