import { chromium, Page, Browser } from '@playwright/test';

export type RuntimeMode = 'headless' | 'interactive';

export interface WailsDevHelperOptions {
  runtimeMode?: RuntimeMode;
}

export class WailsDevHelper {
  private browser: Browser | null = null;
  private page: Page | null = null;
  private appProcess: any = null; // ChildProcess type
  private devUrl: string = '';
  private runtimeMode: RuntimeMode;

  constructor(options: WailsDevHelperOptions = {}) {
    this.runtimeMode = this.resolveRuntimeMode(options.runtimeMode);
  }

  private resolveRuntimeMode(runtimeMode?: RuntimeMode): RuntimeMode {
    const envRuntimeMode = process.env.TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE || process.env.MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE;
    if (envRuntimeMode === 'headless' || envRuntimeMode === 'interactive') {
      return envRuntimeMode;
    }

    if (runtimeMode) {
      return runtimeMode;
    }

    return process.env.CI ? 'headless' : 'interactive';
  }

  /**
   * Launch the Wails development server
   */
  async launchWailsDev(): Promise<void> {
    console.log('🚀 Launching Wails development server...');

    // Dynamic import to avoid TypeScript issues
    const { spawn } = await import('child_process');
    const path = await import('path');

    // Change to project root directory for wails dev
    const projectRoot = path.resolve('..');

    // Launch wails dev from the project root
    this.appProcess = spawn('wails', ['dev'], {
      stdio: 'pipe',
      detached: false,
      cwd: projectRoot,
    });

    if (!this.appProcess) {
      throw new Error('Failed to start wails dev process');
    }

    console.log(`🔧 Wails dev launched with PID: ${this.appProcess.pid}`);

    // Set up process event handlers and capture the dev URL
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Wails dev failed to start within 60 seconds'));
      }, 60000); // 60 second timeout

      this.appProcess.stdout?.on('data', (data: any) => {
        const output = data.toString().trim();
        console.log(`📝 WAILS DEV: ${output}`);

        // Look for the browser development URL
        const urlMatch = output.match(/To develop in the browser.*navigate to: (http:\/\/localhost:\d+)/);
        if (urlMatch) {
          this.devUrl = urlMatch[1];
          console.log(`🎯 Found Wails dev URL: ${this.devUrl}`);
          clearTimeout(timeout);
          resolve();
        }
      });

      this.appProcess.stderr?.on('data', (data: any) => {
        const output = data.toString().trim();
        console.log(`⚠️ WAILS DEV STDERR: ${output}`);

        // Also check stderr for the URL
        const urlMatch = output.match(/To develop in the browser.*navigate to: (http:\/\/localhost:\d+)/);
        if (urlMatch) {
          this.devUrl = urlMatch[1];
          console.log(`🎯 Found Wails dev URL in stderr: ${this.devUrl}`);
          clearTimeout(timeout);
          resolve();
        }
      });

      this.appProcess.on('error', (error: Error) => {
        clearTimeout(timeout);
        console.error('❌ Wails dev process error:', error);
        reject(error);
      });

      this.appProcess.on('exit', (code: number | null, signal: string | null) => {
        clearTimeout(timeout);
        if (code !== null && code !== 0) {
          reject(new Error(`Wails dev exited with code: ${code}`));
        } else if (signal) {
          console.log(`🔚 Wails dev terminated by signal: ${signal}`);
        }
      });
    });
  }

  /**
   * Connect to the Wails development server in browser
   */
  async connectToBrowser(): Promise<Page> {
    console.log(`🔗 Connecting to Wails dev server at: ${this.devUrl}`);

    if (!this.devUrl) {
      throw new Error('No Wails dev URL found. Make sure to call launchWailsDev() first.');
    }

    const headless = this.runtimeMode === 'headless';
    console.log(`🖥️ Launching Playwright browser in ${this.runtimeMode} mode...`);

    this.browser = await chromium.launch({
      headless,
      args: ['--disable-web-security', '--disable-features=VizDisplayCompositor'],
    });

    const context = await this.browser.newContext();
    this.page = await context.newPage();

    // Navigate to the Wails development URL
    await this.page.goto(this.devUrl);

    console.log(`📄 Connected to Wails dev page: ${await this.page.title()}`);

    return this.page;
  }

  /**
   * Launch Wails dev and connect in one step
   */
  async launchAndConnect(): Promise<Page> {
    await this.launchWailsDev();
    return await this.connectToBrowser();
  }

  /**
   * Wait for the application to be fully loaded
   */
  async waitForAppReady(page: Page): Promise<void> {
    console.log('⏳ Waiting for Wails application to be ready...');

    // Wait for Vue app to be mounted
    await page.waitForSelector('.app-header', { timeout: 15000 });

    // Wait for the main content area
    await page.waitForSelector('.content-area', { timeout: 5000 });

    console.log('✅ Wails application is ready!');
  }

  /**
   * Simulate native menu click by sending keyboard shortcuts
   */
  async clickHelpMenu(page: Page): Promise<void> {
    console.log('📋 Triggering Help -> About via keyboard shortcut...');
    // Help -> About is Ctrl+A according to the Go code
    await page.keyboard.press('Control+a');
  }

  async clickFileOpen(page: Page): Promise<void> {
    console.log('📂 Triggering File -> Open via keyboard shortcut...');
    // File -> Open is Ctrl+O according to the Go code
    await page.keyboard.press('Control+o');
  }

  /**
   * Wait for the Help modal to appear
   */
  async waitForHelpModal(page: Page): Promise<void> {
    console.log('🔍 Waiting for Help modal to appear...');
    await page.waitForSelector('#help-modal-overlay', { state: 'visible', timeout: 8000 });
    console.log('✅ Help modal is visible!');
  }

  /**
   * Close the Help modal by clicking outside it
   */
  async closeHelpModal(page: Page): Promise<void> {
    console.log('❌ Closing Help modal by clicking outside...');
    // Click on the overlay to close the modal
    await page.click('#help-modal-overlay');

    // Wait for modal to disappear
    await page.waitForSelector('#help-modal-overlay', { state: 'hidden', timeout: 5000 });
    console.log('✅ Help modal closed!');
  }

  /**
   * Properly shutdown the Wails application by calling the backend shutdown function
   */
  async shutdownApp(): Promise<void> {
    if (this.page) {
      console.log('🛑 Calling Wails application shutdown...');

      try {
        // Call the Wails runtime Quit() function to properly shutdown the application
        // This will trigger the shutdown() function in App.go
        await this.page.evaluate(() => {
          const wailsRuntime = (window as any).runtime;
          if (wailsRuntime && wailsRuntime.Quit) {
            console.log('📞 Calling window.runtime.Quit()...');
            wailsRuntime.Quit();
          } else {
            console.log('⚠️ window.runtime.Quit() not available');
          }
        });

        console.log('✅ Wails application shutdown called');

        // Give a moment for the shutdown to process
        await new Promise(resolve => setTimeout(resolve, 1000));
      } catch (error) {
        console.log(`⚠️ Could not call Wails shutdown: ${error}`);
      }
    }
  }

  /**
   * Cleanup resources and terminate the development server
   */
  async disconnect(): Promise<void> {
    console.log('🔌 Disconnecting from Wails dev server...');

    // First, properly shutdown the Wails application
    await this.shutdownApp();

    if (this.browser) {
      await this.browser.close();
      this.browser = null;
      this.page = null;
    }

    if (this.appProcess && !this.appProcess.killed) {
      console.log(`🛑 Terminating Wails dev process (PID: ${this.appProcess.pid})...`);

      // Send Ctrl+C to gracefully stop wails dev
      this.appProcess.kill('SIGINT');

      // Wait a bit for graceful shutdown
      await new Promise(resolve => setTimeout(resolve, 3000));

      // Force kill if still running
      if (!this.appProcess.killed) {
        console.log('💥 Force killing Wails dev process...');
        this.appProcess.kill('SIGKILL');
      }

      console.log('✅ Wails dev process terminated');
      this.appProcess = null;
    }

    console.log('✅ Disconnected from Wails dev server');
  }
}

export default WailsDevHelper;
