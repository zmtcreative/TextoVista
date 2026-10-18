#Requires -Version 7.0

<#
    .SYNOPSIS
        Uses the Antigravity CLI (agy) to summarize recent Git changes since the latest version tag
        and updates the '## Unreleased' section in CHANGELOG.md.

    .DESCRIPTION
        Identifies the most recent semantic version tag (matching 'vX.Y.Z'), collects all commits
        and diff statistics up to HEAD (or a specified ref), and invokes the Antigravity CLI ('agy')
        non-interactively.

        The AI generates a high-level narrative summary and a categorized bullet list of changes
        (e.g., - Rebrand:, - Add:, - Update:, - Fix:, - CI/CD:, - Docs:), formatted consistently
        with existing entries in CHANGELOG.md. The resulting content is then cleanly inserted
        or updated under the '## Unreleased' section at the top of CHANGELOG.md.

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
        Generates and displays the proposed '## Unreleased' changelog content without modifying CHANGELOG.md.
        Alias: -WhatIf.

    .PARAMETER NonInteractive
        Applies changes to CHANGELOG.md without an interactive confirmation prompt.
        Aliases: -Force, -Yes.

    .PARAMETER ShowPrompt
        Prints the exact prompt sent to the Antigravity CLI for inspection/debugging.

    .EXAMPLE
        ./scripts/Update-Changelog.ps1
        Discovers the most recent version tag, queries commits up to HEAD, generates summary via agy,
        and prompts to update the '## Unreleased' section in CHANGELOG.md.

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

function Get-LatestSemanticVersionTag {
    # Match semantic release tags like v0.5.3 or v1.0.0
    $tag = git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --abbrev=0 2>$null
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($tag)) {
        return $tag.Trim()
    }

    # Fallback: Query all matching tags sorted by version refname
    $allTags = git tag -l "v[0-9]*.[0-9]*.[0-9]*" --sort=-v:refname 2>$null
    if ($LASTEXITCODE -eq 0 -and $allTags) {
        $first = ($allTags | Select-Object -First 1)
        if (-not [string]::IsNullOrWhiteSpace($first)) {
            return $first.Trim()
        }
    }

    return $null
}

function Get-GitCommitHistory {
    param(
        [string]$FromRef,
        [string]$ToRef,
        [bool]$IncludeUncommitted
    )

    $logFormat = "commit %h%nAuthor: %an%nDate: %ad%nSubject: %s%nBody: %b%n"
    $commitLog = git log "$FromRef..$ToRef" --reverse --date=short --format=$logFormat 2>$null
    $shortLog = git log "$FromRef..$ToRef" --reverse --date=short --pretty=format:"- %h %s (%an, %ad)" 2>$null
    $diffStat = git diff --stat "$FromRef..$ToRef" 2>$null

    $uncommittedDiff = ""
    if ($IncludeUncommitted) {
        $status = git status --short 2>$null
        $statDiff = git diff --stat HEAD 2>$null
        if (-not [string]::IsNullOrWhiteSpace($status)) {
            $uncommittedDiff = "Uncommitted working tree status:`n$status`n`nUncommitted changes diffstat:`n$statDiff"
        }
    }

    return [PSCustomObject]@{
        DetailedLog     = ($commitLog -join "`n").Trim()
        ShortLog        = ($shortLog -join "`n").Trim()
        DiffStat        = ($diffStat -join "`n").Trim()
        UncommittedInfo = $uncommittedDiff.Trim()
        CommitCount     = if ([string]::IsNullOrWhiteSpace($shortLog)) { 0 } else { @($shortLog).Count }
    }
}

function Get-ChangelogReferenceStyle {
    param([string]$ChangelogPath)

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        return ""
    }

    $lines = Get-Content -Path $ChangelogPath -Encoding utf8
    $sampleLines = @()
    $capturing = $false
    $entryCount = 0

    foreach ($line in $lines) {
        if ($line -match '^##\s+v\d+\.\d+\.\d+') {
            $entryCount++
            if ($entryCount -gt 2) { break }
            $capturing = $true
        }

        if ($capturing) {
            $sampleLines += $line
        }
    }

    return ($sampleLines -join "`n").Trim()
}

function Get-CurrentUnreleasedContent {
    param([string]$ChangelogPath)

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        return ""
    }

    $lines = Get-Content -Path $ChangelogPath -Encoding utf8
    $inUnreleased = $false
    $unreleasedLines = @()

    foreach ($line in $lines) {
        if ($line -match '^##\s+Unreleased') {
            $inUnreleased = $true
            continue
        }

        if ($inUnreleased -and $line -match '^##\s+') {
            break
        }

        if ($inUnreleased) {
            $unreleasedLines += $line
        }
    }

    return ($unreleasedLines -join "`n").Trim()
}

function Build-AgyChangelogPrompt {
    param(
        [string]$SinceTag,
        [string]$ToRef,
        [PSCustomObject]$GitData,
        [string]$ReferenceStyle,
        [string]$CurrentUnreleased
    )

    $prompt = @"
You are an expert technical release notes writer.
Your task is to summarize the recent changes in this repository into content for the '## Unreleased' section of CHANGELOG.md.

### REPOSITORY CHANGES:
Baseline Tag: $SinceTag
Target Ref:   $ToRef
Number of Commits: $($GitData.CommitCount)

Commit Log:
$($GitData.ShortLog)

Detailed Commit History:
$($GitData.DetailedLog)

Diffstat Summary:
$($GitData.DiffStat)
"@

    if (-not [string]::IsNullOrWhiteSpace($GitData.UncommittedInfo)) {
        $prompt += @"

Working Tree Changes:
$($GitData.UncommittedInfo)
"@
    }

    if (-not [string]::IsNullOrWhiteSpace($ReferenceStyle)) {
        $prompt += @"

### REFERENCE STYLE (from recent CHANGELOG.md entries in this project):
$ReferenceStyle
"@
    }

    if (-not [string]::IsNullOrWhiteSpace($CurrentUnreleased)) {
        $prompt += @"

### CURRENT CONTENT IN ## Unreleased:
$CurrentUnreleased
"@
    }

    $prompt += @"

### OUTPUT REQUIREMENTS:
1. Write 1 or 2 high-level narrative paragraphs summarizing the main achievements, improvements, fixes, or architectural updates.
2. Include a '### CHANGES' subheader.
3. Under '### CHANGES', provide a concise bullet list of the notable changes since $SinceTag.
4. Categorize each bullet item with an appropriate standard prefix consistent with this project:
   - Rebrand: (for project or binary naming changes)
   - Add: (for new features or capabilities)
   - Update: (for dependency, runtime, or component upgrades)
   - Fix: (for bug fixes and corrections)
   - CI/CD: (for GitHub Actions, workflow, or packaging changes)
   - Scripts: (for developer script improvements)
   - Docs: (for documentation updates)
5. Keep bullet descriptions clear, professional, and directly focused on user-visible or architectural value.
6. If the current '## Unreleased' section contains temporary placeholders (such as 'Ongoing development updates will be collected here until the next tagged release.' or generic placeholder notes), discard and replace them with the new summary and bullet list. Only retain existing notes if they describe actual changes or requirements not already captured.
7. CRITICAL: Do NOT output any '> [!WARNING]' alert callout or the '## Unreleased' header itself. Stability warning alerts and section headers are handled automatically by the script.
8. CRITICAL: Do NOT call any tools. Output ONLY pure markdown text without wrapping the entire response in backtick code fences (no ```markdown or ```).
"@

    return $prompt
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

function Get-VersionAlertMarkdown {
    param(
        [string]$VersionString
    )

    if ($VersionString -match '^v?(?<Major>\d+)\.(?<Minor>\d+)') {
        $major = [int]$Matches['Major']
        $minor = [int]$Matches['Minor']

        if ($major -eq 0 -and $minor -eq 0) {
            # Major == 0 and Minor == 0: Alpha software
            return @(
                "> [!WARNING]",
                "> This is still **Alpha** software! It is likely incomplete and may crash without warning. Use at your own risk!"
            ) -join "`n"
        } elseif ($major -eq 0 -and $minor -ge 1) {
            # Major == 0 and Minor >= 1: Beta software
            return @(
                "> [!WARNING]",
                "> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing."
            ) -join "`n"
        } else {
            # Major >= 1: Production software (omit warning entirely)
            return ""
        }
    }

    return ""
}

function Format-UnreleasedContent {
    param(
        [string]$SummaryMarkdown,
        [string]$AlertMarkdown
    )

    $text = $SummaryMarkdown.Trim()

    # Strip any accidental > [!WARNING] alert generated by the AI
    $text = [regex]::Replace($text, '(?ms)^>\s*\[!WARNING\].*?(?=\r?\n\r?\n|\r?\n###\s+CHANGES|\Z)', '').Trim()

    if ([string]::IsNullOrWhiteSpace($AlertMarkdown)) {
        return $text
    }

    # Position: AFTER the summary paragraph(s), but BEFORE the '### CHANGES' header
    if ($text -match '(?m)^###\s+CHANGES') {
        return [regex]::Replace($text, '(?m)^###\s+CHANGES', "$AlertMarkdown`n`n### CHANGES", 1)
    } else {
        return "$text`n`n$AlertMarkdown"
    }
}

function Invoke-AgyChangelogSummary {
    param(
        [string]$Prompt,
        [string]$Model
    )

    $agyArgs = @("--print", $Prompt, "--disable-slash-commands")
    if (-not [string]::IsNullOrWhiteSpace($Model)) {
        $agyArgs += @("--model", $Model)
    }

    Write-Host -ForegroundColor Cyan "  Invoking Antigravity CLI (agy)..."
    $response = & agy @agyArgs 2>&1
    $exitCode = $LASTEXITCODE

    if ($exitCode -ne 0) {
        $errText = ($response -join "`n")
        throw "Antigravity CLI failed with exit code $($exitCode):`n$errText"
    }

    $rawText = ($response -join "`n").Trim()

    # Clean any accidental outer markdown fences if returned
    if ($rawText -match '^```(?:markdown)?\s*([\s\S]*?)\s*```$') {
        $rawText = $Matches[1].Trim()
    }

    return $rawText
}

function Update-ChangelogDocument {
    param(
        [string]$ChangelogPath,
        [string]$NewUnreleasedBody
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        throw "CHANGELOG.md file not found at '$ChangelogPath'."
    }

    $content = Get-Content -Path $ChangelogPath -Raw -Encoding utf8

    # Regex to find '## Unreleased' section up to the next '## ' header or end of file
    $unreleasedPattern = '(?ms)^##\s+Unreleased\s*\r?\n.*?(?=^##\s+[^#]|\Z)'
    $formattedBlock = "## Unreleased`n`n" + $NewUnreleasedBody.Trim() + "`n`n"

    if ($content -match $unreleasedPattern) {
        $newContent = [regex]::Replace($content, $unreleasedPattern, [System.Text.RegularExpressions.MatchEvaluator]{ return $formattedBlock })
    } else {
        # If '## Unreleased' header does not exist, insert it before the first version header
        $versionPattern = '(?m)^##\s+v\d+\.\d+\.\d+'
        if ($content -match $versionPattern) {
            $newContent = [regex]::Replace($content, $versionPattern, [System.Text.RegularExpressions.MatchEvaluator]{ return ($formattedBlock + $_.Value) }, 1)
        } else {
            # Append at end of file
            $newContent = $content.TrimEnd() + "`n`n" + $formattedBlock
        }
    }

    return $newContent
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
        Write-Host -ForegroundColor Yellow "The '## Unreleased' section does not need updating at this time."
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

    # 8. Format section with stability warning alert positioned after summary and before ### CHANGES
    $unreleasedBody = Format-UnreleasedContent -SummaryMarkdown $summaryMarkdown -AlertMarkdown $alertMarkdown

    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host -ForegroundColor Cyan "               Proposed '## Unreleased' Content                   "
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""
    Write-Host -ForegroundColor White "## Unreleased"
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
    $updatedDocument = Update-ChangelogDocument -ChangelogPath $ChangelogPath -NewUnreleasedBody $unreleasedBody
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
