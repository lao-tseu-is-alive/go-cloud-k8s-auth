package rbac

// SimpleEnforcer implements Enforcer with basic role checking.
//
// Phase 1 Rules:
//   - "admin" can perform any action on any resource
//   - "user" can perform "read" on any resource
//   - All other combinations are denied
//
// This implementation is thread-safe (stateless).
type SimpleEnforcer struct{}

// NewSimpleEnforcer creates a new SimpleEnforcer.
func NewSimpleEnforcer() *SimpleEnforcer {
	return &SimpleEnforcer{}
}

// IsAllowed checks if a user with the given roles can perform action on resource.
func (e *SimpleEnforcer) IsAllowed(roles []string, action string, resource string) bool {
	// Admin can do anything
	if e.IsAdmin(roles) {
		return true
	}

	// Regular users can read
	if e.HasRole(roles, "user") && action == "read" {
		return true
	}

	return false
}

// HasRole checks if roles contain the specified role.
func (e *SimpleEnforcer) HasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if roles contain the "admin" role.
func (e *SimpleEnforcer) IsAdmin(roles []string) bool {
	return e.HasRole(roles, "admin")
}
