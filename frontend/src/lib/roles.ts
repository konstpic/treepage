export type SpaceRole = "viewer" | "editor" | "admin" | "";

interface RoleUser {
  roles: string[];
}

const SPACE_EDITOR_ROLES = new Set<SpaceRole>(["editor", "admin"]);

export function isPlatformAdmin(user: RoleUser | null | undefined): boolean {
  if (!user) return false;
  return user.roles.some((role) => role === "super_admin" || role === "admin");
}

export function canEditInSpace(
  spaceRole: SpaceRole | undefined,
  user: RoleUser | null | undefined,
  canEdit?: boolean,
): boolean {
  if (canEdit === true) return true;
  if (!user) return false;
  if (user.roles.includes("super_admin")) return true;
  return SPACE_EDITOR_ROLES.has(spaceRole ?? "");
}

export function canManageBooksInSpace(
  spaceRole: SpaceRole | undefined,
  user: RoleUser | null | undefined,
  canEdit?: boolean,
): boolean {
  return canEditInSpace(spaceRole, user, canEdit);
}

/** @deprecated Use canManageBooksInSpace with space context */
export function canManageBooks(user: RoleUser | null | undefined): boolean {
  if (!user) return false;
  return user.roles.some((role) => role === "super_admin" || role === "admin" || role === "editor");
}
