import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";
import { shouldServeH5Index } from "./vite.h5-paths.js";

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
        target: process.env.VITE_PROXY_TARGET || "http://127.0.0.1:18080",
        changeOrigin: true
      },
      "/uploads": {
        target: process.env.VITE_PROXY_TARGET || "http://127.0.0.1:18080",
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
