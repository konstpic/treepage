import { getRuntimeConfig, getAuthUrlSync } from "@/lib/config";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

let _refreshPromise: Promise<boolean> | null = null;

async function doRefresh(): Promise<boolean> {
  const refreshToken = localStorage.getItem("refresh_token");
  if (!refreshToken) return false;
  const authUrl = getAuthUrlSync();
  try {
    const res = await fetch(`${authUrl}/api/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (res.ok) {
      const data = await res.json();
      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      return true;
    }
  } catch { /* network */ }
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  return false;
}

function refreshTokens(): Promise<boolean> {
  if (!_refreshPromise) {
    _refreshPromise = doRefresh().finally(() => { _refreshPromise = null; });
  }
  return _refreshPromise;
}

async function requestJson<T>(baseUrl: string, path: string, init: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem("access_token");
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(init.headers || {}),
  };
  const res = await fetch(`${baseUrl}${path}`, { ...init, headers });

  if (res.status === 401) {
    if (localStorage.getItem("refresh_token")) {
      const refreshed = await refreshTokens();
      if (refreshed) {
        const newToken = localStorage.getItem("access_token");
        const retry = await fetch(`${baseUrl}${path}`, {
          ...init,
          headers: {
            "Content-Type": "application/json",
            ...(newToken ? { Authorization: `Bearer ${newToken}` } : {}),
            ...(init.headers || {}),
          },
        });
        if (retry.ok) return retry.json();
        throw new ApiError("Session expired", 401);
      }
    }
    throw new ApiError("Session expired", 401);
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const msg = (body as { error?: string }).error || `Request failed (${res.status})`;
    throw new ApiError(msg, res.status);
  }
  return res.json();
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const cfg = await getRuntimeConfig();
  return requestJson<T>(cfg.apiUrl, path, init);
}

export async function authApi<T>(path: string, init?: RequestInit): Promise<T> {
  const cfg = await getRuntimeConfig();
  return requestJson<T>(cfg.authUrl, path, init);
}

export async function getLoginUrl(): Promise<string> {
  const cfg = await getRuntimeConfig();
  return `${cfg.authUrl}/api/auth/login`;
}

export function getPublicBranding() {
  return publicApi<{ project_name: string; project_code: string }>("/api/public/branding");
}

export function getPublicSpaces() {
  return publicApi<{ items: { id: string; slug: string; name: string; description?: string; is_public: boolean }[] }>(
    "/api/public/spaces"
  );
}

async function publicApi<T>(path: string, init?: RequestInit): Promise<T> {
  const cfg = await getRuntimeConfig();
  const res = await fetch(`${cfg.apiUrl}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError((body as { error?: string }).error || `Request failed (${res.status})`, res.status);
  }
  return res.json();
}

/** API that works with or without auth (sends token when present). */
export async function optionalAuthApi<T>(path: string, init?: RequestInit): Promise<T> {
  const cfg = await getRuntimeConfig();
  return requestJson<T>(cfg.apiUrl, path, init);
}

export async function loginLocal(email: string, password: string) {
  const cfg = await getRuntimeConfig();
  const res = await fetch(`${cfg.authUrl}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError((body as { error?: string }).error || "Login failed", res.status);
  }
  return res.json() as Promise<{
    access_token: string;
    refresh_token: string;
    user?: { id: string; email: string; display_name: string; roles: string[] };
  }>;
}

/** Call once at app startup (react-spa serves /config.json in production). */
export function initApiConfig(): Promise<void> {
  return getRuntimeConfig().then(() => undefined);
}
