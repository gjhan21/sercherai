import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

function resolveProxyTarget() {
  const explicit = process.env.VITE_PROXY_TARGET;
  if (explicit && typeof explicit === "string" && explicit.trim()) return explicit.trim();
  return "http://127.0.0.1:18080";
}

const proxyTarget = resolveProxyTarget();

export default defineConfig({
  base: "/m/",
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  server: {
    host: "0.0.0.0",
    port: 5277,
    proxy: {
      "/api": {
        target: proxyTarget,
        changeOrigin: true
      },
      "/uploads": {
        target: "http://127.0.0.1:18080",
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: "dist-h5",
    rollupOptions: {
      input: fileURLToPath(new URL("./m/index.html", import.meta.url))
    }
  }
});
