#Requires -Version 7.2

<#
    .SYNOPSIS
        Streamlined release orchestrator for TextoVista: prepares release notes via AI,
        updates CHANGELOG.md, bumps version across project files, and creates annotated Git tags.

    .DESCRIPTION
        Orchestrates an atomic, deferred-execution release workflow:
        1. Ensures the Git working tree is completely clean before proceeding.
        2. Resolves the target version via explicit SemVer (-Version) or semantic bump (-Bump).
        3. Queries commits since the latest version tag and drafts Keep a Changelog 1.1.0 release notes
           using the Antigravity CLI ('agy'), incorporating any existing notes under '## [Unreleased]'.
        4. Calculates in-memory file diffs for wails.json, package.json, package-lock.json, info.json,
           project.nsi, and CHANGELOG.md.
        5. Displays a unified preview of the release changes and prompts for confirmation.
           - If aborted, ZERO disk changes, Git commits, or tags are created.
           - If approved, atomically applies all updates, creates a single release commit,
             and creates an annotated Git tag locally.
        6. Prompts to push the branch and new tag to the remote repository.

    .PARAMETER Version
        Explicit target semantic version (e.g. '0.8.0', '1.0.0-rc1').
        Mutually exclusive with -Bump.
        Aliases: -Ver.

    .PARAMETER Bump
        Semantic version bump action using standard SemVer / uv semantics:
        'major', 'minor', 'patch', 'stable', 'alpha', 'beta', or 'rc'.
        Mutually exclusive with -Version.

    .PARAMETER ChangelogPath
        Path to CHANGELOG.md. Defaults to CHANGELOG.md in the project root.

    .PARAMETER Model
        Optional Antigravity CLI model to pass to 'agy --model <model>'.

    .PARAMETER SkipChangelog
        If specified, bypasses AI changelog generation and maintains existing changelog content.

    .PARAMETER ShowPrompt
        Displays the exact prompt sent to the Antigravity CLI for inspection/debugging.

    .PARAMETER NonInteractive
        Applies changes and creates tag without interactive confirmation prompts.
        Aliases: -Force, -Yes.

    .EXAMPLE
        ./scripts/Set-Version.ps1 -Bump patch
        Calculates the next patch release, drafts release notes, previews changes,
        and prompts to commit and tag.

    .EXAMPLE
        ./scripts/Set-Version.ps1 -Bump minor
        Bumps the minor version (e.g. 0.7.0 -> 0.8.0), updates all files, and tags.

    .EXAMPLE
        ./scripts/Set-Version.ps1 -Version "0.8.0"
        Sets the version explicitly to 0.8.0.

    .EXAMPLE
        ./scripts/Set-Version.ps1 -Bump beta
        Initiates or increments a beta prerelease (e.g. 0.7.0 -> 0.7.1-beta1, or 0.7.1-beta1 -> 0.7.1-beta2).
#>

[CmdletBinding(SupportsShouldProcess)]
param(
    [Parameter(ParameterSetName = "ExplicitVersion", Mandatory = $true, Position = 0)]
    [Alias("Ver")]
    [string]$Version,

    [Parameter(ParameterSetName = "BumpStep", Mandatory = $true, Position = 0)]
    [ValidateSet("major", "minor", "patch", "stable", "alpha", "beta", "rc")]
    [string]$Bump,

    [Parameter(Mandatory = $false)]
    [string]$ChangelogPath = "",

    [Parameter(Mandatory = $false)]
    [string]$Model = "",

    [Parameter(Mandatory = $false)]
    [switch]$SkipChangelog,

    [Parameter(Mandatory = $false)]
    [switch]$AllowDirty,

    [Parameter(Mandatory = $false)]
    [switch]$ShowPrompt,

    [Parameter(Mandatory = $false)]
    [Alias("Force", "Yes")]
    [switch]$NonInteractive
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# --- Project Paths Resolution ---
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

# --- Target Files Definition ---
$TargetFilesConfig = @(
    @{
        RelPath = "wails.json"
        Format  = "SemVer"
        Rules   = @(
            @{ Pattern = '("productVersion"\s*:\s*")[^"]+(")'; Replacement = '${1}{VALUE}$2' }
        )
    },
    @{
        RelPath = "frontend/package.json"
        Format  = "SemVer"
        Rules   = @(
            @{ Pattern = '("version"\s*:\s*")[^"]+(")'; Replacement = '${1}{VALUE}$2' }
        )
    },
    @{
        RelPath = "build/windows/info.json"
        Format  = "WindowsNumeric"
        Rules   = @(
            @{ Pattern = '("file_version"\s*:\s*")[^"]+(")'; Replacement = '${1}{VALUE}$2' }
        )
    },
    @{
        RelPath = "build/windows/installer/project.nsi"
        Format  = "WindowsNumeric"
        Rules   = @(
            @{ Pattern = '(?m)^(?<before>VIProductVersion\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' },
            @{ Pattern = '(?m)^(?<before>VIFileVersion\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' },
            @{ Pattern = '(?m)^(?<before>VIAddVersionKey\s+"FileVersion"\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' }
        )
    }
)

# ==============================================================================
# Helper Functions: Pre-Flight & Environment Checks
# ==============================================================================

function Assert-Prerequisites {
    param(
        [string]$ProjectRoot,
        [bool]$NeedAgy
    )

    if (-not (Test-Path -Path (Join-Path $ProjectRoot "wails.json") -PathType Leaf)) {
        throw "Could not locate project root (wails.json not found in '$ProjectRoot')."
    }

    if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) {
        throw "Git command-line utility ('git') is not installed or not in PATH."
    }

    $insideWorkTree = git rev-parse --is-inside-work-tree 2>$null
    if ($LASTEXITCODE -ne 0 -or $insideWorkTree.Trim() -ne "true") {
        throw "The project directory '$ProjectRoot' is not a valid Git repository."
    }

    if (-not (Get-Command "npm" -ErrorAction SilentlyContinue)) {
        throw "npm command-line utility was not found in PATH. Node.js/npm is required to sync package-lock.json."
    }

    if ($NeedAgy -and -not (Get-Command "agy" -ErrorAction SilentlyContinue)) {
        throw "Antigravity CLI ('agy') was not found in PATH.`nPlease install agy or use -SkipChangelog to proceed without AI changelog generation."
    }
}

function Assert-GitWorkingTreeClean {
    $status = git status --porcelain 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to query Git working tree status."
    }

    if (-not [string]::IsNullOrWhiteSpace($status)) {
        Write-Host ""
        Write-Host -ForegroundColor Red "ERROR: Git working tree contains uncommitted changes."
        Write-Host -ForegroundColor Yellow "The release workflow requires a clean working tree so release commits remain atomic."
        Write-Host ""
        Write-Host -ForegroundColor White "Uncommitted files:"
        $status.Trim() -split "`r?`n" | ForEach-Object { Write-Host -ForegroundColor DarkYellow "  $_" }
        Write-Host ""
        Write-Host -ForegroundColor Cyan "Please commit or stash your changes before running Set-Version.ps1."
        Write-Host ""
        throw "Git working tree is not clean."
    }
}

# ==============================================================================
# Helper Functions: Version Parsing & Calculations
# ==============================================================================

function Get-CurrentVersionFromWails {
    param([string]$ProjectRoot)
    $wailsPath = Join-Path $ProjectRoot "wails.json"
    if (-not (Test-Path -Path $wailsPath -PathType Leaf)) {
        throw "Source of truth 'wails.json' not found at '$wailsPath'."
    }
    $raw = Get-Content -Raw -Path $wailsPath -Encoding utf8
    $json = $raw | ConvertFrom-Json
    if (-not $json.info -or [string]::IsNullOrWhiteSpace($json.info.productVersion)) {
        throw "Could not determine 'info.productVersion' from wails.json."
    }
    return $json.info.productVersion.Trim()
}

function Parse-VersionString {
    param([string]$Tag)
    $clean = $Tag.Trim() -replace '^[vV]', ''
    $obj = [PSCustomObject]@{
        Raw                = $Tag
        Major              = 0
        Minor              = 0
        Patch              = 0
        Prerelease         = $null
        OriginalPrerelease = $null
        WasNormalized      = $false
        Stage              = $null
        StageNum           = $null
        ValidationError    = $null
        IsValid            = $false
    }

    if ($clean -match '\+') {
        $obj.ValidationError = "Build metadata ('+...') is not permitted in versions for Windows packaging compatibility."
        return $obj
    }

    if ($clean -notmatch '^(?<major>0|[1-9]\d*)\.(?<minor>0|[1-9]\d*)\.(?<patch>0|[1-9]\d*)(?:-(?<prerelease>.+))?$') {
        $obj.ValidationError = "Version '$Tag' does not match semantic version format (expected 'X.Y.Z' or 'vX.Y.Z', e.g. '0.8.0' or 'v0.7.3-rc.1')."
        return $obj
    }

    $obj.Major = [int]$Matches['major']
    $obj.Minor = [int]$Matches['minor']
    $obj.Patch = [int]$Matches['patch']

    if (-not $Matches.ContainsKey('prerelease') -or [string]::IsNullOrWhiteSpace($Matches['prerelease'])) {
        $obj.IsValid = $true
        return $obj
    }

    $prerelease = $Matches['prerelease']
    # Allowed prerelease formats: alpha, beta, rc with optional delimiter and sequence number (e.g., -alpha.1, -alpha1, -beta2, -rc.1, -rc1)
    if ($prerelease -notmatch '^(?<stage>alpha|beta|rc)(?:[.-]?(?<num>0|[1-9]\d*))?$') {
        $obj.ValidationError = "Prerelease '$prerelease' is not permitted. Only 'alpha', 'beta', or 'rc' stages (e.g. 'alpha.1', 'beta.1', 'rc.1') are allowed."
        return $obj
    }

    $stage = $Matches['stage'].ToLower()
    $num = if ($Matches.ContainsKey('num') -and -not [string]::IsNullOrWhiteSpace($Matches['num'])) { [int]$Matches['num'] } else { 0 }

    if ($num -gt 999) {
        $obj.ValidationError = "Prerelease number '$num' exceeds maximum allowed sequence of 999."
        return $obj
    }

    # Normalize to SemVer 2.0.0 dot-separated format: stage.num (e.g. rc.1, beta.2)
    $normalizedPrerelease = if ($num -gt 0) { "${stage}.${num}" } else { $stage }
    $wasNormalized = ($prerelease -ne $normalizedPrerelease)

    $obj.Prerelease = $normalizedPrerelease
    $obj.OriginalPrerelease = $prerelease
    $obj.WasNormalized = $wasNormalized
    $obj.Stage = $stage
    $obj.StageNum = $num
    $obj.IsValid = $true
    return $obj
}

function Format-BuildVersions {
    param($vObj)
    $semver = "$($vObj.Major).$($vObj.Minor).$($vObj.Patch)"
    if ($vObj.Prerelease) {
        $semver += "-$($vObj.Prerelease)"
    }
    $tagName = "v$semver"
    $tagMessage = "Version $semver"

    # Windows 4-part numeric version: Major.Minor.Patch.Suffix
    # Final releases use fixed 65432 so they always compare as newer than any prerelease.
    # Prereleases: alpha -> 1000 + n, beta -> 2000 + n, rc -> 3000 + n
    $stages = @('alpha', 'beta', 'rc')
    $fileVersionSuffix = 65432
    if ($vObj.Prerelease) {
        if ($vObj.Stage) {
            $stageIndex = $stages.IndexOf($vObj.Stage) + 1
            if ($stageIndex -gt 0) {
                $num = if ($null -ne $vObj.StageNum) { [int]$vObj.StageNum } else { 0 }
                $fileVersionSuffix = ($stageIndex * 1000) + $num
            } else {
                $fileVersionSuffix = 1000
            }
        } else {
            $fileVersionSuffix = 1000
        }
    }

    $fileVersion = "$($vObj.Major).$($vObj.Minor).$($vObj.Patch).$fileVersionSuffix"

    return [PSCustomObject]@{
        SemVer      = $semver
        TagName     = $tagName
        TagMessage  = $tagMessage
        FileVersion = $fileVersion
        Major       = $vObj.Major
        Minor       = $vObj.Minor
        Patch       = $vObj.Patch
        Prerelease  = $vObj.Prerelease
        Stage       = $vObj.Stage
        StageNum    = $vObj.StageNum
    }
}

function Calculate-NextVersion {
    param(
        $CurrentObj,
        [string]$ExplicitVersion = "",
        [string]$BumpAction = ""
    )

    if (-not [string]::IsNullOrWhiteSpace($ExplicitVersion)) {
        $parsed = Parse-VersionString -Tag $ExplicitVersion
        if (-not $parsed.IsValid) {
            throw "Invalid explicit version provided: '$ExplicitVersion'. $($parsed.ValidationError)"
        }

        if ($parsed.WasNormalized) {
            Write-Host -ForegroundColor Yellow "  Notice: Normalized non-standard pre-release identifier '$($parsed.OriginalPrerelease)' to SemVer 2.0.0 standard '$($parsed.Prerelease)'."
        }

        # Check for major version jump warning
        if ($parsed.Major -gt ($CurrentObj.Major + 1)) {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "WARNING: Significant major version jump detected!"
            Write-Host -ForegroundColor Yellow "  Current version:  $($CurrentObj.Major).$($CurrentObj.Minor).$($CurrentObj.Patch)"
            Write-Host -ForegroundColor Yellow "  Proposed version: $($parsed.Major).$($parsed.Minor).$($parsed.Patch)"
            Write-Host ""
            $confirmJump = Read-Host "Are you sure you want to jump more than one major version? (y/N)"
            if ($confirmJump -notmatch '^(y|yes)$') {
                throw "Version bump aborted by user due to major version jump."
            }
        }

        return (Format-BuildVersions -vObj $parsed)
    }

    $major = $CurrentObj.Major
    $minor = $CurrentObj.Minor
    $patch = $CurrentObj.Patch
    $prerelease = $null
    $stage = $null
    $stageNum = $null

    switch ($BumpAction.ToLower()) {
        "major" {
            $major++
            $minor = 0
            $patch = 0
        }
        "minor" {
            $minor++
            $patch = 0
        }
        "patch" {
            if ($CurrentObj.Prerelease) {
                # In SemVer/uv, bumping patch on a prerelease promotes to the stable version
                $patch = $CurrentObj.Patch
            } else {
                $patch++
            }
        }
        "stable" {
            if (-not $CurrentObj.Prerelease) {
                throw "Current version '$($CurrentObj.Major).$($CurrentObj.Minor).$($CurrentObj.Patch)' is already a stable release."
            }
            $patch = $CurrentObj.Patch
        }
        "alpha" {
            if (-not $CurrentObj.Prerelease) {
                $patch++
                $stage = "alpha"
                $stageNum = 1
            } elseif ($CurrentObj.Stage -eq "alpha") {
                $stage = "alpha"
                $stageNum = if ($CurrentObj.StageNum) { $CurrentObj.StageNum + 1 } else { 1 }
            } else {
                $stage = "alpha"
                $stageNum = 1
            }
            $prerelease = "alpha.$stageNum"
        }
        "beta" {
            if (-not $CurrentObj.Prerelease) {
                $patch++
                $stage = "beta"
                $stageNum = 1
            } elseif ($CurrentObj.Stage -eq "beta") {
                $stage = "beta"
                $stageNum = if ($CurrentObj.StageNum) { $CurrentObj.StageNum + 1 } else { 1 }
            } else {
                $stage = "beta"
                $stageNum = 1
            }
            $prerelease = "beta.$stageNum"
        }
        "rc" {
            if (-not $CurrentObj.Prerelease) {
                $patch++
                $stage = "rc"
                $stageNum = 1
            } elseif ($CurrentObj.Stage -eq "rc") {
                $stage = "rc"
                $stageNum = if ($CurrentObj.StageNum) { $CurrentObj.StageNum + 1 } else { 1 }
            } else {
                $stage = "rc"
                $stageNum = 1
            }
            $prerelease = "rc.$stageNum"
        }
        default {
            throw "Unknown bump action '$BumpAction'."
        }
    }

    $vObj = [PSCustomObject]@{
        Major      = $major
        Minor      = $minor
        Patch      = $patch
        Prerelease = $prerelease
        Stage      = $stage
        StageNum   = $stageNum
    }

    return (Format-BuildVersions -vObj $vObj)
}

# ==============================================================================
# Helper Functions: Git & Changelog Processing
# ==============================================================================

function Get-LatestSemanticVersionTag {
    $tag = git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" --abbrev=0 2>$null
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($tag)) {
        return $tag.Trim()
    }
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
    param([string]$FromRef)

    $logFormat = "commit %h%nAuthor: %an%nDate: %ad%nSubject: %s%nBody: %b%n"
    $commitLog = git log "$FromRef..HEAD" --reverse --date=short --format=$logFormat 2>$null
    $shortLog = git log "$FromRef..HEAD" --reverse --date=short --pretty=format:"- %h %s (%an, %ad)" 2>$null
    $diffStat = git diff --stat "$FromRef..HEAD" 2>$null

    return [PSCustomObject]@{
        DetailedLog = ($commitLog -join "`n").Trim()
        ShortLog    = ($shortLog -join "`n").Trim()
        DiffStat    = ($diffStat -join "`n").Trim()
        CommitCount = if ([string]::IsNullOrWhiteSpace($shortLog)) { 0 } else { @($shortLog).Count }
    }
}

function Get-ChangelogReferenceStyle {
    param([string]$ChangelogPath)
    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) { return "" }
    $lines = Get-Content -Path $ChangelogPath -Encoding utf8
    $sampleLines = @()
    $capturing = $false
    $entryCount = 0

    foreach ($line in $lines) {
        if ($line -match '^##\s+\[?(?:v)?\d+\.\d+\.\d+\]?') {
            $entryCount++
            if ($entryCount -gt 2) { break }
            $capturing = $true
        }
        if ($capturing) {
            if ($line -match '^\[.+\]:\s+https?://') { break }
            $sampleLines += $line
        }
    }
    return ($sampleLines -join "`n").Trim()
}

function Get-CurrentUnreleasedContent {
    param([string]$ChangelogPath)
    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) { return "" }
    $lines = Get-Content -Path $ChangelogPath -Encoding utf8
    $inUnreleased = $false
    $unreleasedLines = @()

    foreach ($line in $lines) {
        if ($line -match '^##\s+\[?Unreleased\]?') {
            $inUnreleased = $true
            continue
        }
        if ($inUnreleased -and ($line -match '^##\s+' -or $line -match '^\[.+\]:\s+https?://')) {
            break
        }
        if ($inUnreleased) {
            $unreleasedLines += $line
        }
    }
    return ($unreleasedLines -join "`n").Trim()
}

function Get-VersionAlertMarkdown {
    param([string]$VersionString)
    if ($VersionString -match '^v?(?<Major>\d+)\.(?<Minor>\d+)') {
        $major = [int]$Matches['Major']
        $minor = [int]$Matches['Minor']
        if ($major -eq 0 -and $minor -eq 0) {
            return @(
                "> [!WARNING]",
                "> This is still **Alpha** software! It is likely incomplete and may crash without warning. Use at your own risk!"
            ) -join "`n"
        } elseif ($major -eq 0 -and $minor -ge 1) {
            return @(
                "> [!WARNING]",
                "> This is still **Beta** software! It should be functional for general usage but likely still has bugs, and development is still ongoing."
            ) -join "`n"
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
    # Strip any accidental alert callouts from the AI response
    $text = [regex]::Replace($text, '(?ms)^\s*>\s*\[!WARNING\].*?(?=\r?\n\r?\n|\r?\n###|\Z)', '').Trim()

    if ([string]::IsNullOrWhiteSpace($AlertMarkdown)) {
        return $text
    }

    $categoryPattern = '(?m)^###\s+(?:Added|Changed|Deprecated|Removed|Fixed|Security|CHANGES)\b'
    if ($text -match $categoryPattern) {
        return [regex]::new($categoryPattern).Replace($text, "$AlertMarkdown`n`n`$0", 1)
    } else {
        return "$text`n`n$AlertMarkdown"
    }
}

function Build-AgyChangelogPrompt {
    param(
        [string]$SinceTag,
        [string]$TargetVersion,
        [PSCustomObject]$GitData,
        [string]$ReferenceStyle,
        [string]$CurrentUnreleased
    )

    $prompt = @"
You are an expert technical release notes writer.
Your task is to summarize the recent changes in this repository into content for the '## [$TargetVersion]' release section of CHANGELOG.md adhering strictly to the Keep a Changelog (1.1.0) format.

### REPOSITORY CHANGES:
Baseline Tag:   $SinceTag
Target Version: $TargetVersion
Number of Commits: $($GitData.CommitCount)

Commit Log:
$($GitData.ShortLog)

Detailed Commit History:
$($GitData.DetailedLog)

Diffstat Summary:
$($GitData.DiffStat)
"@

    if (-not [string]::IsNullOrWhiteSpace($ReferenceStyle)) {
        $prompt += @"

### REFERENCE STYLE (from recent CHANGELOG.md entries in this project):
$ReferenceStyle
"@
    }

    if (-not [string]::IsNullOrWhiteSpace($CurrentUnreleased)) {
        $prompt += @"

### CURRENT NOTES IN ## [Unreleased] (incorporate any non-placeholder notes describing real changes):
$CurrentUnreleased
"@
    }

    $prompt += @"

### OUTPUT REQUIREMENTS:
1. Write 1 or 2 high-level narrative paragraphs summarizing the main achievements, improvements, fixes, or architectural updates.
2. Group notable changes under standard Keep a Changelog (1.1.0) level-3 subheadings:
   - ### Added (for new features or capabilities)
   - ### Changed (for changes in existing functionality, rebrandings, updates, or refactorings)
   - ### Deprecated (for soon-to-be removed features)
   - ### Removed (for now removed features)
   - ### Fixed (for any bug fixes)
   - ### Security (in case of vulnerabilities)
   Only include subheadings that have actual items.
3. Under each subheader, provide concise, high-quality bullet points describing the changes.
4. Do NOT use redundant prefixes like '- Add:', '- Fix:', '- Update:', '- Rebrand:', etc. in bullet items. The category subheader already conveys the type of change. Start bullet points directly with an action verb or noun.
5. Keep bullet descriptions clear, professional, and directly focused on user-visible or architectural value.
6. CRITICAL: Do NOT output any '> [!WARNING]' alert callout or section headers (e.g., '## [$TargetVersion]'). Stability warning alerts, section headers, and compare link definitions are handled automatically by the script.
7. CRITICAL: Do NOT call any tools. Output ONLY pure markdown text without wrapping the entire response in backtick code fences (no ```markdown or ```).
"@

    return $prompt
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

    Write-Host -ForegroundColor Cyan "  Invoking Antigravity CLI (agy) to draft release notes..."
    $response = & agy @agyArgs 2>&1
    $exitCode = $LASTEXITCODE

    if ($exitCode -ne 0) {
        $errText = ($response -join "`n")
        throw "Antigravity CLI failed with exit code $($exitCode):`n$errText"
    }

    $rawText = ($response -join "`n").Trim()
    if ($rawText -match '^```(?:markdown)?\s*([\s\S]*?)\s*```$') {
        $rawText = $Matches[1].Trim()
    }

    return $rawText
}

function Build-UpdatedChangelog {
    param(
        [string]$ChangelogPath,
        [string]$ReleaseBody,
        [string]$TargetSemVer,
        [string]$TargetTagName,
        [string]$LastTagName,
        [string]$ReleaseDate
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        throw "CHANGELOG.md file not found at '$ChangelogPath'."
    }

    $content = Get-Content -Path $ChangelogPath -Raw -Encoding utf8
    $newReleaseBlock = "## [Unreleased]`n`n## [$TargetSemVer] - $ReleaseDate`n`n" + $ReleaseBody.Trim() + "`n`n"

    $unreleasedPattern = '(?ms)^##\s+\[?Unreleased\]?\s*\r?\n.*?(?=^##\s+\[?(?:v)?\d+\.\d+\.\d+|^\[.+\]:\s+https?://|\Z)'
    if ($content -match $unreleasedPattern) {
        $newContent = [regex]::new($unreleasedPattern).Replace($content, [System.Text.RegularExpressions.MatchEvaluator]{ return $newReleaseBlock }, 1)
    } else {
        $firstVerPattern = '(?m)^##\s+\[?(?:v)?\d+\.\d+\.\d+'
        if ($content -match $firstVerPattern) {
            $newContent = [regex]::new($firstVerPattern).Replace($content, [System.Text.RegularExpressions.MatchEvaluator]{ return ($newReleaseBlock + $_.Value) }, 1)
        } else {
            $newContent = $content.TrimEnd() + "`n`n" + $newReleaseBlock
        }
    }

    # Update comparison links at bottom
    $repoUrl = "https://github.com/zmtcreative/TextoVista"
    if ($newContent -match '(?m)^\[.+\]:\s*(https?://github\.com/[^/\s]+/[^/\s#?]+?)(?:/compare|/releases|$)') {
        $repoUrl = ($Matches[1] -replace '\.git$', '').TrimEnd('/')
    }

    $unreleasedLink = "[Unreleased]: $repoUrl/compare/$TargetTagName...HEAD"
    $newReleaseLink = if (-not [string]::IsNullOrWhiteSpace($LastTagName)) {
        "[$TargetSemVer]: $repoUrl/compare/$LastTagName...$TargetTagName"
    } else {
        "[$TargetSemVer]: $repoUrl/releases/tag/$TargetTagName"
    }
    $combinedLinks = "$unreleasedLink`n$newReleaseLink"

    $unreleasedLinkPattern = '(?m)^\[Unreleased\]:\s+https?://\S+'
    $versionLinkPattern = '(?m)^\[\d+\.\d+\.\d+(?:-[^\]]+)?\]:\s+https?://\S+'

    if ($newContent -match $unreleasedLinkPattern) {
        $newContent = [regex]::new($unreleasedLinkPattern).Replace($newContent, $combinedLinks, 1)
    } elseif ($newContent -match $versionLinkPattern) {
        $newContent = [regex]::new($versionLinkPattern).Replace($newContent, "$combinedLinks`n`$0", 1)
    } else {
        $newContent = $newContent.TrimEnd() + "`n`n$combinedLinks`n"
    }

    return ($newContent.TrimEnd() + "`n")
}

# ==============================================================================
# Helper Functions: In-Memory File Diff Planning
# ==============================================================================

function Get-TargetFileDiffs {
    param(
        [string]$ProjectRoot,
        [string]$SemVer,
        [string]$FileVersion
    )

    $diffPlans = @()
    foreach ($target in $script:TargetFilesConfig) {
        $fullPath = Join-Path $ProjectRoot $target.RelPath
        if (-not (Test-Path -Path $fullPath -PathType Leaf)) {
            throw "Target file '$fullPath' was not found."
        }

        $val = if ($target.Format -eq "SemVer") { $SemVer } else { $FileVersion }
        $content = Get-Content -Raw -Path $fullPath -Encoding utf8
        $newContent = $content

        foreach ($rule in $target.Rules) {
            $repPattern = $rule.Replacement.Replace('{VALUE}', $val)
            $newContent = [regex]::Replace($newContent, $rule.Pattern, $repPattern)
        }

        $lineDiffs = @()
        if ($newContent -ne $content) {
            $origLines = $content -split "`r?`n"
            $updLines = $newContent -split "`r?`n"
            for ($i = 0; $i -lt $origLines.Count; $i++) {
                if ($i -ge $updLines.Count -or $origLines[$i] -ne $updLines[$i]) {
                    $newVal = if ($i -lt $updLines.Count) { $updLines[$i] } else { "<deleted>" }
                    $lineDiffs += [PSCustomObject]@{
                        LineNumber = $i + 1
                        Old        = $origLines[$i].Trim()
                        New        = $newVal.Trim()
                    }
                }
            }
        }

        $diffPlans += [PSCustomObject]@{
            RelPath    = $target.RelPath
            FullPath   = $fullPath
            Format     = $target.Format
            TargetVal  = $val
            IsModified = ($newContent -ne $content)
            NewContent = $newContent
            Diffs      = $lineDiffs
        }
    }

    return $diffPlans
}

# ==============================================================================
# Main Execution Flow
# ==============================================================================

try {
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host -ForegroundColor Cyan "           TextoVista - Release & Version Orchestrator            "
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""

    # 1. Resolve Project Root
    $projectRoot = (Resolve-Path $script:ScriptProjectRoot).Path
    if ([string]::IsNullOrWhiteSpace($ChangelogPath)) {
        $ChangelogPath = Join-Path $projectRoot "CHANGELOG.md"
    }
    $ChangelogPath = (Resolve-Path $ChangelogPath).Path

    # 2. Check Prerequisites & Git State
    Assert-Prerequisites -ProjectRoot $projectRoot -NeedAgy:(-not $SkipChangelog)
    if (-not $AllowDirty) {
        Assert-GitWorkingTreeClean
    } else {
        Write-Host -ForegroundColor Yellow "  Warning: -AllowDirty specified; skipping clean working tree validation."
    }

    # 3. Extract Current Version from Source of Truth (wails.json)
    $currentRawVersion = Get-CurrentVersionFromWails -ProjectRoot $projectRoot
    $currentParsed = Parse-VersionString -Tag $currentRawVersion
    if (-not $currentParsed.IsValid) {
        throw "Current version in wails.json ('$currentRawVersion') is invalid: $($currentParsed.ValidationError)"
    }
    $currentVersionInfo = Format-BuildVersions -vObj $currentParsed

    # 4. Calculate Target Version
    $targetVersionInfo = Calculate-NextVersion `
        -CurrentObj $currentVersionInfo `
        -ExplicitVersion $Version `
        -BumpAction $Bump

    # Check for Git Tag Collision
    $existingTag = git tag -l $targetVersionInfo.TagName 2>$null
    if ($existingTag) {
        throw "Git tag '$($targetVersionInfo.TagName)' already exists locally. Cannot release duplicate version."
    }

    Write-Host -ForegroundColor Green "  Current Version:  " -NoNewline
    Write-Host -ForegroundColor Yellow "$($currentVersionInfo.SemVer) ($($currentVersionInfo.TagName))"
    Write-Host -ForegroundColor Green "  Target Version:   " -NoNewline
    Write-Host -ForegroundColor Yellow "$($targetVersionInfo.SemVer) ($($targetVersionInfo.TagName))"
    Write-Host -ForegroundColor Green "  Windows Version:  " -NoNewline
    Write-Host -ForegroundColor Yellow "$($targetVersionInfo.FileVersion)"
    Write-Host ""

    # 5. Query Commits Since Baseline Tag
    $lastTag = Get-LatestSemanticVersionTag
    $baselineTag = if (-not [string]::IsNullOrWhiteSpace($lastTag)) { $lastTag } else { "" }

    Write-Host -ForegroundColor Cyan "  Querying Git history since '$baselineTag'..."
    $gitData = Get-GitCommitHistory -FromRef $baselineTag

    Write-Host -ForegroundColor Green "  Commits found:    " -NoNewline
    Write-Host -ForegroundColor Yellow "$($gitData.CommitCount)"
    Write-Host ""

    if ($gitData.CommitCount -eq 0) {
        Write-Host -ForegroundColor Yellow "Notice: No new commits found between '$baselineTag' and HEAD."
        if (-not $NonInteractive) {
            $proceedZero = Read-Host "Do you want to proceed with the release anyway? (y/N) [Default: N]"
            if ($proceedZero -notmatch '^(y|yes)$') {
                Write-Host ""
                Write-Host -ForegroundColor Yellow "Release aborted by user: no new commits to release."
                Write-Host ""
                exit 0
            }
        }
    }

    # 6. Draft Release Notes (In-Memory)
    $todayDate = (Get-Date).ToString("yyyy-MM-dd")
    $alertMarkdown = Get-VersionAlertMarkdown -VersionString $targetVersionInfo.SemVer
    $releaseNotesBody = ""

    if (-not $SkipChangelog) {
        $referenceStyle = Get-ChangelogReferenceStyle -ChangelogPath $ChangelogPath
        $currentUnreleased = Get-CurrentUnreleasedContent -ChangelogPath $ChangelogPath

        $prompt = Build-AgyChangelogPrompt `
            -SinceTag $baselineTag `
            -TargetVersion $targetVersionInfo.SemVer `
            -GitData $gitData `
            -ReferenceStyle $referenceStyle `
            -CurrentUnreleased $currentUnreleased

        if ($ShowPrompt) {
            Write-Host -ForegroundColor DarkCyan "---------------------- [AGY PROMPT] ----------------------"
            Write-Host $prompt
            Write-Host -ForegroundColor DarkCyan "----------------------------------------------------------"
            Write-Host ""
        }

        $summaryRaw = Invoke-AgyChangelogSummary -Prompt $prompt -Model $Model
        if ([string]::IsNullOrWhiteSpace($summaryRaw)) {
            throw "Antigravity CLI returned empty release notes."
        }

        $releaseNotesBody = Format-UnreleasedContent -SummaryMarkdown $summaryRaw -AlertMarkdown $alertMarkdown
    } else {
        $currentUnreleased = Get-CurrentUnreleasedContent -ChangelogPath $ChangelogPath
        if (-not [string]::IsNullOrWhiteSpace($currentUnreleased)) {
            $releaseNotesBody = Format-UnreleasedContent -SummaryMarkdown $currentUnreleased -AlertMarkdown $alertMarkdown
        } else {
            $releaseNotesBody = "$alertMarkdown`n`n### Changed`n`n- Maintenance and version bump to $($targetVersionInfo.SemVer)."
        }
    }

    # 7. Plan In-Memory Diffs for Target Files & CHANGELOG.md
    $diffPlans = Get-TargetFileDiffs `
        -ProjectRoot $projectRoot `
        -SemVer $targetVersionInfo.SemVer `
        -FileVersion $targetVersionInfo.FileVersion

    $updatedChangelogContent = Build-UpdatedChangelog `
        -ChangelogPath $ChangelogPath `
        -ReleaseBody $releaseNotesBody `
        -TargetSemVer $targetVersionInfo.SemVer `
        -TargetTagName $targetVersionInfo.TagName `
        -LastTagName $baselineTag `
        -ReleaseDate $todayDate

    $commitMessage = "chore(release): bump version to $($targetVersionInfo.SemVer)"

    # 8. Unified Release Review Display
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host -ForegroundColor Cyan "                 TextoVista - Release Preview                     "
    Write-Host -ForegroundColor Cyan "=================================================================="
    Write-Host ""
    Write-Host -ForegroundColor White "  Release Version: " -NoNewline
    Write-Host -ForegroundColor Green "$($targetVersionInfo.SemVer) ($($targetVersionInfo.TagName))"
    Write-Host -ForegroundColor White "  Windows Version: " -NoNewline
    Write-Host -ForegroundColor Green "$($targetVersionInfo.FileVersion)"
    Write-Host -ForegroundColor White "  Git Commit:      " -NoNewline
    Write-Host -ForegroundColor Yellow "$commitMessage"
    Write-Host -ForegroundColor White "  Git Tag:         " -NoNewline
    Write-Host -ForegroundColor Yellow "$($targetVersionInfo.TagName) (Message: `"$($targetVersionInfo.TagMessage)`")"
    Write-Host ""

    Write-Host -ForegroundColor DarkCyan "-------------------- [CHANGELOG ENTRY PREVIEW] --------------------"
    Write-Host -ForegroundColor White "## [$($targetVersionInfo.SemVer)] - $todayDate"
    Write-Host ""
    Write-Host $releaseNotesBody
    Write-Host -ForegroundColor DarkCyan "------------------------------------------------------------------"
    Write-Host ""

    Write-Host -ForegroundColor DarkCyan "------------------------ [FILES TO UPDATE] -----------------------"
    foreach ($plan in $diffPlans) {
        Write-Host -ForegroundColor Green "  [~] $($plan.RelPath)"
        foreach ($d in $plan.Diffs) {
            Write-Host -ForegroundColor DarkGray "      Line $($d.LineNumber): $($d.Old) -> $($d.New)"
        }
    }
    Write-Host -ForegroundColor Green "  [~] frontend/package-lock.json (synced via 'npm i --package-lock-only')"
    Write-Host -ForegroundColor Green "  [~] CHANGELOG.md"
    Write-Host -ForegroundColor DarkCyan "------------------------------------------------------------------"
    Write-Host ""

    # 9. WhatIf Check
    if ($PSCmdlet.ShouldProcess($targetVersionInfo.TagName, "Release version and tag")) {
        # Continue to interactive prompt
    } else {
        Write-Host -ForegroundColor Yellow "WhatIf (-WhatIf) active: No changes were applied, committed, or tagged."
        Write-Host ""
        exit 0
    }

    # 10. Confirmation Prompt (Atomic Approval)
    if (-not $NonInteractive) {
        $confirm = Read-Host "Do you want to apply these updates, commit, and create tag $($targetVersionInfo.TagName)? (y/N)"
        if ($confirm -notmatch '^(y|yes)$') {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "Operation cancelled by user."
            Write-Host -ForegroundColor Yellow "ZERO files were modified, and no Git commits or tags were created."
            Write-Host ""
            exit 0
        }
    }

    # 11. Atomic Execution: Apply Updates to Disk
    Write-Host ""
    Write-Host -ForegroundColor Cyan "  Applying version updates to project files..."
    foreach ($plan in $diffPlans) {
        Set-Content -Path $plan.FullPath -Value $plan.NewContent -Encoding utf8 -NoNewline
        Write-Host -ForegroundColor Green "    Updated: $($plan.RelPath)"
    }

    # Sync package-lock.json
    Write-Host -ForegroundColor Cyan "  Syncing frontend/package-lock.json..."
    Push-Location (Join-Path $projectRoot "frontend")
    try {
        $npmOut = & npm i --package-lock-only 2>&1
        if ($LASTEXITCODE -ne 0) {
            throw "npm i --package-lock-only failed with exit code $($LASTEXITCODE):`n$npmOut"
        }
        Write-Host -ForegroundColor Green "    Updated: frontend/package-lock.json"
    } finally {
        Pop-Location
    }

    # Write CHANGELOG.md
    Write-Host -ForegroundColor Cyan "  Updating CHANGELOG.md..."
    [System.IO.File]::WriteAllText($ChangelogPath, $updatedChangelogContent, [System.Text.UTF8Encoding]::new($false))
    Write-Host -ForegroundColor Green "    Updated: CHANGELOG.md"

    # 12. Git Commit (Single Atomic Release Commit)
    Write-Host -ForegroundColor Cyan "  Staging and creating release commit..."
    $filesToStage = @(
        "wails.json",
        "frontend/package.json",
        "frontend/package-lock.json",
        "build/windows/info.json",
        "build/windows/installer/project.nsi",
        "CHANGELOG.md"
    )
    git add @filesToStage 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to stage release files in Git."
    }

    git commit -m $commitMessage 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create release commit in Git."
    }
    Write-Host -ForegroundColor Green "    Committed: $commitMessage"

    # 13. Create Annotated Git Tag
    Write-Host -ForegroundColor Cyan "  Creating annotated Git tag '$($targetVersionInfo.TagName)'..."
    git tag -a $targetVersionInfo.TagName -m $targetVersionInfo.TagMessage 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create Git tag '$($targetVersionInfo.TagName)'."
    }
    Write-Host -ForegroundColor Green "    Tagged: $($targetVersionInfo.TagName)"
    Write-Host ""
    Write-Host -ForegroundColor Green "SUCCESS: Version $($targetVersionInfo.SemVer) released and tagged locally!"
    Write-Host ""

    # 14. Remote Push Prompt
    $currentBranch = (git branch --show-current 2>$null)
    if ([string]::IsNullOrWhiteSpace($currentBranch)) {
        $currentBranch = "main"
    }

    $pushApproved = $false
    if (-not $NonInteractive) {
        $confirmPush = Read-Host "Push branch '$currentBranch' and tag '$($targetVersionInfo.TagName)' to remote origin? (y/N)"
        if ($confirmPush -match '^(y|yes)$') {
            $pushApproved = $true
        }
    }

    if ($pushApproved) {
        Write-Host -ForegroundColor Cyan "  Pushing '$currentBranch' and tag '$($targetVersionInfo.TagName)' to remote..."
        git push origin $currentBranch 2>&1 | Out-Host
        git push origin $targetVersionInfo.TagName 2>&1 | Out-Host
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host -ForegroundColor Green "SUCCESS: Release and tag successfully pushed to remote!"
            Write-Host ""
        } else {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "Notice: Git push encountered an issue. Please verify remote connectivity and push manually:"
            Write-Host -ForegroundColor Cyan   "  git push origin $currentBranch && git push origin $($targetVersionInfo.TagName)"
            Write-Host ""
        }
    } else {
        Write-Host ""
        Write-Host -ForegroundColor Yellow "Notice: Release committed and tagged locally, but not pushed to remote."
        Write-Host -ForegroundColor Cyan   "When ready, push using:"
        Write-Host -ForegroundColor White  "  git push origin $currentBranch && git push origin $($targetVersionInfo.TagName)"
        Write-Host ""
    }
}
catch {
    Write-Host ""
    Write-Host -ForegroundColor Red "ERROR: $($_.Exception.Message)"
    Write-Host ""
    exit 1
}
