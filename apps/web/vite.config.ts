import { defineConfig, type UserConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

// https://vitejs.dev/config/
export default defineConfig((): UserConfig => {
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
      // 'hidden' ships sourcemaps into the build output but does not
      // emit a `//# sourceMappingURL=` comment, so the public bundle
      // does not reference the map. Operators can still recover stack
      // traces server-side (e.g. by uploading the hidden maps to an
      // error monitor) without exposing them to end users. Dev and
      // preview builds get full sourcemaps for local debugging.
      sourcemap: process.env.NODE_ENV === 'production' ? 'hidden' : true,
    },
  };
});
