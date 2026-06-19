import { useCallback } from "react";
import type { LocaleId } from "@/lib/locale";
import { useLocaleStore } from "@/lib/locale-store";
import { en } from "./en";
import { ru } from "./ru";

const catalogs: Record<LocaleId, Record<string, unknown>> = { en, ru };

function getByPath(obj: Record<string, unknown>, path: string): string | undefined {
  const parts = path.split(".");
  let cur: unknown = obj;
  for (const part of parts) {
    if (cur == null || typeof cur !== "object") return undefined;
    cur = (cur as Record<string, unknown>)[part];
  }
  return typeof cur === "string" ? cur : undefined;
}

function interpolate(template: string, vars?: Record<string, string | number>): string {
  if (!vars) return template;
  let out = template;
  for (const [key, value] of Object.entries(vars)) {
    out = out.replace(new RegExp(`\\{${key}\\}`, "g"), String(value));
  }
  return out;
}

export function translate(localeId: LocaleId, key: string, vars?: Record<string, string | number>): string {
  const msg = getByPath(catalogs[localeId] as unknown as Record<string, unknown>, key)
    ?? getByPath(en as unknown as Record<string, unknown>, key)
    ?? key;
  return interpolate(msg, vars);
}

/** Localized "{n} page(s)" with Russian plural rules. */
export function formatPagesCount(localeId: LocaleId, count: number): string {
  if (localeId === "ru") {
    const mod10 = count % 10;
    const mod100 = count % 100;
    if (mod10 === 1 && mod100 !== 11) return `${count} страница`;
    if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return `${count} страницы`;
    return `${count} страниц`;
  }
  return count === 1 ? `${count} page` : `${count} pages`;
}

function translateOrRaw(localeId: LocaleId, key: string, raw: string): string {
  const msg = translate(localeId, key);
  return msg === key ? raw : msg;
}

export function bookStatusLabel(localeId: LocaleId, status: string): string {
  return translateOrRaw(localeId, `book.statuses.${status}`, status);
}

export function bookAudienceLabel(localeId: LocaleId, audience: string): string {
  return translateOrRaw(localeId, `books.audiences.${audience}`, audience);
}

export function useI18n() {
  const localeId = useLocaleStore((s) => s.localeId);
  const t = useCallback(
    (key: string, vars?: Record<string, string | number>) => translate(localeId, key, vars),
    [localeId],
  );
  const pagesCount = useCallback((count: number) => formatPagesCount(localeId, count), [localeId]);
  const statusLabel = useCallback((status: string) => bookStatusLabel(localeId, status), [localeId]);
  const audienceLabel = useCallback((audience: string) => bookAudienceLabel(localeId, audience), [localeId]);
  return { t, localeId, pagesCount, statusLabel, audienceLabel };
}
