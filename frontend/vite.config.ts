import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const proxyAuth = env.VITE_PROXY_AUTH || "http://127.0.0.1:8081";
  const proxyApi = env.VITE_PROXY_API || "http://127.0.0.1:8082";

  return {
    plugins: [react()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      host: "0.0.0.0",
      port: 5173,
      watch: { usePolling: true },
      proxy: {
        "/api/auth": {
          target: proxyAuth,
          changeOrigin: true,
        },
        "/api": {
          target: proxyApi,
          changeOrigin: true,
        },
      },
    },
  };
});
