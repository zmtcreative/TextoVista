<#
    .SYNOPSIS
        Sets and bumps project versions across target files, commits changes, and optionally manages Git tags.
    .DESCRIPTION
        Uses wails.json as the Source of Truth for the current project version.
        Calculates the next semantic version and Windows 4-part numeric version,
        prompts the user to confirm the bump, updates wails.json, package.json,
        info.json, and project.nsi, commits the changes with a Conventional Commit
        message ('chore(release): bump version to <version>'), and pushes the commit
        to the remote repository.

        Optionally creates an annotated Git tag locally using the updated version as
        the proposed tag (-CreateVersionTag), and only pushes the tag to the remote
        repository when explicitly requested (-PushTag).
    .PARAMETER NewVersion
        The explicit version or Git tag to set (e.g., '0.6.0' or 'v0.6.0').
        Aliases: -Version, -TagName, -TargetVersion.
    .PARAMETER Message
        The commit and tag message. Defaults to 'chore(release): bump version to X.Y.Z'.
    .PARAMETER Bump
        The version segment to increment: 'Patch', 'Minor', 'Major', or 'Prerelease'.
    .PARAMETER IncrementPatch
        Increment the patch version (e.g., 0.5.2 -> 0.5.3). Alias: -Patch.
    .PARAMETER IncrementMinor
        Increment the minor version (e.g., 0.5.2 -> 0.6.0). Alias: -Minor.
    .PARAMETER IncrementMajor
        Increment the major version (e.g., 0.5.2 -> 1.0.0). Alias: -Major.
    .PARAMETER IncrementPrerelease
        Increment the prerelease sequence number (e.g., beta1 -> beta2).
    .PARAMETER PrereleaseStage
        Set or switch the prerelease stage: 'alpha', 'beta', 'rc', or 'patch'.
    .PARAMETER SetAlpha
        Convenience switch to set the prerelease stage to alpha.
    .PARAMETER SetBeta
        Convenience switch to set the prerelease stage to beta.
    .PARAMETER SetReleaseCandidate
        Convenience switch to set the prerelease stage to rc. Alias: -SetRC.
    .PARAMETER ReleaseVersion
        Finalize a prerelease into a stable release (e.g., 0.5.2-beta1 -> 0.5.2). Alias: -Release.
    .PARAMETER CreateVersionTag
        Create an annotated Git tag locally using the new version from wails.json as the proposed tag.
        Prompts the user to verify or change the tag name before applying. Alias: -Tag, -CreateTag.
    .PARAMETER PushTag
        Push the Git tag to the remote repository.
        - If used with -CreateVersionTag: immediately pushes the confirmed tag to remote.
        - If used without -CreateVersionTag: inspects the most recent Git tag and prompts to push it.
    .PARAMETER ShowCurrentVersion
        Show current version in wails.json, most recent Git tag, and target files status without making changes.
        Alias: -Show, -Current.
    .PARAMETER DryRun
        Preview all file changes, commit details, and Git operations without modifying anything.
        Alias for -WhatIf.
    .PARAMETER NoPush
        Commit changes locally, but do not push commits to the remote repository.
    .PARAMETER NonInteractive
        Run without interactive confirmation prompts. Alias: -Force, -Yes.
    .EXAMPLE
        ./scripts/Set-Version.ps1 -IncrementPatch
        Bumps patch version in wails.json and project files, commits with 'chore(release): bump version to X.Y.Z',
        and pushes the commit to origin. Does not create a Git tag.
    .EXAMPLE
        ./scripts/Set-Version.ps1 -Minor -CreateVersionTag
        Bumps minor version, commits and pushes the commit, prompts to create local Git tag 'vX.Y.0',
        and creates the tag locally only (does not push tag).
    .EXAMPLE
        ./scripts/Set-Version.ps1 -IncrementPatch -CreateVersionTag -PushTag
        Bumps patch version, commits, pushes commit, prompts to verify tag, creates tag locally,
        and pushes the tag to origin.
    .EXAMPLE
        ./scripts/Set-Version.ps1 -PushTag
        Standalone tag push: finds most recent Git tag, prompts for verification, and pushes to origin.
#>

#Requires -Version 7.0

[CmdletBinding(SupportsShouldProcess = $true)]
param (
    [Parameter(Position = 0, HelpMessage = "The explicit target version or Git tag to set (e.g., '0.6.0' or 'v0.6.0').")]
    [Alias("Version", "TagName", "TargetVersion")]
    [string]$NewVersion = "",

    [Parameter(HelpMessage = "The commit message (and tag message if created). Defaults to 'chore(release): bump version to <version>'.")]
    [string]$Message = "",

    [Parameter(HelpMessage = "The version segment to increment.")]
    [ValidateSet('Patch', 'Minor', 'Major', 'Prerelease')]
    [string]$Bump = "",

    [Parameter(HelpMessage = "Increment the patch version.")]
    [Alias("Patch")]
    [switch]$IncrementPatch,

    [Parameter(HelpMessage = "Increment the minor version.")]
    [Alias("Minor")]
    [switch]$IncrementMinor,

    [Parameter(HelpMessage = "Increment the major version.")]
    [Alias("Major")]
    [switch]$IncrementMajor,

    [Parameter(HelpMessage = "Increment the prerelease sequence.")]
    [switch]$IncrementPrerelease,

    [Parameter(HelpMessage = "Set prerelease stage (alpha, beta, rc, patch).")]
    [ValidateSet('alpha', 'beta', 'rc', 'patch')]
    [Alias("Stage", "Prerelease")]
    [string]$PrereleaseStage = "",

    [Parameter(HelpMessage = "Set prerelease stage to alpha.")]
    [switch]$SetAlpha,

    [Parameter(HelpMessage = "Set prerelease stage to beta.")]
    [switch]$SetBeta,

    [Parameter(HelpMessage = "Set prerelease stage to release candidate (rc).")]
    [Alias("SetRC")]
    [switch]$SetReleaseCandidate,

    [Parameter(HelpMessage = "Finalize prerelease to stable release.")]
    [Alias("Release")]
    [switch]$ReleaseVersion,

    [Parameter(HelpMessage = "Create an annotated Git tag locally using the updated version as the proposed tag.")]
    [Alias("Tag", "CreateTag")]
    [switch]$CreateVersionTag,

    [Parameter(HelpMessage = "Push Git tag to remote repository.")]
    [Alias("Push")]
    [switch]$PushTag,

    [Parameter(HelpMessage = "Show current versions from wails.json, Git tags, and target files without making changes.")]
    [Alias("show", "current")]
    [switch]$ShowCurrentVersion,

    [Parameter(HelpMessage = "Preview planned changes without writing to disk or Git.")]
    [switch]$DryRun,

    [Parameter(HelpMessage = "Do not push commits or tags to the remote repository.")]
    [switch]$NoPush,

    [Parameter(HelpMessage = "Run non-interactively without confirmation prompts.")]
    [Alias("Force", "Yes", "f", "y")]
    [switch]$NonInteractive
)

# --- Path Resolution ---
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptDir = Split-Path -Parent $ScriptFullName
$ProjectRoot = if ($ScriptDir -match '[\\/]scripts[\\/]?$') {
    Split-Path -Parent $ScriptDir
} elseif (Test-Path (Join-Path (Get-Location) "go.mod")) {
    (Get-Location).Path
} else {
    $ScriptDir
}

if (-not (Test-Path -Path (Join-Path $ProjectRoot "go.mod"))) {
    Write-Host -ForegroundColor Red "Error: Could not locate 'go.mod' in project root: $ProjectRoot"
    exit 1
}

$WailsConfigPath = Join-Path $ProjectRoot "wails.json"
if (-not (Test-Path -Path $WailsConfigPath -PathType Leaf)) {
    Write-Host -ForegroundColor Red "Error: Source of truth 'wails.json' not found at: $WailsConfigPath"
    exit 1
}

# --- Declarative Target Files Registry ---
$TargetFiles = @(
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
            @{ Pattern = '(?m)^(?<before>VIFileVersion\s+)".*?"(?<after>\s*)$';    Replacement = '${before}"{VALUE}"${after}' },
            @{ Pattern = '(?m)^(?<before>VIAddVersionKey\s+"FileVersion"\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' }
        )
    }
)

# --- Version Parsing, Formatting, and Source of Truth Functions ---

function Get-CurrentVersionFromWails {
    param ([string]$Path)
    if (-not (Test-Path -Path $Path -PathType Leaf)) {
        throw "Source of truth 'wails.json' not found at: $Path"
    }
    $raw = Get-Content -Raw -Path $Path -Encoding utf8
    try {
        $json = $raw | ConvertFrom-Json
        if ($json.info -and -not [string]::IsNullOrWhiteSpace($json.info.productVersion)) {
            return $json.info.productVersion.Trim()
        }
    } catch {
        # Fallback to regex if ConvertFrom-Json fails
    }

    if ($raw -match '"productVersion"\s*:\s*"([^"]+)"') {
        return $Matches[1].Trim()
    }
    throw "Could not determine 'productVersion' from source of truth: $Path"
}

function Get-MostRecentTag {
    $tag = git describe --tags --abbrev=0 2>$null
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($tag)) {
        return $tag.Trim()
    }
    return $null
}

function Parse-VersionString {
    param ([string]$Tag)
    $clean = $Tag.Trim() -replace '^[vV]', ''
    $obj = [PSCustomObject]@{
        Raw        = $Tag
        Major      = 0
        Minor      = 0
        Patch      = 0
        Prerelease = $null
        Stage      = $null
        StageNum   = $null
        IsValid    = $false
    }

    if ($clean -match '^(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>[0-9A-Za-z.-]+))?$') {
        $obj.Major = [int]$Matches.major
        $obj.Minor = [int]$Matches.minor
        $obj.Patch = [int]$Matches.patch
        $obj.Prerelease = $Matches.prerelease
        $obj.IsValid = $true

        if ($Matches.prerelease -and $Matches.prerelease -match '^(?<stage>alpha|beta|rc|patch)[.-]?(?<num>\d+)?$') {
            $obj.Stage = $Matches.stage.ToLower()
            $obj.StageNum = if ([string]::IsNullOrWhiteSpace($Matches.num)) { 1 } else { [int]$Matches.num }
        }
    }
    return $obj
}

function Format-BuildVersions {
    param ($vObj)
    $semver = "$($vObj.Major).$($vObj.Minor).$($vObj.Patch)"
    if ($vObj.Prerelease) {
        $semver += "-$($vObj.Prerelease)"
    }
    $tagName = "v$semver"

    # Windows 4-part numeric version: Major.Minor.Patch.Suffix
    # Final releases use fixed 9000 so they always compare as newer than any prerelease.
    # Prereleases: alpha -> 1000 + n, beta -> 2000 + n, rc -> 3000 + n, patch -> 4000 + n
    $stages = @('alpha', 'beta', 'rc', 'patch')
    $fileVersionSuffix = 9000
    if ($vObj.Prerelease) {
        if ($vObj.Stage) {
            $stageIndex = $stages.IndexOf($vObj.Stage) + 1
            if ($stageIndex -gt 0) {
                $num = if ($vObj.StageNum) { [int]$vObj.StageNum } else { 1 }
                $fileVersionSuffix = ($stageIndex * 1000) + $num
            } else {
                $fileVersionSuffix = 1001
            }
        } else {
            $fileVersionSuffix = 1001
        }
    }

    $fileVersion = "$($vObj.Major).$($vObj.Minor).$($vObj.Patch).$fileVersionSuffix"

    return [PSCustomObject]@{
        SemVer      = $semver
        TagName     = $tagName
        FileVersion = $fileVersion
    }
}

function Calculate-NextVersion {
    param (
        $CurrentObj,
        [string]$ExplicitTag = "",
        [string]$RequestedBump = "",
        [string]$RequestedStage = "",
        [switch]$IsRelease
    )

    if (-not [string]::IsNullOrWhiteSpace($ExplicitTag)) {
        $parsed = Parse-VersionString -Tag $ExplicitTag
        if (-not $parsed.IsValid) {
            throw "Invalid explicit version/tag provided: '$ExplicitTag'. Must follow SemVer (e.g., '0.6.0' or 'v0.6.0-beta1')."
        }
        return $parsed
    }

    $major = $CurrentObj.Major
    $minor = $CurrentObj.Minor
    $patch = $CurrentObj.Patch
    $prerelease = $null
    $stageName = $null
    $stageNum = $null

    # If -Release is requested, strip prerelease from current version
    if ($IsRelease) {
        return [PSCustomObject]@{
            Major      = $major
            Minor      = $minor
            Patch      = $patch
            Prerelease = $null
            Stage      = $null
            StageNum   = $null
            IsValid    = $true
        }
    }

    $effectiveBump = $RequestedBump
    if ([string]::IsNullOrWhiteSpace($effectiveBump)) {
        if ($RequestedStage) {
            # If a prerelease stage was specified without an explicit bump:
            # If current version is already stable, bump patch before applying prerelease stage
            if (-not $CurrentObj.Prerelease) {
                $effectiveBump = "Patch"
            }
        } elseif ($CurrentObj.Prerelease) {
            $effectiveBump = "Prerelease"
        } else {
            $effectiveBump = "Patch"
        }
    }

    switch ($effectiveBump) {
        "Major" {
            $major++
            $minor = 0
            $patch = 0
        }
        "Minor" {
            $minor++
            $patch = 0
        }
        "Patch" {
            $patch++
        }
        "Prerelease" {
            if ($CurrentObj.Prerelease -and $CurrentObj.Stage) {
                $stageName = if ($RequestedStage) { $RequestedStage.ToLower() } else { $CurrentObj.Stage }
                $stageNum = if ($RequestedStage -and ($RequestedStage.ToLower() -ne $CurrentObj.Stage)) {
                    1
                } else {
                    $CurrentObj.StageNum + 1
                }
            } else {
                $patch++
                $stageName = if ($RequestedStage) { $RequestedStage.ToLower() } else { "alpha" }
                $stageNum = 1
            }
            $prerelease = "${stageName}${stageNum}"
        }
    }

    # If a specific stage was passed with Major/Minor/Patch
    if ($RequestedStage -and ($effectiveBump -ne "Prerelease")) {
        $stageName = $RequestedStage.ToLower()
        $stageNum = 1
        $prerelease = "${stageName}${stageNum}"
    }

    return [PSCustomObject]@{
        Major      = $major
        Minor      = $minor
        Patch      = $patch
        Prerelease = $prerelease
        Stage      = $stageName
        StageNum   = $stageNum
        IsValid    = $true
    }
}

# --- File Update Engine ---

function Test-TargetFilesExist {
    param ([string]$BaseDir)
    $allFound = $true
    foreach ($target in $TargetFiles) {
        $fullPath = Join-Path $BaseDir $target.RelPath
        if (-not (Test-Path -Path $fullPath -PathType Leaf)) {
            Write-Host -ForegroundColor Red "Missing target file: $fullPath"
            $allFound = $false
        }
    }
    return $allFound
}

function Get-FileDiffPlan {
    param (
        [string]$BaseDir,
        [string]$SemVer,
        [string]$FileVersion
    )
    $diffPlans = @()

    foreach ($target in $TargetFiles) {
        $fullPath = Join-Path $BaseDir $target.RelPath
        $val = if ($target.Format -eq "SemVer") { $SemVer } else { $FileVersion }
        $content = Get-Content -Raw -Path $fullPath -Encoding utf8
        $newContent = $content
        $lineDiffs = @()

        foreach ($rule in $target.Rules) {
            $repPattern = $rule.Replacement.Replace('{VALUE}', $val)
            $newContent = [regex]::Replace($newContent, $rule.Pattern, $repPattern)
        }

        if ($newContent -ne $content) {
            $origLines = $content -split "`r?`n"
            $updLines = $newContent -split "`r?`n"
            for ($i = 0; $i -lt $origLines.Count; $i++) {
                if ($origLines[$i] -ne $updLines[$i]) {
                    $lineDiffs += [PSCustomObject]@{
                        LineNumber = $i + 1
                        Old = $origLines[$i]
                        New = $updLines[$i]
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

function Apply-FileUpdates {
    param ($DiffPlans)
    foreach ($plan in $DiffPlans) {
        if ($plan.IsModified) {
            Set-Content -Path $plan.FullPath -Value $plan.NewContent -Encoding utf8 -NoNewline
            Write-Host -ForegroundColor Green "  Updated: $($plan.RelPath) -> $($plan.TargetVal)"
        } else {
            Write-Host -ForegroundColor DarkGray "  No changes needed: $($plan.RelPath)"
        }
    }
}

# --- Main Controller Logic ---

Push-Location $ProjectRoot
try {
    $isDryRun = $DryRun -or $PSCmdlet.MyInvocation.BoundParameters.ContainsKey('WhatIf')

    $stage = $PrereleaseStage
    if ($SetAlpha) { $stage = "alpha" }
    elseif ($SetBeta) { $stage = "beta" }
    elseif ($SetReleaseCandidate) { $stage = "rc" }

    $bumpAction = $Bump
    if ($IncrementMajor) { $bumpAction = "Major" }
    elseif ($IncrementMinor) { $bumpAction = "Minor" }
    elseif ($IncrementPatch) { $bumpAction = "Patch" }
    elseif ($IncrementPrerelease) { $bumpAction = "Prerelease" }

    # Detect if any version bump flag was provided
    $hasExplicitBumpInput = (
        (-not [string]::IsNullOrWhiteSpace($NewVersion)) -or
        (-not [string]::IsNullOrWhiteSpace($Bump)) -or
        $IncrementMajor -or $IncrementMinor -or $IncrementPatch -or $IncrementPrerelease -or
        (-not [string]::IsNullOrWhiteSpace($PrereleaseStage)) -or
        $SetAlpha -or $SetBeta -or $SetReleaseCandidate -or $ReleaseVersion -or
        $CreateVersionTag
    )

    # 1. Read Current Version from wails.json (Source of Truth)
    $sourceVersionStr = Get-CurrentVersionFromWails -Path $WailsConfigPath
    $currentObj = Parse-VersionString -Tag $sourceVersionStr
    if (-not $currentObj.IsValid) {
        Write-Host -ForegroundColor Red "Error: Failed to parse version '$sourceVersionStr' from $WailsConfigPath as valid SemVer."
        exit 1
    }
    $currentFormatted = Format-BuildVersions $currentObj
    $mostRecentGitTag = Get-MostRecentTag

    # 2. Handle -ShowCurrentVersion
    if ($ShowCurrentVersion) {
        Write-Host ""
        Write-Host -ForegroundColor Cyan "TextoVista Version Information:"
        Write-Host -NoNewline "  Source of Truth (wails.json) : "; Write-Host -ForegroundColor Green "$($currentFormatted.SemVer)"
        Write-Host -NoNewline "  Windows File Version         : "; Write-Host -ForegroundColor Green "$($currentFormatted.FileVersion)"
        if ($mostRecentGitTag) {
            Write-Host -NoNewline "  Most Recent Git Tag          : "; Write-Host -ForegroundColor Cyan "$mostRecentGitTag"
        } else {
            Write-Host -NoNewline "  Most Recent Git Tag          : "; Write-Host -ForegroundColor Yellow "No tags found"
        }
        Write-Host ""
        Write-Host -ForegroundColor Cyan "Target Files Status:"
        foreach ($target in $TargetFiles) {
            $p = Join-Path $ProjectRoot $target.RelPath
            $exists = Test-Path -Path $p
            $status = if ($exists) { "Found" } else { "Missing" }
            $color = if ($exists) { "Green" } else { "Red" }
            Write-Host -NoNewline "  $($target.RelPath.PadRight(40)) ["
            Write-Host -NoNewline -ForegroundColor $color $status
            Write-Host "]"
        }
        Write-Host ""
        return
    }

    # 3. Handle Standalone -PushTag Mode (when -PushTag is passed alone without -CreateVersionTag or bump flags)
    if ($PushTag -and (-not $CreateVersionTag) -and (-not $hasExplicitBumpInput)) {
        Write-Host ""
        Write-Host -ForegroundColor Cyan "=========================================================="
        Write-Host -ForegroundColor Cyan "             TextoVista Remote Tag Push                   "
        Write-Host -ForegroundColor Cyan "=========================================================="
        if (-not $mostRecentGitTag) {
            Write-Host -ForegroundColor Red "Error: No Git tags found in the repository to push."
            exit 1
        }

        Write-Host -NoNewline "  Most Recent Git Tag : "; Write-Host -ForegroundColor Green "$mostRecentGitTag"
        Write-Host -NoNewline "  Action              : "; Write-Host -ForegroundColor Yellow "Push existing tag '$mostRecentGitTag' to remote 'origin'"
        Write-Host ""

        if ($isDryRun) {
            Write-Host -ForegroundColor Yellow "[DryRun / WhatIf Mode] Simulation complete."
            Write-Host -ForegroundColor Cyan "Planned Git command:"
            Write-Host "  git push origin $mostRecentGitTag"
            Write-Host ""
            return
        }

        $confirmedPush = $false
        if ($NonInteractive) {
            $confirmedPush = $true
        } else {
            Write-Host -NoNewline -ForegroundColor Cyan "Push most recent tag '$mostRecentGitTag' to remote 'origin'? [Y]es / [N]o (default: Y): "
            $resp = Read-Host
            if ([string]::IsNullOrWhiteSpace($resp) -or $resp -imatch '^\s*y(es)?\s*$') {
                $confirmedPush = $true
            }
        }

        if ($confirmedPush) {
            Write-Host -ForegroundColor Cyan "Pushing tag '$mostRecentGitTag' to remote 'origin'..."
            git push origin "$mostRecentGitTag"
            if ($LASTEXITCODE -ne 0) {
                Write-Host -ForegroundColor Red "Error: Failed to push tag '$mostRecentGitTag' to remote."
                exit 1
            }
            Write-Host -ForegroundColor Green "Tag '$mostRecentGitTag' successfully pushed to remote!"
        } else {
            Write-Host -ForegroundColor Yellow "Tag push cancelled by user."
        }
        return
    }

    # 4. Verify working tree is clean (allowing target files and Set-Version.ps1 / New-GitTag.ps1)
    $uncommitted = git status --porcelain=v1
    if (-not [string]::IsNullOrWhiteSpace($uncommitted)) {
        $dirtyLines = $uncommitted -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
        $otherChanges = @()
        foreach ($line in $dirtyLines) {
            $file = $line.Substring(3).Trim()
            $isTarget = $false
            if ($file -match 'Set-Version\.ps1$' -or $file -match 'New-GitTag\.ps1$') {
                $isTarget = $true
            }
            foreach ($tf in $TargetFiles) {
                if ($file -eq $tf.RelPath -or $file -eq ($tf.RelPath -replace '/', '\')) {
                    $isTarget = $true
                    break
                }
            }
            if (-not $isTarget) {
                $otherChanges += $line
            }
        }

        if ($otherChanges.Count -gt 0) {
            Write-Host ""
            Write-Host -ForegroundColor Red "Error: Working directory has uncommitted changes that are not part of the release files:"
            Write-Host ""
            $otherChanges | ForEach-Object { Write-Host -ForegroundColor Yellow "  $_" }
            Write-Host ""
            Write-Host "Please commit, stash, or revert these changes before setting version."
            exit 1
        }
    }

    # 5. Verify all target files exist
    if (-not (Test-TargetFilesExist -BaseDir $ProjectRoot)) {
        Write-Host -ForegroundColor Red "Aborting: One or more target files are missing."
        exit 1
    }

    # 6. Calculate target version using wails.json version as baseline
    $nextObj = Calculate-NextVersion `
        -CurrentObj $currentObj `
        -ExplicitTag $NewVersion `
        -RequestedBump $bumpAction `
        -RequestedStage $stage `
        -IsRelease:($ReleaseVersion -eq $true)

    $nextFormatted = Format-BuildVersions $nextObj
    $newSemVer = $nextFormatted.SemVer
    $newFileVer = $nextFormatted.FileVersion
    $proposedTagName = $nextFormatted.TagName

    # 7. Prepare commit message (Default Conventional Commit: 'chore(release): bump version to <version>')
    if ([string]::IsNullOrWhiteSpace($Message)) {
        $finalCommitMessage = "chore(release): bump version to $newSemVer"
    } else {
        $finalCommitMessage = $Message
    }

    # 8. Check current branch and push capability
    $currentBranch = (git branch --show-current 2>$null).Trim()
    if ([string]::IsNullOrWhiteSpace($currentBranch)) {
        $currentBranch = "HEAD (detached)"
    }
    $canPushBranch = (-not $NoPush) -and ($currentBranch -ne "HEAD (detached)")

    # 9. Compute file diff plan
    $diffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $newSemVer -FileVersion $newFileVer

    # 10. Display Planned Operations Summary
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host -ForegroundColor Cyan "             TextoVista Version Bump Plan                 "
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host "  Source of Truth (wails.json) : $($currentFormatted.SemVer) (Win: $($currentFormatted.FileVersion))"
    if ($mostRecentGitTag) {
        Write-Host "  Most Recent Git Tag          : $mostRecentGitTag"
    }
    Write-Host -NoNewline "  Target Version               : "; Write-Host -ForegroundColor Green "$newSemVer (Win: $newFileVer)"
    Write-Host -NoNewline "  Commit Message               : "; Write-Host -ForegroundColor Yellow "$finalCommitMessage"
    Write-Host -NoNewline "  Git Branch                   : "; Write-Host "$currentBranch"
    Write-Host -NoNewline "  Push Commit to Remote        : "
    if ($NoPush) {
        Write-Host -ForegroundColor Yellow "No (-NoPush specified: Local commit only)"
    } elseif ($currentBranch -eq "HEAD (detached)") {
        Write-Host -ForegroundColor Yellow "No (Detached HEAD: Local commit only)"
    } else {
        Write-Host -ForegroundColor Green "Yes (origin/$currentBranch)"
    }

    Write-Host -NoNewline "  Create Git Tag Locally       : "
    if ($CreateVersionTag) {
        Write-Host -ForegroundColor Green "Yes (-CreateVersionTag specified: Proposed tag '$proposedTagName')"
    } else {
        Write-Host -ForegroundColor DarkGray "No (Default: tag will not be created)"
    }

    Write-Host -NoNewline "  Push Git Tag to Remote       : "
    if ($PushTag) {
        Write-Host -ForegroundColor Green "Yes (-PushTag specified)"
    } else {
        Write-Host -ForegroundColor DarkGray "No (Local tag only if created)"
    }

    Write-Host ""
    Write-Host -ForegroundColor Cyan "Target File Diffs:"
    foreach ($plan in $diffPlans) {
        if ($plan.IsModified) {
            Write-Host "  File: $($plan.RelPath)" -ForegroundColor Magenta
            foreach ($d in $plan.Diffs) {
                Write-Host "    - $($d.Old.Trim())" -ForegroundColor Red
                Write-Host "    + $($d.New.Trim())" -ForegroundColor Green
            }
        } else {
            Write-Host "  File: $($plan.RelPath) (already at target version)" -ForegroundColor DarkGray
        }
    }
    Write-Host ""

    # 11. Handle -WhatIf / -DryRun
    if ($isDryRun) {
        Write-Host -ForegroundColor Yellow "[DryRun / WhatIf Mode] Simulation complete. No files or Git references modified."
        Write-Host ""
        Write-Host -ForegroundColor Cyan "Planned Git commands that would execute:"
        $relList = ($diffPlans | Where-Object { $_.IsModified } | ForEach-Object { $_.RelPath }) -join " "
        if (-not [string]::IsNullOrWhiteSpace($relList)) {
            Write-Host "  git add $relList"
            Write-Host "  git commit -m `"$finalCommitMessage`""
            if ($canPushBranch) {
                Write-Host "  git push origin $currentBranch"
            }
        }
        if ($CreateVersionTag) {
            Write-Host "  git tag -a `"$proposedTagName`" -m `"$finalCommitMessage`""
            if ($PushTag) {
                Write-Host "  git push origin $proposedTagName"
            }
        } elseif ($PushTag -and $mostRecentGitTag) {
            Write-Host "  git push origin $mostRecentGitTag (after user prompt)"
        }
        Write-Host ""
        return
    }

    # 12. Interactive Prompt: Confirm Version Bump
    if (-not $NonInteractive) {
        Write-Host -NoNewline -ForegroundColor Cyan "Apply version bump to '$newSemVer', stage files, and commit? [Y]es / [N]o / [E]dit version or message (default: Y): "
        $response = Read-Host
        if ([string]::IsNullOrWhiteSpace($response) -or $response -imatch '^\s*y(es)?\s*$') {
            # User confirmed
        } elseif ($response -imatch '^\s*e(dit)?\s*$') {
            $customVer = Read-Host "  Enter version to set (press ENTER to keep '$newSemVer')"
            if (-not [string]::IsNullOrWhiteSpace($customVer)) {
                $newSemVer = $customVer.Trim() -replace '^[vV]', ''
                $parsedCustom = Parse-VersionString -Tag $newSemVer
                if ($parsedCustom.IsValid) {
                    $nextFormatted = Format-BuildVersions $parsedCustom
                    $newSemVer = $nextFormatted.SemVer
                    $newFileVer = $nextFormatted.FileVersion
                    $proposedTagName = $nextFormatted.TagName
                    $diffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $newSemVer -FileVersion $newFileVer
                } else {
                    Write-Host -ForegroundColor Red "Invalid version format '$customVer'. Aborting."
                    return
                }
            }
            $customMsg = Read-Host "  Enter commit message (press ENTER to keep '$finalCommitMessage')"
            if (-not [string]::IsNullOrWhiteSpace($customMsg)) {
                $finalCommitMessage = $customMsg
            }
            Write-Host ""
            Write-Host -ForegroundColor Cyan "Updated Target Version: $newSemVer | Commit Message: $finalCommitMessage"
            $confirm2 = Read-Host "Proceed with these updated values? [Y/n]"
            if ($confirm2 -inotmatch '^\s*(y(es)?|$)\s*$') {
                Write-Host -ForegroundColor Yellow "Operation cancelled by user. No changes made."
                return
            }
        } else {
            Write-Host -ForegroundColor Yellow "Operation cancelled by user. No changes made."
            return
        }
    }

    # 13. Execute File Updates
    Write-Host -ForegroundColor Cyan "Applying file updates..."
    Apply-FileUpdates -DiffPlans $diffPlans

    # 14. Git Stage & Commit Version Bump
    Write-Host -ForegroundColor Cyan "Staging files in Git..."
    $modifiedFiles = $diffPlans | Where-Object { $_.IsModified } | ForEach-Object { $_.RelPath }
    if ($modifiedFiles.Count -gt 0) {
        git add $modifiedFiles 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to stage files in Git."
            exit 1
        }

        Write-Host -ForegroundColor Cyan "Committing changes: $finalCommitMessage"
        git commit -m "$finalCommitMessage" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to commit changes to Git."
            exit 1
        }
        Write-Host -ForegroundColor Green "  Commit created successfully."

        # Push Version Bump Commit to Remote
        if ($canPushBranch) {
            Write-Host -ForegroundColor Cyan "Pushing version bump commit to remote 'origin/$currentBranch'..."
            git push origin "$currentBranch" 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host -ForegroundColor Yellow "Warning: Failed to push branch '$currentBranch' to remote. You may need to push manually."
            } else {
                Write-Host -ForegroundColor Green "  Version bump commit pushed successfully to remote."
            }
        } else {
            Write-Host -ForegroundColor Yellow "  Skipped remote commit push (NoPush specified or detached HEAD)."
        }
    } else {
        Write-Host -ForegroundColor Yellow "  No file changes to commit (files already at target version)."
    }

    # 15. Tag Handling Logic
    $tagToApply = $proposedTagName

    if ($CreateVersionTag) {
        # Prompt user to verify or change the proposed tag
        $confirmedTag = $false
        if ($NonInteractive) {
            $confirmedTag = $true
        } else {
            Write-Host ""
            Write-Host -NoNewline -ForegroundColor Cyan "Create local Git tag using proposed tag '$proposedTagName'? [Y]es / [N]o / [E]dit tag name (default: Y): "
            $tagResp = Read-Host
            if ([string]::IsNullOrWhiteSpace($tagResp) -or $tagResp -imatch '^\s*y(es)?\s*$') {
                $confirmedTag = $true
            } elseif ($tagResp -imatch '^\s*e(dit)?\s*$') {
                $customTagInput = Read-Host "  Enter tag name (press ENTER to keep '$proposedTagName')"
                if (-not [string]::IsNullOrWhiteSpace($customTagInput)) {
                    $tagToApply = $customTagInput.Trim()
                }
                $confirmedTag = $true
            } else {
                Write-Host -ForegroundColor Yellow "Skipped Git tag creation."
            }
        }

        if ($confirmedTag) {
            Write-Host -ForegroundColor Cyan "Creating local Git tag: $tagToApply"
            git tag -a "$tagToApply" -m "$finalCommitMessage" 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host -ForegroundColor Red "Error: Failed to create Git tag '$tagToApply'."
                exit 1
            }
            Write-Host -ForegroundColor Green "  Tag '$tagToApply' created locally."

            # Check if tag should be pushed to remote
            if ($PushTag) {
                # Pushed immediately since tag was confirmed
                Write-Host -ForegroundColor Cyan "Pushing tag '$tagToApply' to remote 'origin'..."
                git push origin "$tagToApply" 2>&1 | Out-Null
                if ($LASTEXITCODE -ne 0) {
                    Write-Host -ForegroundColor Yellow "Warning: Failed to push tag '$tagToApply' to remote. You may need to push manually."
                } else {
                    Write-Host -ForegroundColor Green "  Tag '$tagToApply' pushed successfully to remote."
                }
            } else {
                Write-Host -ForegroundColor Yellow "  Tag '$tagToApply' exists locally only. (To push to remote, run: git push origin $tagToApply or rerun with -PushTag)."
            }
        }
    } elseif ($PushTag) {
        # -PushTag specified WITHOUT -CreateVersionTag: Check current most recent tag and prompt
        $latestTag = Get-MostRecentTag
        Write-Host ""
        if (-not $latestTag) {
            Write-Host -ForegroundColor Yellow "Warning: -PushTag was requested without -CreateVersionTag, but no Git tags were found in repository."
        } else {
            $confirmLatestPush = $false
            if ($NonInteractive) {
                $confirmLatestPush = $true
            } else {
                Write-Host -NoNewline -ForegroundColor Cyan "Push most recent Git tag '$latestTag' to remote 'origin'? [Y]es / [N]o (default: Y): "
                $pResp = Read-Host
                if ([string]::IsNullOrWhiteSpace($pResp) -or $pResp -imatch '^\s*y(es)?\s*$') {
                    $confirmLatestPush = $true
                }
            }

            if ($confirmLatestPush) {
                Write-Host -ForegroundColor Cyan "Pushing tag '$latestTag' to remote 'origin'..."
                git push origin "$latestTag" 2>&1 | Out-Null
                if ($LASTEXITCODE -ne 0) {
                    Write-Host -ForegroundColor Yellow "Warning: Failed to push tag '$latestTag' to remote. You may need to push manually."
                } else {
                    Write-Host -ForegroundColor Green "  Tag '$latestTag' pushed successfully to remote."
                }
            } else {
                Write-Host -ForegroundColor Yellow "Skipped remote tag push."
            }
        }
    }

    Write-Host ""
    Write-Host -ForegroundColor Green "Set-Version completed successfully!"
    Write-Host ""
}
finally {
    Pop-Location
}
