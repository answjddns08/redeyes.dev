import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vueDevTools from "vite-plugin-vue-devtools";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), vueDevTools(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  base: "/",
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ["vue", "vue-router", "pinia"],
          utils: ["axios"],
        },
      },
    },
    outDir: "/var/www/html/blog",
    emptyOutDir: true,
  },
  server: {
    host: "0.0.0.0",
    port: 5173,
    proxy: {
      "/api": {
        target: "https://blog.redeyes.dev",
        changeOrigin: true,
      },
    },
  },
});
