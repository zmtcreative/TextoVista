# Optimized Playwright Testing Script for Wails Application
# This script runs the Playwright frontend suites using the Wails dev environment.
# Tests automatically manage their own Wails dev server - no build required.

param(
    [ValidateSet("main", "fast", "interactive", "all")]
    [string]$TestSuite = "main",

    [ValidateSet("auto", "headless", "interactive")]
    [string]$RuntimeMode = "auto",

    [switch]$ShowReport = $false
)

Write-Host "🚀 TextoVista - Optimized Playwright Testing Script" -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan

# Function to check if a command exists
function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    } catch {
        return $false
    }
}

# Check prerequisites
Write-Host "🔍 Checking prerequisites..." -ForegroundColor Yellow

if (-not (Test-Command "wails")) {
    Write-Error "❌ Wails CLI not found. Please install Wails v2 first."
    exit 1
}

if (-not (Test-Command "npm")) {
    Write-Error "❌ npm not found. Please install Node.js first."
    exit 1
}

# Set working directory to project root
$projectRoot = Split-Path $PSScriptRoot -Parent
Set-Location $projectRoot

Write-Host "📂 Project root: $projectRoot" -ForegroundColor Green
Write-Host "ℹ️ Tests will manage Wails dev server automatically (no build required)" -ForegroundColor Blue

# Install Playwright dependencies
Write-Host "📦 Installing Playwright dependencies..." -ForegroundColor Yellow
Set-Location "frontend"

$packageLockPath = Join-Path $PWD "package-lock.json"
$installCommand = if ($env:CI -or (Test-Path $packageLockPath)) { "ci" } else { "install" }

try {
    Write-Host "📥 Using npm $installCommand for frontend dependencies" -ForegroundColor Gray
    & npm $installCommand
    if ($LASTEXITCODE -ne 0) {
        throw "npm $installCommand failed"
    }
    Write-Host "✅ Dependencies installed!" -ForegroundColor Green
} catch {
    Write-Error "❌ Failed to install dependencies: $_"
    exit 1
}

# Install Playwright browsers if needed
Write-Host "🌐 Ensuring Playwright browsers are installed..." -ForegroundColor Yellow
try {
    npx playwright install chromium
    if ($LASTEXITCODE -ne 0) {
        throw "Playwright browser installation failed"
    }
    Write-Host "✅ Playwright browsers ready!" -ForegroundColor Green
} catch {
    Write-Error "❌ Failed to install Playwright browsers: $_"
    exit 1
}

# Prepare test results directory
$testResultsDir = Join-Path $PWD "test-results"
if (Test-Path $testResultsDir) {
    Write-Host "🧹 Cleaning previous test results..." -ForegroundColor Gray
    Remove-Item $testResultsDir -Recurse -Force
}
New-Item -ItemType Directory -Path $testResultsDir -Force | Out-Null
Write-Host "📁 Test results directory prepared: $testResultsDir" -ForegroundColor Green

# Determine which test suite to run
Write-Host "🧪 Running Optimized Playwright Test Suite (with Wails dev server)..." -ForegroundColor Yellow

$selectedScript = $null
$runtimeModeValue = $null

# Add test suite selection
switch ($TestSuite.ToLower()) {
    "main" {
        $selectedScript = "test:e2e:main"
        $runtimeModeValue = "headless"
        Write-Host "🎯 Running Main Test Suite (comprehensive, headless-safe)" -ForegroundColor Blue
    }
    "fast" {
        $selectedScript = "test:e2e:fast"
        $runtimeModeValue = "headless"
        Write-Host "⚡ Running Fast Sequential Test Suite (performance demo, headless-safe)" -ForegroundColor Blue
    }
    "interactive" {
        $selectedScript = "test:e2e:interactive"
        $runtimeModeValue = "interactive"
        Write-Host "👁️ Running Interactive Native Shortcut Suite" -ForegroundColor Blue
    }
    "all" {
        $selectedScript = "test:e2e:all"
        Write-Host "🔄 Running All Available Test Suites" -ForegroundColor Blue
    }
}

$requestedRuntimeMode = if ($RuntimeMode -eq "auto") { $runtimeModeValue } else { $RuntimeMode }

if ($requestedRuntimeMode) {
    Write-Host "🖥️ Runtime mode: $requestedRuntimeMode" -ForegroundColor Blue
} else {
    Write-Host "🖥️ Runtime mode: suite defaults" -ForegroundColor Blue
}

Write-Host "ℹ️ Tests will automatically:" -ForegroundColor Cyan
Write-Host "   • Start Wails dev server (wails dev)" -ForegroundColor Gray
Write-Host "   • Connect using the selected Playwright runtime mode" -ForegroundColor Gray
Write-Host "   • Run optimized test suite with shared app instances" -ForegroundColor Gray
Write-Host "   • Clean up Wails dev server when complete" -ForegroundColor Gray

try {
    $previousRuntimeMode = $env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE
    $previousLegacyRuntimeMode = $env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE
    if ($requestedRuntimeMode) {
        $env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE = $requestedRuntimeMode
        $env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE = $requestedRuntimeMode
    } else {
        Remove-Item Env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
        Remove-Item Env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
    }

    Write-Host "🎬 Executing: npm run $selectedScript" -ForegroundColor Gray

    & npm run $selectedScript
    $testExitCode = $LASTEXITCODE

    if ($null -ne $previousRuntimeMode) {
        $env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE = $previousRuntimeMode
    } else {
        Remove-Item Env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
    }
    if ($null -ne $previousLegacyRuntimeMode) {
        $env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE = $previousLegacyRuntimeMode
    } else {
        Remove-Item Env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
    }

    if ($testExitCode -eq 0) {
        Write-Host "✅ All tests passed!" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Some tests failed (exit code: $testExitCode)" -ForegroundColor Yellow
    }
} catch {
    if ($null -ne $previousRuntimeMode) {
        $env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE = $previousRuntimeMode
    } else {
        Remove-Item Env:TEXTOVISTA_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
    }
    if ($null -ne $previousLegacyRuntimeMode) {
        $env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE = $previousLegacyRuntimeMode
    } else {
        Remove-Item Env:MARKDOWN_READER_PLAYWRIGHT_RUNTIME_MODE -ErrorAction SilentlyContinue
    }
    Write-Error "❌ Failed to run tests: $_"
    exit 1
}

# Show test report if requested
if ($ShowReport) {
    Write-Host "📊 Opening test report..." -ForegroundColor Yellow
    try {
        npx playwright show-report --host 127.0.0.1 --port 9323
    } catch {
        Write-Host "ℹ️ Could not open test report automatically" -ForegroundColor Blue
        Write-Host "💡 You can view the report manually with: npx playwright show-report" -ForegroundColor Blue
    }
}

# Return to project root
Set-Location $projectRoot

Write-Host "🎉 Testing completed!" -ForegroundColor Cyan
Write-Host "📸 Screenshots and videos are available in: frontend/test-results/" -ForegroundColor Cyan
Write-Host "📋 HTML report is available in: frontend/playwright-report/" -ForegroundColor Cyan

exit $testExitCode
