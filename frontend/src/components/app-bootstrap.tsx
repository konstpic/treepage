import { useEffect, useState, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { AnimatePresence } from "framer-motion";
import { SplashScreen } from "@/components/splash-screen";
import { TREE_LOGO_DRAW_MS } from "@/components/treepage-logo";
import { useAuthStore } from "@/lib/store";
import { useThemeStore } from "@/lib/theme-store";
import { useLocaleStore } from "@/lib/locale-store";
import {
  useSplashStore,
  shouldShowInitialSplash,
  markInitialSplashSeen,
} from "@/lib/splash-store";

const MIN_SPLASH_MS = TREE_LOGO_DRAW_MS + 350;

export function AppBootstrap({ children }: { children: ReactNode }) {
  const location = useLocation();
  const phase = useSplashStore((s) => s.phase);
  const splashKey = useSplashStore((s) => s.splashKey);
  const requestInitialSplash = useSplashStore((s) => s.requestInitialSplash);
  const finishSplash = useSplashStore((s) => s.finishSplash);

  const isHydrated = useAuthStore((s) => s.isHydrated);
  const themeLoaded = useThemeStore((s) => s.isLoaded);
  const localeLoaded = useLocaleStore((s) => s.isLoaded);

  const [animDone, setAnimDone] = useState(false);
  const [initialChecked, setInitialChecked] = useState(false);

  const appReady = isHydrated && themeLoaded && localeLoaded;
  const showSplash = phase === "splash";

  useEffect(() => {
    if (!appReady || initialChecked) return;
    setInitialChecked(true);
    if (shouldShowInitialSplash(location.pathname)) {
      requestInitialSplash();
    }
  }, [appReady, initialChecked, location.pathname, requestInitialSplash]);

  useEffect(() => {
    if (phase !== "splash") {
      setAnimDone(false);
      return;
    }
    setAnimDone(false);
    const timer = window.setTimeout(() => setAnimDone(true), MIN_SPLASH_MS);
    return () => window.clearTimeout(timer);
  }, [phase, splashKey]);

  useEffect(() => {
    if (!showSplash || !animDone || !appReady) return;
    markInitialSplashSeen();
    finishSplash();
  }, [showSplash, animDone, appReady, finishSplash]);

  useEffect(() => {
    if (phase === "scatter" || phase === "splash") {
      document.body.style.overflow = "hidden";
      return () => {
        document.body.style.overflow = "";
      };
    }
  }, [phase]);

  return (
    <>
      {children}
      <AnimatePresence>
        {showSplash && <SplashScreen key={splashKey} />}
      </AnimatePresence>
    </>
  );
}
