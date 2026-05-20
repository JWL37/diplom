import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

const orchestratorUrl = process.env.ORCHESTRATOR_URL ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: orchestratorUrl,
        changeOrigin: true
      },
      "/ping": {
        target: orchestratorUrl,
        changeOrigin: true
      }
    }
  }
});
