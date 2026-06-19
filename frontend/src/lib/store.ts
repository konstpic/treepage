import { create } from "zustand";
import { getAuthUrlSync } from "@/lib/config";

export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  roles: string[];
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isHydrated: boolean;
  setAuth: (accessToken: string, refreshToken: string) => void;
  setUser: (user: User | null) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isHydrated: false,

  setAuth: (accessToken, refreshToken) => {
    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("refresh_token", refreshToken);
    set({ isAuthenticated: true });
  },

  setUser: (user) => set({ user }),

  logout: () => {
    const rt = localStorage.getItem("refresh_token");
    const at = localStorage.getItem("access_token");
    const authUrl = getAuthUrlSync();
    if (rt || at) {
      fetch(`${authUrl}/api/auth/logout`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(at ? { Authorization: `Bearer ${at}` } : {}),
        },
        body: JSON.stringify({ refresh_token: rt }),
      }).catch(() => {});
    }
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    set({ user: null, isAuthenticated: false });
  },

  hydrate: () => {
    const token = localStorage.getItem("access_token");
    set({ isAuthenticated: !!token, isHydrated: true });
  },
}));

interface BrandingState {
  projectName: string;
  setProjectName: (name: string) => void;
}

export const useBrandingStore = create<BrandingState>((set) => ({
  projectName: "TreePage",
  setProjectName: (name) => set({ projectName: name }),
}));
