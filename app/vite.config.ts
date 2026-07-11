import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from "path"
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
export default defineConfig(async ({ command }) => {
  // vite-plugin-html pulls in the HTML minifier and template engine. They are
  // only needed for production builds, so keep them off the dev-server startup
  // path. This also leaves Vite's native SPA fallback in charge during dev.
  const htmlPlugins = command === "build"
    ? (await import("vite-plugin-html")).createHtmlPlugin({ minify: true })
    : [];

  return {
    plugins: [
      react(),
      ...htmlPlugins,
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
        },
      },
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
      host: "0.0.0.0",
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
      },
    },
  };
});
