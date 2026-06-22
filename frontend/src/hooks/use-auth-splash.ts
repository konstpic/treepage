import { useCallback } from "react";
import { useSplashStore } from "@/lib/splash-store";

/** Scatter form + splash after successful auth submit (call only when credentials are valid). */
export function useAuthFormSplash() {
  const requestAuthFormSplash = useSplashStore((s) => s.requestAuthFormSplash);

  const startAuthFormSplash = useCallback(
    (action: () => void, welcomeName?: string) => {
      requestAuthFormSplash(action, welcomeName);
    },
    [requestAuthFormSplash],
  );

  return { startAuthFormSplash };
}

/** Splash with welcome text (OIDC callback — no form scatter). */
export function useWelcomeSplash() {
  const requestWelcomeSplash = useSplashStore((s) => s.requestWelcomeSplash);

  const startWelcomeSplash = useCallback(
    (action: () => void, welcomeName: string) => {
      requestWelcomeSplash(action, welcomeName);
    },
    [requestWelcomeSplash],
  );

  return { startWelcomeSplash };
}
