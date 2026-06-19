export interface RuntimeConfig {
  apiUrl: string;
  authUrl: string;
}

let runtimeConfig: RuntimeConfig | null = null;
let configPromise: Promise<RuntimeConfig> | null = null;

/** Same-origin API (empty base) — Vite/nginx proxy forwards /api/* to backends. */
export const SAME_ORIGIN_CONFIG: RuntimeConfig = { apiUrl: "", authUrl: "" };

export async function getRuntimeConfig(): Promise<RuntimeConfig> {
  if (runtimeConfig) return runtimeConfig;
  if (!configPromise) {
    configPromise = loadRuntimeConfig();
  }
  runtimeConfig = await configPromise;
  return runtimeConfig;
}

async function loadRuntimeConfig(): Promise<RuntimeConfig> {
  const useProxy = import.meta.env.VITE_USE_PROXY === "true";
  const devApi = import.meta.env.VITE_API_URL || "";
  const devAuth = import.meta.env.VITE_AUTH_URL || devApi;

  if (import.meta.env.DEV) {
    if (useProxy || !devApi) {
      return SAME_ORIGIN_CONFIG;
    }
    return { apiUrl: devApi, authUrl: devAuth };
  }

  try {
    const res = await fetch("/config.json");
    if (res.ok) {
      const data = (await res.json()) as RuntimeConfig;
      const apiUrl = data.apiUrl ?? devApi;
      const authUrl = data.authUrl ?? devAuth;
      if (apiUrl || authUrl) {
        return { apiUrl, authUrl };
      }
      return SAME_ORIGIN_CONFIG;
    }
  } catch {
    // fall through to build-time defaults
  }

  if (devApi) {
    return { apiUrl: devApi, authUrl: devAuth };
  }
  return SAME_ORIGIN_CONFIG;
}

export function getApiUrlSync(): string {
  return runtimeConfig?.apiUrl ?? import.meta.env.VITE_API_URL ?? "";
}

export function getAuthUrlSync(): string {
  return runtimeConfig?.authUrl ?? import.meta.env.VITE_AUTH_URL ?? getApiUrlSync();
}
