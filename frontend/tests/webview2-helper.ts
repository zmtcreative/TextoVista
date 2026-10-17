import { chromium, Page, Browser } from '@playwright/test';

const CDP_PORT = 9222;

export class WebView2Helper {
  private browser: Browser | null = null;
  private page: Page | null = null;
  private appProcess: any = null; // ChildProcess type

  /**
   * Launch the Wails application with debugging enabled
   */
  async launchApp(): Promise<void> {
    console.log('🚀 Launching Wails application...');

    // Dynamic import to avoid TypeScript issues
    const { spawn } = await import('child_process');
    const path = await import('path');

    // Try development mode first (wails dev), which has debugging enabled by default
    const useDevMode = true; // Set to false to use production build

    if (useDevMode) {
      console.log('🔧 Using development mode (wails dev)');
      return this.launchDevApp();
    }

    // Production build approach
    const appPath = path.resolve('../build/bin/texto-vista.exe');
    console.log(`📂 Application path: ${appPath}`);

    // Set environment variables for WebView2 debugging
    const env: any = {};
    try {
      // Access process.env safely
      Object.assign(env, (global as any).process?.env || {});
    } catch {
      // Fallback if process is not available
    }

    env['WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS'] = `--remote-debugging-port=${CDP_PORT} --enable-logging --log-level=0`;
    env['WEBVIEW2_USER_DATA_FOLDER'] = path.resolve('../tmp/webview2-test-data');

    // Launch the application with debugging environment
    this.appProcess = spawn(appPath, [], {
      stdio: 'pipe',
      detached: false,
      env: env,
    });

    if (!this.appProcess) {
      throw new Error('Failed to start the application process');
    }

    console.log(`🔧 Application launched with PID: ${this.appProcess.pid}`);
    console.log(`🔍 Debug port: ${CDP_PORT}`);
    console.log(`📁 User data folder: ${env.WEBVIEW2_USER_DATA_FOLDER}`);

    // Set up process event handlers
    this.appProcess.stdout?.on('data', (data: any) => {
      console.log(`📝 APP STDOUT: ${data.toString().trim()}`);
    });

    this.appProcess.stderr?.on('data', (data: any) => {
      console.log(`⚠️ APP STDERR: ${data.toString().trim()}`);
    });

    this.appProcess.on('error', (error: Error) => {
      console.error('❌ Application process error:', error);
    });

    this.appProcess.on('exit', (code: number | null, signal: string | null) => {
      if (code !== null) {
        console.log(`🔚 Application exited with code: ${code}`);
      } else {
        console.log(`🔚 Application terminated by signal: ${signal}`);
      }
    });

    // Wait for the application to start and WebView2 to be ready
    await this.waitForWebView2Ready();
  }

  /**
   * Launch using wails dev (development mode) which has debugging enabled by default
   */
  private async launchDevApp(): Promise<void> {
    const { spawn } = await import('child_process');
    const path = await import('path');

    console.log('🔧 Starting Wails development server...');

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

    // Set up process event handlers
    this.appProcess.stdout?.on('data', (data: any) => {
      const output = data.toString().trim();
      console.log(`📝 WAILS DEV: ${output}`);

      // Look for the debugging port in the output
      if (output.includes('remote-debugging-port')) {
        console.log('🎯 Found debugging port in wails dev output');
      }
    });

    this.appProcess.stderr?.on('data', (data: any) => {
      console.log(`⚠️ WAILS DEV STDERR: ${data.toString().trim()}`);
    });

    this.appProcess.on('error', (error: Error) => {
      console.error('❌ Wails dev process error:', error);
    });

    this.appProcess.on('exit', (code: number | null, signal: string | null) => {
      if (code !== null) {
        console.log(`🔚 Wails dev exited with code: ${code}`);
      } else {
        console.log(`🔚 Wails dev terminated by signal: ${signal}`);
      }
    });

    // Wait longer for dev mode to start
    console.log('⏳ Waiting for wails dev to start (this may take a while)...');
    await new Promise(resolve => setTimeout(resolve, 10000)); // 10 second delay for dev startup

    // Wait for the application to start and WebView2 to be ready
    await this.waitForWebView2Ready();
  }

  /**
   * Wait for WebView2 to be ready for connections
   */
  private async waitForWebView2Ready(): Promise<void> {
    const maxAttempts = 30; // 30 seconds timeout
    let attempts = 0;

    while (attempts < maxAttempts) {
      try {
        console.log(`🔍 Attempting to connect to CDP on port ${CDP_PORT}... (attempt ${attempts + 1}/${maxAttempts})`);
        const browser = await chromium.connectOverCDP(`http://localhost:${CDP_PORT}`);
        const contexts = browser.contexts();
        console.log(`✅ Connected to WebView2! Found ${contexts.length} context(s)`);
        await browser.close();
        return; // Success!
      } catch (error) {
        console.log(`⏳ WebView2 not ready yet, retrying... (${error})`);
        attempts++;
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }

    throw new Error(`Failed to connect to WebView2 after ${maxAttempts} attempts`);
  }

  /**
   * Connect to the WebView2 application via Chrome DevTools Protocol
   */
  async connect(): Promise<Page> {
    console.log(`🔗 Connecting to WebView2 on port ${CDP_PORT}...`);

    this.browser = await chromium.connectOverCDP(`http://localhost:${CDP_PORT}`);

    // Get the contexts and find the main application context
    const contexts = this.browser.contexts();
    console.log(`📱 Found ${contexts.length} context(s)`);

    if (contexts.length === 0) {
      throw new Error('No WebView2 contexts found');
    }

    // Usually the application context is the first one
    const context = contexts[0];
    const pages = context.pages();

    if (pages.length === 0) {
      throw new Error('No pages found in WebView2 context');
    }

    this.page = pages[0];
    console.log(`📄 Connected to page: ${await this.page.title()}`);

    return this.page;
  }

  /**
   * Launch app and connect in one step
   */
  async launchAndConnect(): Promise<Page> {
    await this.launchApp();
    return await this.connect();
  }

  /**
   * Wait for the application to be fully loaded
   */
  async waitForAppReady(page: Page): Promise<void> {
    console.log('⏳ Waiting for application to be ready...');

    // Wait for Vue app to be mounted
    await page.waitForSelector('.app-header', { timeout: 15000 });

    // Wait for the main content area
    await page.waitForSelector('.content-area', { timeout: 5000 });

    console.log('✅ Application is ready!');
  }

  /**
   * Simulate native menu click by sending keyboard shortcuts
   * Since Playwright can't directly interact with native OS menus,
   * we'll use keyboard shortcuts to trigger menu actions
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
   * Handle native file dialog (this is tricky with WebView2)
   * We'll need to use a different approach since we can't directly control native dialogs
   */
  async handleFileDialog(page: Page, filePath: string): Promise<void> {
    console.log(`📁 Attempting to handle file dialog for: ${filePath}`);

    // Note: This is a limitation - we can't directly control native file dialogs
    // In a real test environment, you might need to:
    // 1. Mock the file dialog
    // 2. Use automation tools like AutoHotkey
    // 3. Test the file loading functionality directly via the backend API

    console.log('⚠️ Native file dialog handling is limited in this test setup');
    console.log('💡 Consider testing file loading via backend API calls instead');
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
   * Cleanup resources and terminate the application
   */
  async disconnect(): Promise<void> {
    console.log('🔌 Disconnecting from WebView2...');

    // First, properly shutdown the Wails application
    await this.shutdownApp();

    if (this.browser) {
      await this.browser.close();
      this.browser = null;
      this.page = null;
    }

    if (this.appProcess && !this.appProcess.killed) {
      console.log(`🛑 Terminating application process (PID: ${this.appProcess.pid})...`);

      // Try graceful shutdown first
      this.appProcess.kill('SIGTERM');

      // Wait a bit for graceful shutdown
      await new Promise(resolve => setTimeout(resolve, 2000));

      // Force kill if still running
      if (!this.appProcess.killed) {
        console.log('💥 Force killing application process...');
        this.appProcess.kill('SIGKILL');
      }

      console.log('✅ Application process terminated');
      this.appProcess = null;
    }

    console.log('✅ Disconnected from WebView2');
  }
}

export default WebView2Helper;
