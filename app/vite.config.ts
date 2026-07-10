import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from "path"
import { createHtmlPlugin } from 'vite-plugin-html' //@ts-ignore
import { createTranslationPlugin } from "./src/translator"

const backendTarget = "http://localhost:8094";

const backendProxy = {
  target: backendTarget,
  changeOrigin: true,
};

const apiRewriteProxy = {
  ...backendProxy,
  rewrite: (path: string) => `/api${path}`,
};

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    createHtmlPlugin({
      minify: true,
    }),
    createTranslationPlugin(),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  css: {
    preprocessorOptions: {
      less: {
        javascriptEnabled: true,
      }
    }
  },
  build: {
    manifest: true,
    chunkSizeWarningLimit: 2048,
    rollupOptions: {
      output: {
        entryFileNames: `assets/[name].[hash].js`,
        chunkFileNames: `assets/[name].[hash].js`,
      },
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: false,
    hmr: {
      overlay: true,
    },
    proxy: {
      "/api": {
        ...backendProxy,
        ws: true,
      },
      "/v1": {
        ...backendProxy,
      },
      "/login": apiRewriteProxy,
      "/state": apiRewriteProxy,
      "/register": apiRewriteProxy,
      "/verify": apiRewriteProxy,
      "/reset": apiRewriteProxy,
      "/userinfo": apiRewriteProxy,
      "/quota": apiRewriteProxy,
      "/broadcast": apiRewriteProxy,
      "/conversation": apiRewriteProxy,
    }
  }
});
