import { create } from "zustand";

const STORAGE_KEY = "treepage_onboarding_v1_done";

interface OnboardingState {
  active: boolean;
  step: number;
  start: () => void;
  next: () => void;
  back: () => void;
  skip: () => void;
  finish: () => void;
  restart: () => void;
  shouldAutoStart: () => boolean;
}

export const useOnboardingStore = create<OnboardingState>((set) => ({
  active: false,
  step: 0,

  shouldAutoStart: () => {
    if (typeof localStorage === "undefined") return false;
    return localStorage.getItem(STORAGE_KEY) !== "1";
  },

  start: () => set({ active: true, step: 0 }),

  next: () => set((s) => ({ step: s.step + 1 })),

  back: () => set((s) => ({ step: Math.max(0, s.step - 1) })),

  skip: () => {
    localStorage.setItem(STORAGE_KEY, "1");
    set({ active: false, step: 0 });
  },

  finish: () => {
    localStorage.setItem(STORAGE_KEY, "1");
    set({ active: false, step: 0 });
  },

  restart: () => {
    localStorage.removeItem(STORAGE_KEY);
    set({ active: true, step: 0 });
  },
}));
