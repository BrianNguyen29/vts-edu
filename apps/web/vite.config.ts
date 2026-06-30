import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

// https://vitejs.dev/config/
export default defineConfig(() => {
  return {
    plugins: [react()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      port: 5173,
      proxy: {
        // Dev proxy points to local Go API. In production the SPA uses the absolute VITE_API_BASE_URL directly.
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: false,
          secure: false,
        },
        '/app-config.json': {
          target: 'http://localhost:8080',
          changeOrigin: false,
          secure: false,
        },
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: true,
    },
  };
});
