import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolveProxyTarget } from "./vite.proxy-target.js";

const proxyTarget = resolveProxyTarget();

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  server: {
    host: "0.0.0.0",
    port: 5175,
    proxy: {
      "/api": {
        target: proxyTarget,
        changeOrigin: true
      },
      "/uploads": {
        target: proxyTarget,
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: "dist"
  }
});
