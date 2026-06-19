import { create } from "zustand";

export type SplashPhase = "idle" | "scatter" | "splash";
export type ScatterMode = "auth-form" | null;

/** How long form pieces fly off-screen. */
export const SCATTER_ANIM_MS = 1200;
/** Blank pause after scatter — splash starts only after this. */
export const SCATTER_PAUSE_MS = 550;

interface SplashState {
  phase: SplashPhase;
  scatterMode: ScatterMode;
  pendingAction: (() => void) | null;
  splashKey: number;
  requestAuthFormSplash: (action: () => void) => void;
  requestInitialSplash: () => void;
  finishSplash: () => void;
}

let scatterTimer: ReturnType<typeof setTimeout> | null = null;

export const useSplashStore = create<SplashState>((set, get) => ({
  phase: "idle",
  scatterMode: null,
  pendingAction: null,
  splashKey: 0,

  requestAuthFormSplash: (action) => {
    if (scatterTimer) clearTimeout(scatterTimer);
    set((s) => ({
      phase: "scatter",
      scatterMode: "auth-form",
      pendingAction: action,
      splashKey: s.splashKey + 1,
    }));
    scatterTimer = setTimeout(() => {
      if (get().phase === "scatter") {
        set({ phase: "splash", scatterMode: null });
      }
    }, SCATTER_ANIM_MS + SCATTER_PAUSE_MS);
  },

  requestInitialSplash: () => {
    set((s) => ({
      phase: "splash",
      scatterMode: null,
      pendingAction: null,
      splashKey: s.splashKey + 1,
    }));
  },

  finishSplash: () => {
    const action = get().pendingAction;
    set({ phase: "idle", scatterMode: null, pendingAction: null });
    action?.();
  },
}));

const INITIAL_SPLASH_KEY = "treepage_initial_splash";

export function shouldShowInitialSplash(pathname: string): boolean {
  if (typeof sessionStorage === "undefined") return false;
  if (sessionStorage.getItem(INITIAL_SPLASH_KEY) === "1") return false;
  if (pathname.startsWith("/admin")) return false;
  if (pathname.startsWith("/auth/callback")) return false;
  return true;
}

export function markInitialSplashSeen(): void {
  sessionStorage.setItem(INITIAL_SPLASH_KEY, "1");
}
