#Requires -Version 7.0

<#
    .SYNOPSIS
        Uses the Antigravity CLI (agy) to summarize recent Git changes since the latest version tag
        and updates the '## [Unreleased]' section in CHANGELOG.md.

    .DESCRIPTION
        Identifies the most recent semantic version tag (matching 'vX.Y.Z'), collects all commits
        and diff statistics up to HEAD (or a specified ref), and invokes the Antigravity CLI ('agy')
        non-interactively.

        The AI generates a high-level narrative summary and categorized bullet lists under standard
        Keep a Changelog subheadings (### Added, ### Changed, ### Deprecated, ### Removed,
        ### Fixed, ### Security), formatted consistently with existing entries in CHANGELOG.md.
        The resulting content is then cleanly inserted or updated under the '## [Unreleased]'
        section at the top of CHANGELOG.md.

    .PARAMETER SinceTag
        The Git version tag to compare from (e.g., 'v0.5.3').
        Defaults to the most recent 'vX.Y.Z' tag discovered in the repository.
        Alias: -FromTag.

    .PARAMETER ToRef
        The Git reference or branch to compare up to. Defaults to 'HEAD'.

    .PARAMETER TargetVersion
        The target semantic version (e.g., '0.5.3', '0.0.1', '1.0.0') used to determine
        the stability warning alert (Alpha, Beta, or Production).
        Defaults to the version in wails.json (or parsed from the baseline version tag).
        Aliases: -Version, -ProductVersion.

    .PARAMETER ChangelogPath
        Path to CHANGELOG.md. Defaults to the CHANGELOG.md file in the project root.

    .PARAMETER Model
        Optional Antigravity CLI model to specify via 'agy --model <model>'.

    .PARAMETER IncludeWorkingTree
        If specified, includes unstaged and staged uncommitted working tree modifications in the summary.

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
        Discovers the most recent version tag, queries commits up to HEAD, generates summary via agy,
        and prompts to update the '## [Unreleased]' section in CHANGELOG.md.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1 -DryRun
        Previews the AI-generated release notes without modifying CHANGELOG.md.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1 -SinceTag "v0.5.2" -NonInteractive
        Summarizes changes since v0.5.2 and updates CHANGELOG.md directly without confirmation.
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [Alias("FromTag")]
    [string]$SinceTag = "",

    [Parameter(Mandatory = $false)]
    [string]$ToRef = "HEAD",

    [Parameter(Mandatory = $false)]
    [Alias("Version", "ProductVersion")]
    [string]$TargetVersion = "",

    [Parameter(Mandatory = $false)]
    [string]$ChangelogPath = "",

    [Parameter(Mandatory = $false)]
    [string]$Model = "",

    [Parameter(Mandatory = $false)]
    [switch]$IncludeWorkingTree,

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

function Resolve-TargetVersion {
    param(
        [string]$ExplicitVersion,
        [string]$ProjectRoot,
        [string]$FallbackTag
    )

    if (-not [string]::IsNullOrWhiteSpace($ExplicitVersion)) {
        return $ExplicitVersion.Trim()
    }

    $wailsPath = Join-Path $ProjectRoot "wails.json"
    if (Test-Path -Path $wailsPath -PathType Leaf) {
        try {
            $raw = Get-Content -Raw -Path $wailsPath -Encoding utf8
            $json = $raw | ConvertFrom-Json
            if ($json.info -and -not [string]::IsNullOrWhiteSpace($json.info.productVersion)) {
                return $json.info.productVersion.Trim()
            }
        } catch {}
    }

    if (-not [string]::IsNullOrWhiteSpace($FallbackTag)) {
        return ($FallbackTag.Trim() -replace '^v', '')
    }

    return "0.1.0"
}

function Update-ChangelogDocument {
    param(
        [string]$ChangelogPath,
        [string]$NewUnreleasedBody,
        [string]$SinceTag = ""
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        throw "CHANGELOG.md file not found at '$ChangelogPath'."
    }

    $content = Get-Content -Path $ChangelogPath -Raw -Encoding utf8

    # Regex to find '## [Unreleased]' or '## Unreleased' section up to the next '## ' header or link definitions or end of file
    $unreleasedPattern = '(?ms)^##\s+\[?Unreleased\]?\s*\r?\n.*?(?=^##\s+[^#]|^\[.+\]:\s+https?://|\Z)'
    $formattedBlock = "## [Unreleased]`n`n" + $NewUnreleasedBody.Trim() + "`n`n"

    if ($content -match $unreleasedPattern) {
        $newContent = [regex]::Replace($content, $unreleasedPattern, [System.Text.RegularExpressions.MatchEvaluator]{ return $formattedBlock })
    } else {
        # If '## [Unreleased]' header does not exist, insert it before the first version header
        $versionPattern = '(?m)^##\s+\[?(?:v)?\d+\.\d+\.\d+'
        if ($content -match $versionPattern) {
            $newContent = [regex]::new($versionPattern).Replace($content, [System.Text.RegularExpressions.MatchEvaluator]{ return ($formattedBlock + $_.Value) }, 1)
        } else {
            # Append at end of file
            $newContent = $content.TrimEnd() + "`n`n" + $formattedBlock
        }
    }

    # Maintain compare link for [Unreleased] at the bottom
    if (-not [string]::IsNullOrWhiteSpace($SinceTag)) {
        $repoUrl = "https://github.com/zmtcreative/TextoVista"
        if ($newContent -match '(?m)^\[.+\]:\s*(https?://github\.com/[^/\s]+/[^/\s#?]+?)(?:/compare|/releases|$)') {
            $repoUrl = ($Matches[1] -replace '\.git$', '').TrimEnd('/')
        }
        $unreleasedLink = "[Unreleased]: $repoUrl/compare/$SinceTag...HEAD"
        $unreleasedLinkPattern = '(?m)^\[Unreleased\]:\s+https?://\S+'
        $versionLinkPattern = '(?m)^\[\d+\.\d+\.\d+(?:-[^\]]+)?\]:\s+https?://\S+'
        if ($newContent -match $unreleasedLinkPattern) {
            $newContent = [regex]::new($unreleasedLinkPattern).Replace($newContent, $unreleasedLink, 1)
        } elseif ($newContent -match $versionLinkPattern) {
            $newContent = [regex]::new($versionLinkPattern).Replace($newContent, "$unreleasedLink`n`$0", 1)
        } else {
            $newContent = $newContent.TrimEnd() + "`n`n$unreleasedLink`n"
        }
    }

    return ($newContent.TrimEnd() + "`n")
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

    # 3b. Determine evaluated semantic version & stability warning alert
    $evaluatedVersion = Resolve-TargetVersion -ExplicitVersion $TargetVersion -ProjectRoot $projectRoot -FallbackTag $SinceTag
    $alertMarkdown = Get-VersionAlertMarkdown -VersionString $evaluatedVersion
    $stabilityStatus = if ($alertMarkdown -match '\*\*Alpha\*\*') { "Alpha" } elseif ($alertMarkdown -match '\*\*Beta\*\*') { "Beta" } else { "Production (no alert)" }

    Write-Host -ForegroundColor Green "  Evaluated project version:          " -NoNewline
    Write-Host -ForegroundColor Yellow "$evaluatedVersion ($stabilityStatus)"
    Write-Host -ForegroundColor Green "  Target ref:                         " -NoNewline
    Write-Host -ForegroundColor Yellow "$ToRef"
    Write-Host -ForegroundColor Green "  Changelog path:                     " -NoNewline
    Write-Host -ForegroundColor Yellow "$ChangelogPath"
    Write-Host ""

    # 4. Collect git changes
    Write-Host -ForegroundColor Cyan "  Querying Git history since '$SinceTag'..."
    $gitData = Get-GitCommitHistory -FromRef $SinceTag -ToRef $ToRef -IncludeUncommitted:$IncludeWorkingTree

    Write-Host -ForegroundColor Green "  Commits found:                      " -NoNewline
    Write-Host -ForegroundColor Yellow "$($gitData.CommitCount)"

    if ($gitData.CommitCount -eq 0 -and [string]::IsNullOrWhiteSpace($gitData.UncommittedInfo)) {
        Write-Host ""
        Write-Host -ForegroundColor Yellow "Notice: No new commits or changes found between '$SinceTag' and '$ToRef'."
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
