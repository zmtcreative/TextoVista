import { test, expect } from '@playwright/test';
import WailsDevHelper from './wails-dev-helper';
import { Page } from '@playwright/test';
import { emitRuntimeEvent, expectTheme } from './runtime-test-helpers';
import { promises as fs } from 'node:fs';
import path from 'node:path';

test.describe('Fast Sequential Tests - Single App Instance', () => {
  let wailsDev: WailsDevHelper;
  let page: Page;

  // Start application ONCE before all tests
  test.beforeAll(async () => {
    console.log('🚀 Starting Wails application once for all tests...');
    wailsDev = new WailsDevHelper({ runtimeMode: 'headless' });
    page = await wailsDev.launchAndConnect();
    await wailsDev.waitForAppReady(page);
    console.log('✅ Application ready - will be shared across all tests');
  });

  // Shutdown application ONCE after all tests
  test.afterAll(async () => {
    console.log('🔚 Shutting down shared application...');
    if (wailsDev) {
      await wailsDev.disconnect();
    }
    console.log('✅ Shared application shutdown complete');
  });

  // Between tests: just reset state (no restart needed!)
  test.beforeEach(async () => {
    console.log('🔄 Resetting application state (fast)...');

    // Quick state reset - much faster than app restart
    try {
      // Close any open modals
      await page.keyboard.press('Escape');
      await page.waitForTimeout(200);

      // Navigate back to fresh state by reloading the page
      await page.reload({ waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(500);

      // Wait for app to be ready
      await page.waitForSelector('.app-header', { timeout: 5000 });

      console.log('✅ State reset complete (no app restart needed)');
    } catch (error) {
      console.log(`⚠️ State reset error (continuing anyway): ${error}`);
    }
  });

  // Test 1: Basic application functionality
  test('should display application content correctly', async () => {
    console.log('🧪 Running Test 1: Basic app functionality');

    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('.content-area')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();
    await expect(page.locator('#content')).toContainText('No markdown file specified');
    await expectTheme(page, 'light');

    await page.screenshot({ path: 'test-results/fast-01-app-content.png' });
    console.log('✅ Test 1 PASSED - Basic functionality verified');
  });

  // Test 2: Help modal (no app restart needed!)
  test('should handle Help modal functionality', async () => {
    console.log('🧪 Running Test 2: Help modal');

    // Verify Help modal elements exist
    await expect(page.locator('#help-modal-overlay')).toBeAttached();
    await expect(page.locator('#help-modal-content')).toBeAttached();
    await expect(page.locator('#help-modal-overlay')).toBeHidden();

    await emitRuntimeEvent(page, 'show-help', 'About', '<div><h3>TextoVista</h3><p>Fast testing!</p></div>');

    await expect(page.locator('#help-modal-overlay')).toBeVisible();
    await expect(page.locator('#help-modal-text')).toContainText('Fast testing!');

    await page.click('#help-modal-close');
    await expect(page.locator('#help-modal-overlay')).toBeHidden();

    await page.screenshot({ path: 'test-results/fast-02-help-modal.png' });
    console.log('✅ Test 2 PASSED - Help modal functionality verified');
  });

  // Test 3: File operations (still using same app instance!)
  test('should handle file operations', async () => {
    console.log('🧪 Running Test 3: File operations');

    // Test keyboard shortcuts
    await page.keyboard.press('Control+o');
    await page.waitForTimeout(300);

    // Verify app is still responsive
    await expect(page.locator('#content')).toBeVisible();
    await expect(page.locator('.app-header')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-03-file-operations.png' });
    console.log('✅ Test 3 PASSED - File operations handled correctly');
  });

  // Test 4: Theme functionality (app still running!)
  test('should handle theme operations', async () => {
    console.log('🧪 Running Test 4: Theme operations');

    const themeButton = page.locator('.theme-toggle-btn');

    await expect(themeButton).toBeVisible();
    await expectTheme(page, 'light');

    await themeButton.click();
    await expectTheme(page, 'dark');

    await page.screenshot({ path: 'test-results/fast-04-theme-operations.png' });
    console.log('✅ Test 4 PASSED - Theme operations handled correctly');
  });

  // Test 5: Multiple rapid interactions (same app instance!)
  test('should handle rapid user interactions', async () => {
    console.log('🧪 Running Test 5: Rapid interactions');

    // Rapid fire multiple interactions
    await page.keyboard.press('Control+o');  // File open
    await page.waitForTimeout(100);

    await page.keyboard.press('Control+p');  // Print
    await page.waitForTimeout(100);

    await page.keyboard.press('F5');         // Refresh
    await page.waitForTimeout(300);

    // Verify app handled all interactions gracefully
    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('#content')).toBeVisible();

    await page.screenshot({ path: 'test-results/fast-05-rapid-interactions.png' });
    console.log('✅ Test 5 PASSED - Rapid interactions handled correctly');
  });

  // Performance verification test
  test('should demonstrate the performance improvement', async () => {
    console.log('🧪 Running Performance Test: App reuse verification');

    const startTime = Date.now();

    // Verify we're still using the same application instance
    await expect(page.locator('.app-header')).toBeVisible();
    await expect(page.locator('.content-area')).toBeVisible();

    // Quick interactions to verify responsiveness
    await page.keyboard.press('Control+o');
    await page.waitForTimeout(200);
    await page.keyboard.press('Escape');

    const endTime = Date.now();
    const testTime = endTime - startTime;

    console.log(`⚡ Performance test completed in ${testTime}ms`);
    console.log('🎯 This test ran WITHOUT restarting the application!');

    await page.screenshot({ path: 'test-results/fast-06-performance-demo.png' });
    console.log('✅ Performance Test PASSED - Application reuse verified');
  });

  test('should manually refresh markdown content after tmp file edit', async () => {
    console.log('🧪 Running Manual Refresh E2E against tmp file');

    const tmpDir = path.resolve(process.cwd(), '..', 'tmp');
    const tmpFilePath = path.join(tmpDir, 'playwright-manual-refresh.md');
    const initialContent = '# Manual Refresh E2E\n\nInitial marker: before refresh';
    const updatedContent = '# Manual Refresh E2E\n\nUpdated marker: after refresh';

    await fs.mkdir(tmpDir, { recursive: true });
    await fs.writeFile(tmpFilePath, initialContent, 'utf8');

    try {
      // Keep this test deterministic by disabling auto-refresh for the current session.
      await page.evaluate(async () => {
        const appApi = (window as any)?.go?.main?.App;
        if (!appApi?.GetSettings || !appApi?.SaveSettingsSessionOnly) {
          throw new Error('Go App settings APIs are unavailable');
        }

        const currentSettings = await appApi.GetSettings();
        currentSettings.application.use_auto_refresh = false;
        await appApi.SaveSettingsSessionOnly(currentSettings);
      });

      await page.evaluate(async (filePath: string) => {
        const appApi = (window as any)?.go?.main?.App;
        if (!appApi?.LoadAndDisplayMarkdown) {
          throw new Error('Go App LoadAndDisplayMarkdown API is unavailable');
        }
        await appApi.LoadAndDisplayMarkdown(filePath);
      }, tmpFilePath);

      await expect(page.locator('#content')).toContainText('Initial marker: before refresh');

      await fs.writeFile(tmpFilePath, updatedContent, 'utf8');

      await page.click('.refresh-btn');
      await expect(page.locator('#content')).toContainText('Updated marker: after refresh');
      await expect(page.locator('#content')).not.toContainText('Initial marker: before refresh');

      await page.screenshot({ path: 'test-results/fast-07-manual-refresh-tmp-file.png' });
      console.log('✅ Manual Refresh E2E PASSED - content updated after toolbar refresh');
    } finally {
      await fs.unlink(tmpFilePath).catch(() => {
        // Keep cleanup best-effort; tmp folder is intentionally unmanaged.
      });
    }
  });

  test('should not update content while auto-refresh is disabled until Refresh is clicked', async () => {
    console.log('🧪 Running Disabled Auto-Refresh E2E against tmp file');

    const tmpDir = path.resolve(process.cwd(), '..', 'tmp');
    const tmpFilePath = path.join(tmpDir, 'playwright-disabled-auto-refresh.md');
    const initialContent = '# Disabled Auto Refresh E2E\n\nStable marker: before manual refresh';
    const updatedContent = '# Disabled Auto Refresh E2E\n\nChanged marker: after disk edit';

    await fs.mkdir(tmpDir, { recursive: true });
    await fs.writeFile(tmpFilePath, initialContent, 'utf8');

    try {
      await page.evaluate(async () => {
        const appApi = (window as any)?.go?.main?.App;
        if (!appApi?.GetSettings || !appApi?.SaveSettingsSessionOnly) {
          throw new Error('Go App settings APIs are unavailable');
        }

        const currentSettings = await appApi.GetSettings();
        currentSettings.application.use_auto_refresh = false;
        await appApi.SaveSettingsSessionOnly(currentSettings);
      });

      await page.evaluate(async (filePath: string) => {
        const appApi = (window as any)?.go?.main?.App;
        if (!appApi?.LoadAndDisplayMarkdown) {
          throw new Error('Go App LoadAndDisplayMarkdown API is unavailable');
        }
        await appApi.LoadAndDisplayMarkdown(filePath);
      }, tmpFilePath);

      await expect(page.locator('#content')).toContainText('Stable marker: before manual refresh');

      await fs.writeFile(tmpFilePath, updatedContent, 'utf8');

      // Wait briefly to allow watcher events to occur; content should remain unchanged while auto-refresh is disabled.
      await page.waitForTimeout(800);
      await expect(page.locator('#content')).toContainText('Stable marker: before manual refresh');
      await expect(page.locator('#content')).not.toContainText('Changed marker: after disk edit');

      await page.click('.refresh-btn');
      await expect(page.locator('#content')).toContainText('Changed marker: after disk edit');
      await expect(page.locator('#content')).not.toContainText('Stable marker: before manual refresh');

      await page.screenshot({ path: 'test-results/fast-08-disabled-auto-refresh-manual-only.png' });
      console.log('✅ Disabled Auto-Refresh E2E PASSED - manual refresh required as expected');
    } finally {
      await fs.unlink(tmpFilePath).catch(() => {
        // Keep cleanup best-effort; tmp folder is intentionally unmanaged.
      });
    }
  });
});
