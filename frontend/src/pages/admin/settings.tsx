import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Languages, Loader2, Palette } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { FadeIn } from "@/components/motion-wrapper";
import { useAdminGuard } from "./layout";
import { DEFAULT_UI_THEME, UI_THEMES, type UIThemeId } from "@/lib/theme";
import { UI_LANGUAGES, type LocaleId } from "@/lib/locale";
import { useThemeStore } from "@/lib/theme-store";
import { useLocaleStore } from "@/lib/locale-store";
import { useI18n } from "@/lib/i18n";
import { cn } from "@/lib/utils";

interface SystemSettings {
  auth: Record<string, unknown>;
  git: Record<string, unknown>;
  platform: Record<string, unknown>;
  ui_theme: UIThemeId;
  ui_language: LocaleId;
}

export function AdminSettingsPage() {
  const { ready, user } = useAdminGuard();
  const qc = useQueryClient();
  const { t } = useI18n();
  const isSuperAdmin = user?.roles.includes("super_admin") ?? false;
  const setTheme = useThemeStore((s) => s.setTheme);
  const setLocale = useLocaleStore((s) => s.setLocale);

  const [auth, setAuth] = useState<Record<string, unknown>>({});
  const [git, setGit] = useState<Record<string, unknown>>({});
  const [platform, setPlatform] = useState<Record<string, unknown>>({});
  const [uiTheme, setUiTheme] = useState<UIThemeId>(DEFAULT_UI_THEME);
  const [savedTheme, setSavedTheme] = useState<UIThemeId>(DEFAULT_UI_THEME);
  const [uiLanguage, setUiLanguage] = useState<LocaleId>("en");
  const [savedLanguage, setSavedLanguage] = useState<LocaleId>("en");
  const [error, setError] = useState("");
  const [saved, setSaved] = useState(false);
  const [themeSaved, setThemeSaved] = useState(false);
  const [languageSaved, setLanguageSaved] = useState(false);
  const [autoTranslateSaved, setAutoTranslateSavedFlag] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-settings"],
    queryFn: () => api<SystemSettings>("/api/admin/system-settings"),
    enabled: ready,
  });

  useEffect(() => {
    if (data) {
      setAuth(data.auth || {});
      setGit(data.git || {});
      setPlatform(data.platform || {});
      if (data.ui_theme) {
        setUiTheme(data.ui_theme);
        setSavedTheme(data.ui_theme);
        setTheme(data.ui_theme);
      }
      if (data.ui_language) {
        setUiLanguage(data.ui_language);
        setSavedLanguage(data.ui_language);
        setLocale(data.ui_language);
      }
    }
  }, [data, setTheme, setLocale]);

  const save = useMutation({
    mutationFn: () =>
      api<SystemSettings>("/api/admin/system-settings", {
        method: "PUT",
        body: JSON.stringify({ auth, git, platform }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-settings"] });
      setError("");
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("common.failed")),
  });

  const saveTheme = useMutation({
    mutationFn: (themeId: UIThemeId) =>
      api<{ ui_theme: UIThemeId }>("/api/admin/system-settings/ui-theme", {
        method: "PUT",
        body: JSON.stringify({ ui_theme: themeId }),
      }),
    onSuccess: (res) => {
      setSavedTheme(res.ui_theme);
      setTheme(res.ui_theme);
      qc.invalidateQueries({ queryKey: ["admin-settings"] });
      setThemeSaved(true);
      setTimeout(() => setThemeSaved(false), 2000);
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.themeSaveFailed")),
  });

  const saveLanguage = useMutation({
    mutationFn: (localeId: LocaleId) =>
      api<{ ui_language: LocaleId }>("/api/admin/system-settings/ui-language", {
        method: "PUT",
        body: JSON.stringify({ ui_language: localeId }),
      }),
    onSuccess: (res) => {
      setSavedLanguage(res.ui_language);
      setLocale(res.ui_language);
      qc.invalidateQueries({ queryKey: ["admin-settings"] });
      setLanguageSaved(true);
      setTimeout(() => setLanguageSaved(false), 2000);
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.languageSaveFailed")),
  });

  const saveAutoTranslate = useMutation({
    mutationFn: (enabled: boolean) => {
      const nextPlatform = { ...platform, auto_translate_docs: enabled };
      return api<SystemSettings>("/api/admin/system-settings", {
        method: "PUT",
        body: JSON.stringify({ auth, git, platform: nextPlatform }),
      });
    },
    onSuccess: (_res, enabled) => {
      setPlatform((prev) => ({ ...prev, auto_translate_docs: enabled }));
      qc.invalidateQueries({ queryKey: ["admin-settings"] });
      qc.invalidateQueries({ queryKey: ["document"] });
      qc.invalidateQueries({ queryKey: ["book"] });
      setAutoTranslateSavedFlag(true);
      setTimeout(() => setAutoTranslateSavedFlag(false), 2000);
      setError("");
    },
    onError: (e) => setError(e instanceof ApiError ? e.message : t("admin.autoTranslateSaveFailed")),
  });

  function selectTheme(themeId: UIThemeId) {
    setUiTheme(themeId);
    setTheme(themeId);
    if (isSuperAdmin && themeId !== savedTheme) {
      saveTheme.mutate(themeId);
    }
  }

  function selectLanguage(localeId: LocaleId) {
    setUiLanguage(localeId);
    setLocale(localeId);
    if (isSuperAdmin && localeId !== savedLanguage) {
      saveLanguage.mutate(localeId);
    }
  }

  function toggleAutoTranslate(enabled: boolean) {
    setPlatform((prev) => ({ ...prev, auto_translate_docs: enabled }));
    if (isSuperAdmin) {
      saveAutoTranslate.mutate(enabled);
    }
  }

  if (!ready) return null;

  function boolField(
    section: Record<string, unknown>,
    setSection: (v: Record<string, unknown>) => void,
    key: string,
    label: string
  ) {
    return (
      <label className="flex items-center justify-between gap-4 rounded-xl border border-default px-4 py-3">
        <span className="text-sm text-fg-secondary">{label}</span>
        <input
          type="checkbox"
          checked={Boolean(section[key])}
          disabled={!isSuperAdmin}
          onChange={(e) => setSection({ ...section, [key]: e.target.checked })}
          className="checkbox-field"
        />
      </label>
    );
  }

  function textField(
    section: Record<string, unknown>,
    setSection: (v: Record<string, unknown>) => void,
    key: string,
    label: string,
    placeholder?: string
  ) {
    return (
      <label className="block">
        <span className="mb-1 block text-sm text-muted">{label}</span>
        <input
          className="input-field"
          disabled={!isSuperAdmin}
          placeholder={placeholder}
          value={String(section[key] ?? "")}
          onChange={(e) => {
            const raw = e.target.value;
            const num = Number(raw);
            setSection({
              ...section,
              [key]: raw !== "" && !Number.isNaN(num) && /^\d+$/.test(raw) ? num : raw,
            });
          }}
        />
      </label>
    );
  }

  return (
    <FadeIn>
      {isLoading ? (
        <div className="flex justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : (
        <div className="space-y-6">
          {!isSuperAdmin && (
            <p className="text-sm text-warning">{t("admin.readOnly")}</p>
          )}

          <div className="glass p-6">
            <div className="flex items-center gap-2">
              <Languages className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">{t("admin.language")}</h2>
            </div>
            <p className="mt-1 text-sm text-muted">{t("admin.languageHint")}</p>

            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              {UI_LANGUAGES.map((lang) => (
                <button
                  key={lang.id}
                  type="button"
                  disabled={!isSuperAdmin}
                  onClick={() => selectLanguage(lang.id)}
                  className={cn(
                    "theme-preview text-left",
                    uiLanguage === lang.id && "theme-preview-active"
                  )}
                >
                  <p className="font-medium text-fg">{lang.nativeLabel}</p>
                  <p className="mt-1 text-xs text-subtle">{lang.label}</p>
                </button>
              ))}
            </div>

            {saveLanguage.isPending && (
              <p className="mt-3 flex items-center gap-2 text-sm text-muted">
                <Loader2 className="h-4 w-4 animate-spin" />
                {t("admin.savingLanguage")}
              </p>
            )}
            {languageSaved && !saveLanguage.isPending && (
              <p className="mt-3 text-sm text-success-soft">{t("admin.languageSaved")}</p>
            )}
          </div>

          <div className="glass p-6">
            <div className="flex items-center gap-2">
              <Palette className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">{t("admin.appearance")}</h2>
            </div>
            <p className="mt-1 text-sm text-muted">{t("admin.appearanceHint")}</p>

            <div className="mt-4 grid gap-3 sm:grid-cols-2">
              {UI_THEMES.map((theme) => (
                <button
                  key={theme.id}
                  type="button"
                  disabled={!isSuperAdmin}
                  onClick={() => selectTheme(theme.id)}
                  className={cn(
                    "theme-preview text-left",
                    uiTheme === theme.id && "theme-preview-active"
                  )}
                >
                  <p className="font-medium text-fg">{theme.label}</p>
                  <p className="mt-1 text-xs text-subtle">{theme.description}</p>
                </button>
              ))}
            </div>

            {saveTheme.isPending && (
              <p className="mt-3 flex items-center gap-2 text-sm text-muted">
                <Loader2 className="h-4 w-4 animate-spin" />
                {t("admin.savingTheme")}
              </p>
            )}
            {themeSaved && !saveTheme.isPending && (
              <p className="mt-3 text-sm text-success-soft">{t("admin.themeSaved")}</p>
            )}
          </div>

          <div className="glass p-6">
            <div className="flex items-center gap-2">
              <Languages className="h-5 w-5 text-primary" />
              <h2 className="text-lg font-semibold text-fg">{t("admin.autoTranslateDocs")}</h2>
            </div>
            <p className="mt-1 text-sm text-muted">{t("admin.autoTranslateHint")}</p>

            <label className="mt-4 flex items-center justify-between gap-4 rounded-xl border border-default px-4 py-3">
              <span className="text-sm text-fg-secondary">{t("admin.autoTranslateDocs")}</span>
              <input
                type="checkbox"
                className="checkbox-field"
                checked={Boolean(platform.auto_translate_docs)}
                disabled={!isSuperAdmin || saveAutoTranslate.isPending}
                onChange={(e) => toggleAutoTranslate(e.target.checked)}
              />
            </label>

            {saveAutoTranslate.isPending && (
              <p className="mt-3 flex items-center gap-2 text-sm text-muted">
                <Loader2 className="h-4 w-4 animate-spin" />
                {t("admin.savingAutoTranslate")}
              </p>
            )}
            {autoTranslateSaved && !saveAutoTranslate.isPending && (
              <p className="mt-3 text-sm text-success-soft">{t("admin.autoTranslateSaved")}</p>
            )}
          </div>

          <div className="glass p-6">
            <h2 className="text-lg font-semibold text-fg">{t("admin.authentication")}</h2>
            <div className="mt-4 space-y-3">
              {boolField(auth, setAuth, "oidc_enabled", t("admin.enableOidc"))}
              {boolField(auth, setAuth, "local_auth_fallback", t("admin.localAuthFallback"))}
            </div>
          </div>

          <div className="glass p-6">
            <h2 className="text-lg font-semibold text-fg">{t("admin.gitIntegration")}</h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              {textField(git, setGit, "access_token_ref", t("admin.globalTokenRef"), "GIT_ACCESS_TOKEN")}
              {textField(git, setGit, "webhook_secret_ref", t("admin.webhookSecretRef"), "GIT_WEBHOOK_SECRET")}
              {textField(git, setGit, "default_sync_interval_seconds", t("admin.syncInterval"), "300")}
              {textField(git, setGit, "default_sync_mode", t("admin.syncMode"), "scheduled")}
            </div>
          </div>

          <div className="glass p-6">
            <h2 className="text-lg font-semibold text-fg">{t("admin.platform")}</h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              {textField(platform, setPlatform, "search_default_limit", t("admin.searchDefaultLimit"), "20")}
              {textField(platform, setPlatform, "search_max_limit", t("admin.searchMaxLimit"), "100")}
              {boolField(platform, setPlatform, "cache_enabled", t("admin.cacheEnabled"))}
              {textField(platform, setPlatform, "logging_level", t("admin.loggingLevel"), "info")}
            </div>
          </div>

          {error && <p className="text-sm text-danger-soft">{error}</p>}
          {saved && <p className="text-sm text-success-soft">{t("admin.settingsSaved")}</p>}

          {isSuperAdmin && (
            <button type="button" className="btn-primary" disabled={save.isPending} onClick={() => save.mutate()}>
              {save.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : t("admin.saveSettings")}
            </button>
          )}
        </div>
      )}
    </FadeIn>
  );
}
