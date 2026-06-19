export const SPACES_SORT_MODES = ["name_asc", "name_desc", "slug_asc"] as const;

export type SpacesSortMode = (typeof SPACES_SORT_MODES)[number];

export const DEFAULT_SPACES_SORT: SpacesSortMode = "name_asc";

const BASE_KEY = "treepage_spaces_sort";

export function isSpacesSortMode(value: string): value is SpacesSortMode {
  return SPACES_SORT_MODES.includes(value as SpacesSortMode);
}

export function spacesSortStorageKey(userId?: string | null): string {
  return userId ? `${BASE_KEY}:${userId}` : `${BASE_KEY}:guest`;
}

export function readSpacesSort(userId?: string | null): SpacesSortMode {
  try {
    const cached = localStorage.getItem(spacesSortStorageKey(userId));
    if (cached && isSpacesSortMode(cached)) return cached;
  } catch {
    /* private mode */
  }
  return DEFAULT_SPACES_SORT;
}

export function writeSpacesSort(userId: string | null | undefined, mode: SpacesSortMode) {
  try {
    localStorage.setItem(spacesSortStorageKey(userId), mode);
  } catch {
    /* private mode */
  }
}

export function sortSpaces<T extends { name: string; slug: string }>(items: T[], mode: SpacesSortMode): T[] {
  const list = [...items];
  switch (mode) {
    case "name_desc":
      return list.sort((a, b) => b.name.localeCompare(a.name, undefined, { sensitivity: "base" }));
    case "slug_asc":
      return list.sort((a, b) => a.slug.localeCompare(b.slug, undefined, { sensitivity: "base" }));
    default:
      return list.sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: "base" }));
  }
}
