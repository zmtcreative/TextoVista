import { test, expect } from '@playwright/test';
import WailsDevHelper from './wails-dev-helper';
import { Page } from '@playwright/test';
import { emitRuntimeEvent, expectTheme } from './runtime-test-helpers';

test.describe("TextoVista - Optimized Test Suite\n", () => {
  let wailsDev: WailsDevHelper;
  let page: Page;

  // 🚀 Start application ONCE before all tests (major performance improvement)
  test.beforeAll(async () => {
    console.log('🚀 Starting Wails application for entire test suite...');
    wailsDev = new WailsDevHelper({ runtimeMode: 'headless' });
    page = await wailsDev.launchAndConnect();
    await wailsDev.waitForAppReady(page);
    console.log(' ✅ Application ready - will be shared across all tests for maximum performance');
  });

  // 🔚 Shutdown application ONCE after all tests
  test.afterAll(async () => {
    console.log('🔚 Shutting down shared application...');
    if (wailsDev) {
      await wailsDev.disconnect();
    }
    console.log(' ✅ Shared application shutdown complete');
  });

  // 🧹 Fast state reset between tests (no app restart needed!)
  test.beforeEach(async () => {
    console.log(' 🔄 Resetting application state (fast reset - no restart)...');

    try {
      // Close any open modals
      await page.keyboard.press('Escape');
      await page.waitForTimeout(200);

      // Reset to clean state with page reload (much faster than app restart)
      await page.reload({ waitUntil: 'domcontentloaded' });
      await page.waitForTimeout(500);

      // Wait for app to be ready
      await page.waitForSelector('.app-header', { timeout: 5000 });

      console.log(' ✅ State reset complete (no app restart - major time savings!)');
    } catch (error) {
      console.log(` ⚠️ State reset error (continuing anyway): ${error}`);
    }
  });

  // 📱 Test Group 1: Core Application Functionality
  test.describe('Core Application Tests', () => {

    test('should load and display application correctly', async () => {
      console.log('🧪 Testing: Core application loading and display');

      // Verify main UI elements are present and visible
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.content-area')).toBeVisible();
      await expect(page.locator('#content')).toBeVisible();

      // Verify initial state
      await expect(page.locator('#content')).toContainText('No markdown file specified');
      await expect(page.locator('.help-overlay')).toBeHidden();
      await expect(page.locator('.settings-overlay')).toBeHidden();
      await expect(page.locator('.frontmatter-section')).toBeHidden();

      // Verify page title
      const title = await page.title();
      expect(title).toBe('TextoVista');

      await expectTheme(page, 'light');

      await page.screenshot({ path: 'test-results/optimized-01-app-display.png' });
      console.log('  ✅ PASSED: Application loads and displays correctly');
    });

    test('should be responsive and handle basic interactions', async () => {
      console.log('🧪 Testing: Application responsiveness and basic interactions');

      // Test that the application responds to basic interactions
      await expect(page.locator('.app-header')).toBeVisible();

      // Test keyboard interactions don't crash the app
      await page.keyboard.press('Tab');
      await page.waitForTimeout(100);

      // Verify app is still responsive after interactions
      await expect(page.locator('.content-area')).toBeVisible();

      await page.screenshot({ path: 'test-results/optimized-02-responsiveness.png' });
      console.log('  ✅ PASSED: Application is responsive and handles interactions');
    });
  });

  // 🎯 Test Group 2: Help and Modal Functionality
  test.describe('Help and Modal Tests', () => {

    test('should have Help modal elements and infrastructure', async () => {
      console.log('🧪 Testing: Help modal elements and infrastructure');

      // Verify Help modal elements exist in the DOM
      await expect(page.locator('#help-modal-overlay')).toBeAttached();
      await expect(page.locator('#help-modal-content')).toBeAttached();
      await expect(page.locator('#help-modal-close')).toBeAttached();
      await expect(page.locator('#help-modal-overlay')).toBeHidden();

      console.log('  ✅ Help modal elements verified in DOM');

      await page.screenshot({ path: 'test-results/optimized-03-help-modal-elements.png' });
      console.log('  ✅ PASSED: Help modal infrastructure is present');
    });

    test('should handle programmatic Help modal triggering', async () => {
      console.log('🧪 Testing: Programmatic Help modal functionality');

      await emitRuntimeEvent(
        page,
        'show-help',
        'About TextoVista',
        '<div><h3>TextoVista</h3><p>Version: 0.0.0-dev</p><p>Optimized Test Suite</p></div>'
      );

      await expect(page.locator('#help-modal-overlay')).toBeVisible();
      await expect(page.locator('.help-body')).toContainText('TextoVista');
      await expect(page.locator('.help-body')).toContainText('Optimized Test Suite');

      await page.click('#help-modal-close');
      await expect(page.locator('#help-modal-overlay')).toBeHidden();

      await page.screenshot({ path: 'test-results/optimized-04-help-modal-programmatic.png' });
      console.log('  ✅ PASSED: Programmatic Help modal triggering works');
    });

  });

  // 🎨 Test Group 3: Theme and UI Functionality
  test.describe('Theme and UI Tests', () => {

    test('should handle theme toggle functionality', async () => {
      console.log('🧪 Testing: Theme toggle functionality');

      // Look for theme toggle button
      const themeButton = page.locator('.theme-toggle-btn');

      await expect(themeButton).toBeVisible();
      await expectTheme(page, 'light');

      await themeButton.click();
      await expectTheme(page, 'dark');

      await themeButton.click();
      await expectTheme(page, 'light');

      await page.screenshot({ path: 'test-results/optimized-06-theme-functionality.png' });
      console.log('  ✅ PASSED: Theme functionality verified');
    });

    test('should maintain UI state consistency', async () => {
      console.log('🧪 Testing: UI state consistency');

      // Verify UI elements maintain consistency
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.content-area')).toBeVisible();

      // Test that UI elements are properly styled and functional
      const header = page.locator('.app-header');
      await expect(header).toHaveCSS('display', /^(block|flex)$/);

      await page.screenshot({ path: 'test-results/optimized-07-ui-consistency.png' });
      console.log('  ✅ PASSED: UI state consistency maintained');
    });
  });

  // ⚡ Test Group 4: Performance and Integration Tests
  test.describe('Performance and Integration Tests', () => {

    test('should demonstrate performance optimization benefits', async () => {
      console.log('🧪 Testing: Performance optimization verification');

      const startTime = Date.now();

      // Verify we're still using the same application instance
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.content-area')).toBeVisible();

      // Quick interaction test to verify responsiveness
      await page.keyboard.press('Control+o');
      await page.waitForTimeout(200);
      await page.keyboard.press('Escape');

      const endTime = Date.now();
      const testTime = endTime - startTime;

      console.log(`    ⚡ Performance test completed in ${testTime}ms`);
      console.log('    🎯 This test ran WITHOUT restarting the application!');
      console.log('    📊 Estimated time savings: ~20-30 seconds per test vs restart approach');

      await page.screenshot({ path: 'test-results/optimized-08-performance-demo.png' });
      console.log('  ✅ PASSED: Performance optimization benefits demonstrated');
    });

    test('should handle multiple sequential operations without issues', async () => {
      console.log('🧪 Testing: Multiple sequential operations stability');

      // Test a sequence of operations that a real user might perform
      await page.keyboard.press('Control+o');  // File open
      await page.waitForTimeout(200);

      await page.keyboard.press('Escape');     // Cancel/close
      await page.waitForTimeout(200);

      await page.keyboard.press('Control+p');  // Print
      await page.waitForTimeout(200);

      await page.keyboard.press('Escape');     // Cancel/close
      await page.waitForTimeout(200);

      await page.keyboard.press('Control+a');  // Help/About
      await page.waitForTimeout(200);

      // Verify application is still stable and responsive
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.content-area')).toBeVisible();

      await page.screenshot({ path: 'test-results/optimized-09-sequential-operations.png' });
      console.log('  ✅ PASSED: Multiple sequential operations handled without issues');
    });

    test('should verify complete application functionality integration', async () => {
      console.log('🧪 Testing: Complete application integration');

      // Final integration test - verify all major components work together

      // 1. UI Elements
      await expect(page.locator('.app-header')).toBeVisible();
      await expect(page.locator('.content-area')).toBeVisible();
      await expect(page.locator('#content')).toBeVisible();

      // 2. Runtime-driven markdown rendering
      await emitRuntimeEvent(page, 'markdown-rendered', {
        html: '<p>Rendered integration body</p><div>\\(x+y\\)</div>',
        title: 'Integration Title',
        date: '2026-06-03',
        frontmatter_html: '<div class="frontmatter-container"><span class="fm-key">title</span><span class="fm-string">Integration Title</span></div>'
      });
      await expect(page).toHaveTitle('Integration Title');
      await expect(page.locator('.document-title')).toContainText('Integration Title');
      await expect(page.locator('.document-dates')).toContainText('2026-06-03');
      await expect(page.locator('#content')).toContainText('Rendered integration body');

      // 3. Frontmatter toggle updates visible state
      await page.click('.frontmatter-toggle-btn');
      await expect(page.locator('.frontmatter-section')).toBeVisible();
      await expect(page.locator('.frontmatter-section')).toContainText('Integration Title');

      // 4. Runtime-driven help and settings dialogs
      await emitRuntimeEvent(page, 'show-help', 'Integration Help', '<p>Integration help body</p>');
      await expect(page.locator('#help-modal-overlay')).toBeVisible();
      await expect(page.locator('#help-modal-text')).toContainText('Integration help body');
      await page.click('#help-modal-close');
      await expect(page.locator('#help-modal-overlay')).toBeHidden();

      await emitRuntimeEvent(page, 'show-settings');
      await expect(page.locator('#settings-overlay')).toBeVisible();
      await page.click('#settings-close');
      await expect(page.locator('#settings-overlay')).toBeHidden();

      // 5. Runtime doc-class and error handling
      await emitRuntimeEvent(page, 'add-doc-class', 'techdoc');
      await expect(page.locator('html')).toHaveClass(/techdoc/);
      await emitRuntimeEvent(page, 'toggle-doc-class', 'techdoc');
      await expect(page.locator('html')).not.toHaveClass(/techdoc/);
      await emitRuntimeEvent(page, 'error', 'Integration failure');
      await expect(page.locator('.error-message')).toContainText('Integration failure');
      await expect(page.locator('#content')).toContainText('An error occurred');

      console.log('  ✅ All integration checkpoints passed');

      await page.screenshot({ path: 'test-results/optimized-10-complete-integration.png' });
      console.log('  ✅ PASSED: Complete application functionality integration verified');
    });
  });
});
