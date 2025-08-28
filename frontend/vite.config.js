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
    // 청크 크기 최적화
    rollupOptions: {
      output: {
        manualChunks: {
          // 벤더 라이브러리 분리
          vendor: ['vue', 'vue-router', 'pinia'],
          icons: ['@fortawesome/fontawesome-svg-core', '@fortawesome/vue-fontawesome'],
          utils: ['axios', 'marked']
        }
      }
    },
    outDir: "/var/www/html/blog",
    emptyOutDir: true,
    // 청크 크기 경고 임계값 증가
    chunkSizeWarningLimit: 1000,
    // 소스맵 비활성화 (프로덕션 빌드 크기 줄이기)
    sourcemap: false,
    // 최소화 옵션
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 콘솔 로그 제거
        drop_debugger: true // 디버거 구문 제거
      }
    }
  },
  // 개발 서버 최적화
  server: {
    fs: {
      strict: false
    }
  }
});
