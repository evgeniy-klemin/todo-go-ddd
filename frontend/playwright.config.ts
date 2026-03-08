import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  retries: 0,
  use: {
    baseURL: 'http://localhost:15173',
    headless: true,
  },
  webServer: [
    {
      command: 'cd .. && rm -f todotest.db && go run cmd/todoserver/todoserver.go -port 18080',
      url: 'http://localhost:18080/items',
      reuseExistingServer: false,
      timeout: 60000,
    },
    {
      command: 'npx vite --port 15173',
      url: 'http://localhost:15173',
      reuseExistingServer: false,
      timeout: 30000,
      env: {
        VITE_API_PORT: '18080',
      },
    },
  ],
});
