package service

const (
	SpaceRoleViewer = "viewer"
	SpaceRoleEditor = "editor"
	SpaceRoleAdmin  = "admin"
)

var spaceRoleRank = map[string]int{
	SpaceRoleViewer: 1,
	SpaceRoleEditor: 2,
	SpaceRoleAdmin:  3,
}

func SpaceRoleRank(name string) int {
	if r, ok := spaceRoleRank[name]; ok {
		return r
	}
	return 0
}

func MaxSpaceRole(roles ...string) string {
	best := ""
	bestRank := 0
	for _, role := range roles {
		if rank := SpaceRoleRank(role); rank > bestRank {
			bestRank = rank
			best = role
		}
	}
	return best
}

// GlobalRolesAsSpaceRole maps platform roles to a space-role floor.
func GlobalRolesAsSpaceRole(globalRoles []string) string {
	if HasRole(globalRoles, "super_admin", "admin") {
		return SpaceRoleAdmin
	}
	if HasRole(globalRoles, "editor") {
		return SpaceRoleEditor
	}
	if HasRole(globalRoles, "viewer") {
		return SpaceRoleViewer
	}
	return ""
}

func HasSpaceRole(effectiveRole, minRole string) bool {
	return SpaceRoleRank(effectiveRole) >= SpaceRoleRank(minRole)
}

func CanEditInSpace(effectiveRole string, globalRoles []string) bool {
	if HasRole(globalRoles, "super_admin") {
		return true
	}
	return HasSpaceRole(effectiveRole, SpaceRoleEditor)
}

func CanUseLLMInSpace(effectiveRole string, globalRoles []string) bool {
	return CanEditInSpace(effectiveRole, globalRoles)
}

func CanManageBooksInSpace(effectiveRole string, globalRoles []string) bool {
	return CanEditInSpace(effectiveRole, globalRoles)
}
