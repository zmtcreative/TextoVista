#Requires -Version 7.2

<#
    .SYNOPSIS
        Shared helper functions for Git history queries, changelog parsing, version formatting,
        and AI-assisted release notes generation across TextoVista scripts.

    .DESCRIPTION
        Centralizes common routines used by Set-Version.ps1 and Update-Changelog.ps1:
        - Semantic version tag and commit history discovery.
        - Strict SemVer 2.0.0 version string parsing and pre-release identifier normalization.
        - Windows PE 4-part numeric FileVersion mapping.
        - Keep a Changelog 1.1.0 section and reference style extraction.
        - Antigravity CLI ('agy') prompt construction and execution.
#>

Set-StrictMode -Version Latest

# ==============================================================================
# Git & Tag Discovery Helpers
# ==============================================================================

function Get-LatestSemanticVersionTag {
    [CmdletBinding()]
    param()

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
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$FromRef,

        [Parameter(Mandatory = $false)]
        [string]$ToRef = "HEAD",

        [Parameter(Mandatory = $false)]
        [bool]$IncludeUncommitted = $false
    )

    $logFormat = "commit %h%nAuthor: %an%nDate: %ad%nSubject: %s%nBody: %b%n"
    $range = if (-not [string]::IsNullOrWhiteSpace($FromRef)) { "$FromRef..$ToRef" } else { $ToRef }

    $commitLog = git log $range --reverse --date=short --format=$logFormat 2>$null
    $shortLog = git log $range --reverse --date=short --pretty=format:"- %h %s (%an, %ad)" 2>$null
    $diffStat = git diff --stat $range 2>$null

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

# ==============================================================================
# Version Parsing & Formatting Helpers
# ==============================================================================

function Parse-VersionString {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Tag
    )

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
    # Allowed prerelease formats: alpha, beta, rc with optional delimiter and sequence number (e.g. -alpha.1, -alpha1, -beta2, -rc.1)
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
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        $vObj
    )

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

function Get-VersionAlertMarkdown {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$VersionString
    )

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

# ==============================================================================
# Changelog Extraction & Formatting Helpers
# ==============================================================================

function Get-ChangelogReferenceStyle {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$ChangelogPath
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        return ""
    }

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
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$ChangelogPath
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        return ""
    }

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

function Format-UnreleasedContent {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$SummaryMarkdown,

        [Parameter(Mandatory = $false)]
        [string]$AlertMarkdown = ""
    )

    $text = $SummaryMarkdown.Trim()
    # Strip any accidental alert callouts from the AI response
    $text = [regex]::Replace($text, '(?ms)^\s*>\s*\[!WARNING\].*?(?=\r?\n\r?\n|\r?\n###|\Z)', '').Trim()

    # Strip any local file:/// markdown links back to clean inline backticks (e.g. [`file.go`](file:///...) or [file.go](file:///...) -> `file.go`)
    $text = [regex]::Replace($text, '\[`?([^\]\r\n]+?)`?\]\(file:///[^\)]*\)', [System.Text.RegularExpressions.MatchEvaluator]{
        param($m)
        return ('`' + $m.Groups[1].Value.Trim('`') + '`')
    })

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

# ==============================================================================
# Antigravity CLI (AI) Helpers
# ==============================================================================

function Build-AgyChangelogPrompt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$SinceTag,

        [Parameter(Mandatory = $false)]
        [Alias("TargetVersion")]
        [string]$TargetHeading = "Unreleased",

        [Parameter(Mandatory = $false)]
        [string]$ToRef = "HEAD",

        [Parameter(Mandatory = $true)]
        [PSCustomObject]$GitData,

        [Parameter(Mandatory = $false)]
        [string]$ReferenceStyle = "",

        [Parameter(Mandatory = $false)]
        [string]$CurrentUnreleased = ""
    )

    $effectiveHeading = if (-not [string]::IsNullOrWhiteSpace($TargetHeading)) { $TargetHeading } else { "Unreleased" }
    $sectionTarget = if ($effectiveHeading.ToLower() -ne "unreleased") {
        "'## [$effectiveHeading]' release section"
    } else {
        "'## [Unreleased]' section"
    }

    $prompt = @"
You are an expert technical release notes writer.
Your task is to summarize the recent changes in this repository into content for the $sectionTarget of CHANGELOG.md adhering strictly to the Keep a Changelog (1.1.0) format.

### REPOSITORY CHANGES:
Baseline Tag:   $SinceTag
Target Ref:     $ToRef
Target Heading: $effectiveHeading
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
6. If current notes in '## [Unreleased]' describe actual changes, incorporate them into the appropriate category.
7. CRITICAL: Do NOT use markdown links, URL schemes, or 'file:///' links when referencing files, paths, or code symbols. Refer to files and symbols using ONLY plain relative paths enclosed in backticks (e.g., `path/to/file.ext`, `ClassName`).
8. CRITICAL: Do NOT output any '> [!WARNING]' alert callout or section headers (e.g., '## [$TargetHeading]'). Stability warning alerts, section headers, and compare link definitions are handled automatically by the script.
9. CRITICAL: Do NOT call any tools. Output ONLY pure markdown text without wrapping the entire response in backtick code fences (no ```markdown or ```).
"@

    return $prompt
}

function Invoke-AgyChangelogSummary {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Prompt,

        [Parameter(Mandatory = $false)]
        [string]$Model = ""
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
    if ($rawText -match '^```(?:markdown)?\s*([\s\S]*?)\s*```$') {
        $rawText = $Matches[1].Trim()
    }

    return $rawText
}
