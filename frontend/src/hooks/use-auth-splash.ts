import { useCallback } from "react";
import { useSplashStore } from "@/lib/splash-store";

/** Scatter form + splash — only for auth page «Sign in» submit. */
export function useAuthFormSplash() {
  const requestAuthFormSplash = useSplashStore((s) => s.requestAuthFormSplash);

  const startAuthFormSplash = useCallback(
    (action: () => void) => {
      requestAuthFormSplash(action);
    },
    [requestAuthFormSplash],
  );

  return { startAuthFormSplash };
}
