import { create } from "zustand";
import {
  applyUILocale,
  DEFAULT_LOCALE,
  readCachedLocale,
  type LocaleId,
} from "@/lib/locale";

interface LocaleState {
  localeId: LocaleId;
  isLoaded: boolean;
  setLocale: (localeId: LocaleId) => void;
  setLoaded: (loaded: boolean) => void;
}

export const useLocaleStore = create<LocaleState>((set) => ({
  localeId: readCachedLocale(),
  isLoaded: false,
  setLocale: (localeId) => {
    applyUILocale(localeId);
    set({ localeId });
  },
  setLoaded: (isLoaded) => set({ isLoaded }),
}));

export { DEFAULT_LOCALE };
