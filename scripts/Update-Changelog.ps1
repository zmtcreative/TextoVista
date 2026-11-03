#Requires -Version 7.2

<#
    .SYNOPSIS
        Uses the Antigravity CLI (agy) to summarize recent Git changes since the latest SemVer tag
        and updates the '## [Unreleased]' section in CHANGELOG.md.

    .DESCRIPTION
        Identifies the most recent semantic version tag (matching 'vX.Y.Z'), collects all commits
        up to HEAD (or a specified ref), and invokes the Antigravity CLI ('agy') non-interactively.

        The AI generates a high-level narrative summary and categorized bullet lists under standard
        Keep a Changelog subheadings (### Added, ### Changed, ### Deprecated, ### Removed,
        ### Fixed, ### Security), formatted consistently with existing entries in CHANGELOG.md.
        Existing non-placeholder notes under '## [Unreleased]' are provided to the prompt to be
        preserved and incorporated. The resulting content is then cleanly written under the
        '## [Unreleased]' section at the top of CHANGELOG.md.

    .PARAMETER SinceTag
        The Git version tag to compare from (e.g., 'v0.7.0').
        Defaults to the most recent 'vX.Y.Z' tag discovered in the repository.
        Alias: -FromTag.

    .PARAMETER ToRef
        The Git reference or branch to compare up to. Defaults to 'HEAD'.

    .PARAMETER ChangelogPath
        Path to CHANGELOG.md. Defaults to the CHANGELOG.md file in the project root.

    .PARAMETER Model
        Optional Antigravity CLI model to specify via 'agy --model <model>'.

    .PARAMETER DryRun
        Generates and displays the proposed '## [Unreleased]' changelog content without modifying CHANGELOG.md.
        Alias: -WhatIf.

    .PARAMETER NonInteractive
        Applies changes to CHANGELOG.md without an interactive confirmation prompt.
        Aliases: -Force, -Yes.

    .PARAMETER ShowPrompt
        Prints the exact prompt sent to the Antigravity CLI for inspection/debugging.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1
        Discovers the most recent SemVer tag, queries commits up to HEAD, drafts release notes
        via agy, and prompts to update the '## [Unreleased]' section in CHANGELOG.md.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1 -DryRun
        Previews the AI-generated release notes without modifying CHANGELOG.md.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1 -SinceTag "v0.7.0" -NonInteractive
        Summarizes changes since v0.7.0 and updates CHANGELOG.md directly without confirmation.
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [Alias("FromTag")]
    [string]$SinceTag = "",

    [Parameter(Mandatory = $false)]
    [string]$ToRef = "HEAD",

    [Parameter(Mandatory = $false)]
    [string]$ChangelogPath = "",

    [Parameter(Mandatory = $false)]
    [string]$Model = "",

    [Parameter(Mandatory = $false)]
    [Alias("WhatIf")]
    [switch]$DryRun,

    [Parameter(Mandatory = $false)]
    [Alias("Force", "Yes")]
    [switch]$NonInteractive,

    [Parameter(Mandatory = $false)]
    [switch]$ShowPrompt
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# --- Path Resolution & Project Setup ---
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptDir = if (-not [string]::IsNullOrWhiteSpace($ScriptFullName)) {
    Split-Path -Parent $ScriptFullName
} elseif (-not [string]::IsNullOrWhiteSpace($PSScriptRoot)) {
    $PSScriptRoot
} else {
    (Get-Location).Path
}

$ScriptProjectRoot = if ($ScriptDir -match '[\\/]scripts[\\/]?$') {
    Split-Path -Parent $ScriptDir
} elseif (Test-Path (Join-Path (Get-Location) "wails.json")) {
    (Get-Location).Path
} else {
    $ScriptDir
}

if (-not (Test-Path -Path (Join-Path $ScriptProjectRoot "wails.json") -PathType Leaf)) {
    throw "Could not locate project root (wails.json not found in '$ScriptProjectRoot')."
}

# --- Shared Helpers Dot-Sourcing ---
$HelperPath = Join-Path $ScriptDir "helpers/Changelog-Helpers.ps1"
if (-not (Test-Path -Path $HelperPath -PathType Leaf)) {
    throw "Helper script 'Changelog-Helpers.ps1' was not found at '$HelperPath'."
}
. $HelperPath

# ==============================================================================
# Helper Functions
# ==============================================================================

function Get-ProjectRootDirectory {
    return (Resolve-Path $script:ScriptProjectRoot).Path
}

function Assert-Prerequisites {
    param([string]$ProjectRoot)

    # Verify Git
    if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) {
        throw "Git command-line utility ('git') is not installed or not in PATH."
    }

    # Verify Git repo
    $insideWorkTree = git rev-parse --is-inside-work-tree 2>$null
    if ($LASTEXITCODE -ne 0 -or $insideWorkTree.Trim() -ne "true") {
        throw "The project directory '$ProjectRoot' is not a valid Git repository."
    }

    # Verify Antigravity CLI ('agy')
    if (-not (Get-Command "agy" -ErrorAction SilentlyContinue)) {
        throw "Antigravity CLI ('agy') was not found in PATH.`nPlease ensure the Antigravity CLI is installed and available in your environment."
    }
}

# ==============================================================================
# Main Execution Flow
# ==============================================================================

try {
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host -ForegroundColor Cyan "             TextoVista - Changelog Update Tool                   "
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""

    # 1. Resolve paths
    $projectRoot = Get-ProjectRootDirectory
    if ([string]::IsNullOrWhiteSpace($ChangelogPath)) {
        $ChangelogPath = Join-Path $projectRoot "CHANGELOG.md"
    }
    $ChangelogPath = (Resolve-Path $ChangelogPath).Path

    # 2. Check prerequisites
    Assert-Prerequisites -ProjectRoot $projectRoot

    # 3. Determine baseline tag
    if ([string]::IsNullOrWhiteSpace($SinceTag)) {
        $SinceTag = Get-LatestSemanticVersionTag
        if ([string]::IsNullOrWhiteSpace($SinceTag)) {
            throw "Could not automatically determine the most recent 'vX.Y.Z' version tag. Please specify -SinceTag manually."
        }
        Write-Host -ForegroundColor Green "  Discovered most recent version tag: " -NoNewline
        Write-Host -ForegroundColor Yellow "$SinceTag"
    } else {
        Write-Host -ForegroundColor Green "  Using specified baseline tag:       " -NoNewline
        Write-Host -ForegroundColor Yellow "$SinceTag"
    }

    # 3b. Determine project version & stability warning alert
    $projectVersion = Get-CurrentVersionFromWails -ProjectRoot $projectRoot
    $alertMarkdown = Get-VersionAlertMarkdown -VersionString $projectVersion
    $stabilityStatus = if ($alertMarkdown -match '\*\*Alpha\*\*') { "Alpha" } elseif ($alertMarkdown -match '\*\*Beta\*\*') { "Beta" } else { "Production (no alert)" }

    Write-Host -ForegroundColor Green "  Current project version:            " -NoNewline
    Write-Host -ForegroundColor Yellow "$projectVersion ($stabilityStatus)"
    Write-Host -ForegroundColor Green "  Target ref:                         " -NoNewline
    Write-Host -ForegroundColor Yellow "$ToRef"
    Write-Host -ForegroundColor Green "  Changelog path:                     " -NoNewline
    Write-Host -ForegroundColor Yellow "$ChangelogPath"
    Write-Host ""

    # 4. Collect git changes
    Write-Host -ForegroundColor Cyan "  Querying Git history since '$SinceTag'..."
    $gitData = Get-GitCommitHistory -FromRef $SinceTag -ToRef $ToRef

    Write-Host -ForegroundColor Green "  Commits found:                      " -NoNewline
    Write-Host -ForegroundColor Yellow "$($gitData.CommitCount)"

    if ($gitData.CommitCount -eq 0) {
        Write-Host ""
        Write-Host -ForegroundColor Yellow "Notice: No new commits found between '$SinceTag' and '$ToRef'."
        Write-Host -ForegroundColor Yellow "The '## [Unreleased]' section does not need updating at this time."
        Write-Host ""
        exit 0
    }

    # 5. Extract reference style and unreleased content
    $referenceStyle = Get-ChangelogReferenceStyle -ChangelogPath $ChangelogPath
    $currentUnreleased = Get-CurrentUnreleasedContent -ChangelogPath $ChangelogPath

    # 6. Formulate prompt
    $prompt = Build-AgyChangelogPrompt `
        -SinceTag $SinceTag `
        -ToRef $ToRef `
        -GitData $gitData `
        -ReferenceStyle $referenceStyle `
        -CurrentUnreleased $currentUnreleased

    if ($ShowPrompt) {
        Write-Host ""
        Write-Host -ForegroundColor DarkCyan "---------------------- [AGY PROMPT] ----------------------"
        Write-Host $prompt
        Write-Host -ForegroundColor DarkCyan "----------------------------------------------------------"
        Write-Host ""
    }

    # 7. Invoke Antigravity CLI
    $summaryMarkdown = Invoke-AgyChangelogSummary -Prompt $prompt -Model $Model

    if ([string]::IsNullOrWhiteSpace($summaryMarkdown)) {
        throw "Antigravity CLI returned empty changelog content."
    }

    # 8. Format section with stability warning alert positioned after summary and before category subheadings
    $unreleasedBody = Format-UnreleasedContent -SummaryMarkdown $summaryMarkdown -AlertMarkdown $alertMarkdown

    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host -ForegroundColor Cyan "               Proposed '## [Unreleased]' Content                 "
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""
    Write-Host -ForegroundColor White "## [Unreleased]"
    Write-Host ""
    Write-Host $unreleasedBody
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""

    # 9. Dry Run check
    if ($DryRun) {
        Write-Host -ForegroundColor Yellow "DryRun (-WhatIf) active: CHANGELOG.md was NOT modified."
        Write-Host ""
        exit 0
    }

    # 10. Confirmation prompt
    if (-not $NonInteractive) {
        $confirm = Read-Host "Do you want to write these updates to CHANGELOG.md? (Y/n)"
        if ($confirm -match '^(n|no)$') {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "Operation cancelled by user. CHANGELOG.md was not modified."
            Write-Host ""
            exit 0
        }
    }

    # 11. Update CHANGELOG.md
    Write-Host -ForegroundColor Cyan "  Applying updates to $ChangelogPath..."
    $updatedDocument = Update-ChangelogDocument -ChangelogPath $ChangelogPath -NewUnreleasedBody $unreleasedBody -SinceTag $SinceTag
    [System.IO.File]::WriteAllText($ChangelogPath, $updatedDocument, [System.Text.UTF8Encoding]::new($false))

    Write-Host -ForegroundColor Green "  CHANGELOG.md successfully updated!"
    Write-Host ""
}
catch {
    Write-Host ""
    Write-Host -ForegroundColor Red "ERROR: $($_.Exception.Message)"
    Write-Host ""
    exit 1
}
