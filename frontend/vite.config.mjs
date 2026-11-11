import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Le Dockerfile et env.sh servent le dossier `build/` (serve -s build) : on garde
// donc cette sortie plutôt que le `dist` par défaut de Vite.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'build',
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
