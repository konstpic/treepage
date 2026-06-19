import { useEffect, type ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore, useBrandingStore } from "@/lib/store";
import { authApi, initApiConfig, getPublicBranding } from "@/lib/api";
import { fetchPublicUITheme, readCachedUITheme } from "@/lib/theme";
import { fetchPublicUILocale, readCachedLocale } from "@/lib/locale";
import { getRuntimeConfig } from "@/lib/config";
import { useThemeStore } from "@/lib/theme-store";
import { useLocaleStore } from "@/lib/locale-store";

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 30_000, retry: 1 } },
});

export function Providers({ children }: { children: ReactNode }) {
  const hydrate = useAuthStore((s) => s.hydrate);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const isHydrated = useAuthStore((s) => s.isHydrated);
  const setUser = useAuthStore((s) => s.setUser);
  const logout = useAuthStore((s) => s.logout);
  const setProjectName = useBrandingStore((s) => s.setProjectName);
  const setTheme = useThemeStore((s) => s.setTheme);
  const setThemeLoaded = useThemeStore((s) => s.setLoaded);
  const setLocale = useLocaleStore((s) => s.setLocale);
  const setLocaleLoaded = useLocaleStore((s) => s.setLoaded);

  useEffect(() => { hydrate(); }, [hydrate]);

  useEffect(() => {
    let cancelled = false;

    initApiConfig()
      .then(() => getRuntimeConfig())
      .then(async (cfg) => {
        const [themeId, localeId] = await Promise.all([
          fetchPublicUITheme(cfg.apiUrl).catch(() => readCachedUITheme()),
          fetchPublicUILocale(cfg.apiUrl).catch(() => readCachedLocale()),
        ]);
        if (!cancelled) {
          setTheme(themeId);
          setThemeLoaded(true);
          setLocale(localeId);
          setLocaleLoaded(true);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setTheme(readCachedUITheme());
          setThemeLoaded(true);
          setLocale(readCachedLocale());
          setLocaleLoaded(true);
        }
      });

    initApiConfig()
      .then(() => getPublicBranding())
      .then((b) => {
        if (!cancelled) {
          setProjectName(b.project_name);
          document.title = b.project_name;
        }
      })
      .catch(() => {});

    return () => { cancelled = true; };
  }, [setProjectName, setTheme, setThemeLoaded, setLocale, setLocaleLoaded]);

  useEffect(() => {
    if (!isHydrated || !isAuthenticated) return;
    authApi<{ id: string; email: string; display_name: string; avatar_url?: string; roles: string[] }>("/api/auth/me")
      .then(setUser)
      .catch(() => logout());
  }, [isHydrated, isAuthenticated, setUser, logout]);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
