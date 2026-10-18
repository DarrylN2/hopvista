import { defineConfig } from "vite";

export default defineConfig({
  base: "/",
  build: {
    outDir: "../internal/localapi/static",
    emptyOutDir: false,
    assetsDir: "",
    cssCodeSplit: false,
    rollupOptions: {
      output: {
        entryFileNames: "app.js",
        assetFileNames: (asset) => asset.names.some((name) => name.endsWith(".css")) ? "style.css" : "[name][extname]",
      },
    },
  },
});
