import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

const devServerHost = "localhost";
const backendServerPort = 8080;

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
		host: "0.0.0.0",
		cors: false,
    open: true,
		proxy: {
			"/api": {
				target: `http://${devServerHost}:${backendServerPort}`,
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api/, "")
			}
		}
  },
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "src/setupTests",
    mockReset: true
  },
  define: {
    'process.env': {}
  }
});

