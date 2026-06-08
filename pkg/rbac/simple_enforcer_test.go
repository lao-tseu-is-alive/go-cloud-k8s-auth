package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleEnforcer_IsAllowed(t *testing.T) {
	enforcer := NewSimpleEnforcer()

	tests := []struct {
		name     string
		roles    []string
		action   string
		resource string
		expected bool
	}{
		{
			name:     "Admin can read anything",
			roles:    []string{"admin"},
			action:   "read",
			resource: "users",
			expected: true,
		},
		{
			name:     "Admin can write anything",
			roles:    []string{"admin"},
			action:   "write",
			resource: "users",
			expected: true,
		},
		{
			name:     "Admin can delete anything",
			roles:    []string{"admin"},
			action:   "delete",
			resource: "users",
			expected: true,
		},
		{
			name:     "User can read resource",
			roles:    []string{"user"},
			action:   "read",
			resource: "users",
			expected: true,
		},
		{
			name:     "User cannot write resource",
			roles:    []string{"user"},
			action:   "write",
			resource: "users",
			expected: false,
		},
		{
			name:     "User cannot delete resource",
			roles:    []string{"user"},
			action:   "delete",
			resource: "users",
			expected: false,
		},
		{
			name:     "No roles cannot read",
			roles:    []string{},
			action:   "read",
			resource: "users",
			expected: false,
		},
		{
			name:     "Unknown roles cannot read",
			roles:    []string{"guest"},
			action:   "read",
			resource: "users",
			expected: false,
		},
		{
			name:     "Multiple roles containing admin",
			roles:    []string{"user", "admin"},
			action:   "write",
			resource: "users",
			expected: true,
		},
		{
			name:     "Multiple roles containing user",
			roles:    []string{"guest", "user"},
			action:   "read",
			resource: "users",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := enforcer.IsAllowed(tt.roles, tt.action, tt.resource)
			assert.Equal(t, tt.expected, allowed)
		})
	}
}

func TestSimpleEnforcer_HasRole(t *testing.T) {
	enforcer := NewSimpleEnforcer()

	assert.True(t, enforcer.HasRole([]string{"admin", "user"}, "admin"))
	assert.True(t, enforcer.HasRole([]string{"admin", "user"}, "user"))
	assert.False(t, enforcer.HasRole([]string{"admin", "user"}, "guest"))
	assert.False(t, enforcer.HasRole([]string{}, "user"))
}

func TestSimpleEnforcer_IsAdmin(t *testing.T) {
	enforcer := NewSimpleEnforcer()

	assert.True(t, enforcer.IsAdmin([]string{"admin"}))
	assert.True(t, enforcer.IsAdmin([]string{"user", "admin"}))
	assert.False(t, enforcer.IsAdmin([]string{"user"}))
	assert.False(t, enforcer.IsAdmin([]string{}))
}
