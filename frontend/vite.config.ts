import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // server: {
  //   proxy: {
  //     "/query": "http://localhost:8080",
  //     "/databases": "http://localhost:8080",
  //     "/create": "http://localhost:8080",
  //     "/delete": "http://localhost:8080",
  //     "/connect": "http://localhost:8080",
  //     "/schema": "http://localhost:8080",
  //   },
  // },
  build: {
    outDir: "dist",
  },
});
