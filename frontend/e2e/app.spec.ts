import { test, expect } from '@playwright/test'

/**
 * E2E tests for MCP Server Manager
 *
 * 前提: `wails dev` が http://localhost:34115 で起動していること
 * 実行: npm run e2e
 */

test.describe('アプリ基本動作', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('MCP Servers 画面が表示される', async ({ page }) => {
    await expect(page.getByText('MCP Servers')).toBeVisible()
  })

  test('サーバーがない場合の空状態が表示される', async ({ page }) => {
    // DB が空ならこのメッセージが表示される
    const emptyMsg = page.getByText('No servers configured')
    const serverList = page.locator('[class*="space-y-2"] > div')
    const isEmpty = (await serverList.count()) === 0
    if (isEmpty) {
      await expect(emptyMsg).toBeVisible()
    }
  })

  test('+ Add Server ボタンが存在する', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Add Server/ })).toBeVisible()
  })

  test('Import Claude Desktop ボタンが存在する', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Import Claude Desktop/ })).toBeVisible()
  })
})

test.describe('サーバー追加フォーム', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('+ Add Server をクリックするとフォームが開く', async ({ page }) => {
    await page.getByRole('button', { name: /Add Server/ }).click()
    await expect(page.getByText('New Server')).toBeVisible()
    await expect(page.getByPlaceholder('my-mcp-server')).toBeVisible()
  })

  test('Cancel でフォームが閉じる', async ({ page }) => {
    await page.getByRole('button', { name: /Add Server/ }).click()
    await page.getByRole('button', { name: 'Cancel' }).click()
    await expect(page.getByText('New Server')).not.toBeVisible()
  })

  test('transport を stdio → HTTP に切り替えると URL フィールドが表示される', async ({ page }) => {
    await page.getByRole('button', { name: /Add Server/ }).click()
    await page.selectOption('select', 'http')
    await expect(page.getByPlaceholder('http://localhost:8080')).toBeVisible()
    await expect(page.getByPlaceholder('npx')).not.toBeVisible()
  })

  test('stdio サーバーを追加するとリストに表示される', async ({ page }) => {
    const serverName = `e2e-test-${Date.now()}`
    await page.getByRole('button', { name: /Add Server/ }).click()
    await page.getByPlaceholder('my-mcp-server').fill(serverName)
    await page.getByPlaceholder('npx').fill('echo')
    await page.getByRole('button', { name: 'Add', exact: true }).click()
    await expect(page.getByText(serverName)).toBeVisible({ timeout: 5000 })
    // クリーンアップ: 追加したサーバーを削除
    const deleteBtn = page.locator(`text=${serverName}`).locator('../..').getByRole('button', { name: 'Delete' })
    await deleteBtn.click()
    page.on('dialog', d => d.accept())
  })
})

test.describe('ナビゲーション', () => {
  test('/ が /servers にリダイレクトされる', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/servers/)
  })
})
