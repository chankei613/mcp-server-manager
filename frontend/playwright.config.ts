import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  timeout: 30_000,
  use: {
    // wails dev はデフォルトで http://localhost:34115 を使う
    baseURL: 'http://localhost:34115',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  // wails dev が別プロセスで起動済みであることを前提とする
  // webServer は不要（Wails がフロントを serve する）
})
