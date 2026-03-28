import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";
import { shouldServeH5Index } from "./vite.h5-paths.js";
import { resolveProxyTarget } from "./vite.proxy-target.js";

const proxyTarget = resolveProxyTarget();

function h5HistoryFallback() {
  const rewriteRequest = (req, _res, next) => {
    if (shouldServeH5Index(req.url || "")) {
      req.url = "/m/index.html";
    }
    next();
  };

  return {
    name: "sercherai-h5-history-fallback",
    configureServer(server) {
      server.middlewares.use(rewriteRequest);
    },
    configurePreviewServer(server) {
      server.middlewares.use(rewriteRequest);
    }
  };
}

export default defineConfig({
  base: "/m/",
  plugins: [vue(), h5HistoryFallback()],
  server: {
    host: "0.0.0.0",
    port: 5177,
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
    outDir: "dist-h5",
    rollupOptions: {
      input: path.resolve(__dirname, "m/index.html")
    }
  }
});
