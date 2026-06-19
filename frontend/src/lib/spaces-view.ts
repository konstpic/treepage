export const SPACES_VIEW_MODES = ["grid", "table", "list"] as const;

export type SpacesViewMode = (typeof SPACES_VIEW_MODES)[number];

export const DEFAULT_SPACES_VIEW: SpacesViewMode = "grid";

const BASE_KEY = "treepage_spaces_view";

export function isSpacesViewMode(value: string): value is SpacesViewMode {
  return SPACES_VIEW_MODES.includes(value as SpacesViewMode);
}

export function spacesViewStorageKey(userId?: string | null): string {
  return userId ? `${BASE_KEY}:${userId}` : `${BASE_KEY}:guest`;
}

export function readSpacesView(userId?: string | null): SpacesViewMode {
  try {
    const cached = localStorage.getItem(spacesViewStorageKey(userId));
    if (cached && isSpacesViewMode(cached)) return cached;
  } catch {
    /* private mode */
  }
  return DEFAULT_SPACES_VIEW;
}

export function writeSpacesView(userId: string | null | undefined, mode: SpacesViewMode) {
  try {
    localStorage.setItem(spacesViewStorageKey(userId), mode);
  } catch {
    /* private mode */
  }
}
