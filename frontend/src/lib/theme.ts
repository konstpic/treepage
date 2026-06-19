export const UI_THEMES = [
  { id: "fox_white", label: "Fox White", description: "Default GitLab-inspired light UI with orange brand accents" },
  { id: "coral_night", label: "Coral Night", description: "Purple/cyan dark theme" },
  { id: "light", label: "Light", description: "Neutral minimal light theme" },
  { id: "dark", label: "Dark", description: "Neutral slate dark theme" },
] as const;

export type UIThemeId = (typeof UI_THEMES)[number]["id"];

export const DEFAULT_UI_THEME: UIThemeId = "fox_white";

const STORAGE_KEY = "treepage_ui_theme";

export function isUIThemeId(value: string): value is UIThemeId {
  return UI_THEMES.some((t) => t.id === value);
}

export function applyUITheme(themeId: UIThemeId) {
  document.documentElement.dataset.theme = themeId;
  try {
    localStorage.setItem(STORAGE_KEY, themeId);
  } catch {
    /* private mode */
  }
}

/** Sync read before React — prevents flash (see index.html). */
export function readCachedUITheme(): UIThemeId {
  try {
    const cached = localStorage.getItem(STORAGE_KEY);
    if (cached && isUIThemeId(cached)) return cached;
  } catch {
    /* ignore */
  }
  return DEFAULT_UI_THEME;
}

export async function fetchPublicUITheme(apiUrl: string): Promise<UIThemeId> {
  const res = await fetch(`${apiUrl}/api/public/ui-theme`);
  if (!res.ok) return readCachedUITheme();
  const data = (await res.json()) as { ui_theme?: string };
  if (data.ui_theme && isUIThemeId(data.ui_theme)) {
    applyUITheme(data.ui_theme);
    return data.ui_theme;
  }
  return readCachedUITheme();
}
