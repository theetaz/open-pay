import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests',
  timeout: 120000,
  expect: { timeout: 10000 },
  use: {
    baseURL: 'http://localhost:3000',
    headless: false,
    launchOptions: {
      slowMo: 800,
    },
    viewport: { width: 1440, height: 900 },
    screenshot: 'on',
    video: 'on',
  },
})
