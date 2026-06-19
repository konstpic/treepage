/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
  readonly VITE_AUTH_URL: string;
  readonly VITE_USE_PROXY: string;
  readonly VITE_DEV_LOGIN: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
