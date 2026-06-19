import { create } from "zustand";
import {
  applyUITheme,
  DEFAULT_UI_THEME,
  readCachedUITheme,
  type UIThemeId,
} from "@/lib/theme";

interface ThemeState {
  themeId: UIThemeId;
  isLoaded: boolean;
  setTheme: (themeId: UIThemeId) => void;
  setLoaded: (loaded: boolean) => void;
}

export const useThemeStore = create<ThemeState>((set) => ({
  themeId: readCachedUITheme(),
  isLoaded: false,
  setTheme: (themeId) => {
    applyUITheme(themeId);
    set({ themeId });
  },
  setLoaded: (isLoaded) => set({ isLoaded }),
}));

export { DEFAULT_UI_THEME };
