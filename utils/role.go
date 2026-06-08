package utils

import "slices"

var accountRoles = []string{"admin", "user"}

func HasRole(roles []string, role string) bool {
	return slices.Contains(roles, role)
}

func HasAnyRole(roles []string, requiredRoles ...string) bool {
	for _, role := range requiredRoles {
		if HasRole(roles, role) {
			return true
		}
	}

	return false
}

func IsAdmin(roles []string) bool {
	return HasRole(roles, "admin")
}

func AccountRoles() []string {
	roles := make([]string, len(accountRoles))
	copy(roles, accountRoles)

	return roles
}

func IsViewer(roles []string) bool {
	return HasRole(roles, "viewer")
}

func IsOperator(roles []string) bool {
	return HasRole(roles, "operator")
}

func IsPPK(roles []string) bool {
	return HasRole(roles, "ppk")
}

func IsKetua(roles []string) bool {
	return HasRole(roles, "ketua")
}
