import { useCallback } from "react";
import { useSplashStore } from "@/lib/splash-store";

/** Scatter form + splash after successful auth submit (call only when credentials are valid). */
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
