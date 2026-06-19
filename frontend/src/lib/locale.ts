export const UI_LANGUAGES = [
  { id: "en", label: "English", nativeLabel: "English" },
  { id: "ru", label: "Russian", nativeLabel: "Русский" },
] as const;

export type LocaleId = (typeof UI_LANGUAGES)[number]["id"];

export const DEFAULT_LOCALE: LocaleId = "en";

const STORAGE_KEY = "treepage_ui_language";

export function isLocaleId(value: string): value is LocaleId {
  return UI_LANGUAGES.some((l) => l.id === value);
}

export function applyUILocale(localeId: LocaleId) {
  document.documentElement.lang = localeId;
  try {
    localStorage.setItem(STORAGE_KEY, localeId);
  } catch {
    /* private mode */
  }
}

export function readCachedLocale(): LocaleId {
  try {
    const cached = localStorage.getItem(STORAGE_KEY);
    if (cached && isLocaleId(cached)) return cached;
  } catch {
    /* ignore */
  }
  return DEFAULT_LOCALE;
}

export async function fetchPublicUILocale(apiUrl: string): Promise<LocaleId> {
  const res = await fetch(`${apiUrl}/api/public/ui-language`);
  if (!res.ok) return readCachedLocale();
  const data = (await res.json()) as { ui_language?: string };
  if (data.ui_language && isLocaleId(data.ui_language)) {
    applyUILocale(data.ui_language);
    return data.ui_language;
  }
  return readCachedLocale();
}
