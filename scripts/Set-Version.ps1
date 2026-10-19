<#
    .SYNOPSIS
        Sets and bumps project versions across target files, commits changes, and optionally manages Git tags.
    .DESCRIPTION
        Uses wails.json as the Source of Truth for the current project version.
        Calculates the next semantic version and Windows 4-part numeric version,
        prompts the user to confirm the bump, updates wails.json, package.json,
        info.json, and project.nsi, syncs frontend/package-lock.json via npm,
        commits the changes with a Conventional Commit message
        ('chore(release): bump version to <version>'), and pushes the commit
        to the remote repository.

        Optionally creates an annotated Git tag locally using the updated version as
        the proposed tag (-CreateVersionTag), and only pushes the tag to the remote
        repository when explicitly requested (-PushTag).
    .PARAMETER NewVersion
        The explicit version or Git tag to set (e.g., '0.6.0' or 'v0.6.0').
        Aliases: -Version, -TagName, -TargetVersion.
    .PARAMETER Message
        The commit message for setting the version. Defaults to 'chore(release): bump version to <version>'.
        Alias: -CommitMessage.
    .PARAMETER TagMessage
        The annotation message for creating the Git tag. Defaults to 'Version: <version>' for SemVer tags (e.g. 'Version: 1.0.0') or 'Version: <tag>' for non-SemVer tags.
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
        Set or switch the prerelease stage: 'alpha', 'beta', or 'rc'.
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
        Prompts the user to verify or change the tag name before applying. Allows editing to custom or
        non-SemVer tags (such as 'test-vX.Y.Z') with a warning, skipping CHANGELOG.md verification for
        non-SemVer tags. Alias: -Tag, -CreateTag.
    .PARAMETER PushTag
        Push the Git tag to the remote repository.
        - If used with -CreateVersionTag: immediately pushes the confirmed tag to remote.
        - If used without -CreateVersionTag: inspects the most recent Git tag and prompts to push it.
    .PARAMETER CurrentVersion
        Skip version bumping operations and create the Git tag using the existing version from wails.json.
        Only valid in combination with -CreateVersionTag.
    .PARAMETER ShowCurrentVersion
        Show current version in wails.json, most recent Git tag, and target files status without making changes.
        Alias: -Show.
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
        ./scripts/Set-Version.ps1 -CreateVersionTag -CurrentVersion
        Creates a Git tag locally for the current version in wails.json without bumping or modifying project files.
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

    [Parameter(HelpMessage = "The commit message for setting the version. Defaults to 'chore(release): bump version to <version>'.")]
    [Alias("CommitMessage")]
    [string]$Message = "",

    [Parameter(HelpMessage = "The tag message for setting the tag. Defaults to 'Version: <version>' for SemVer tags (or 'Version: <tag>' for non-SemVer tags).")]
    [string]$TagMessage = "",

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

    [Parameter(HelpMessage = "Set prerelease stage (alpha, beta, or rc).")]
    [ValidateSet('alpha', 'beta', 'rc')]
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

    [Parameter(HelpMessage = "Skip version bumping operations and create the Git tag using the current version from wails.json. Only valid with -CreateVersionTag.")]
    [switch]$CurrentVersion,

    [Parameter(HelpMessage = "Push Git tag to remote repository.")]
    [Alias("Push")]
    [switch]$PushTag,

    [Parameter(HelpMessage = "Show current versions from wails.json, Git tags, and target files without making changes.")]
    [Alias("show")]
    [switch]$ShowCurrentVersion,

    [Parameter(HelpMessage = "Preview planned changes without writing to disk or Git.")]
    [switch]$DryRun,

    [Parameter(HelpMessage = "Do not push commits or tags to the remote repository.")]
    [switch]$NoPush,

    [Parameter(HelpMessage = "Run non-interactively without confirmation prompts.")]
    [Alias("Force", "Yes", "f", "y")]
    [switch]$NonInteractive
)

# --- Path Resolution & Project Setup ---
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptDir = Split-Path -Parent $ScriptFullName
$ProjectRoot = if ($ScriptDir -match '[\\/]scripts[\\/]?$') {
    Split-Path -Parent $ScriptDir
}
elseif (Test-Path (Join-Path (Get-Location) "go.mod")) {
    (Get-Location).Path
}
else {
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
            @{ Pattern = '(?m)^(?<before>VIFileVersion\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' },
            @{ Pattern = '(?m)^(?<before>VIAddVersionKey\s+"FileVersion"\s+)".*?"(?<after>\s*)$'; Replacement = '${before}"{VALUE}"${after}' }
        )
    }
)

# ==============================================================================
# Helper Functions: Parameter & Flag Resolvers
# ==============================================================================

function Resolve-PrereleaseStage {
    param (
        [string]$Stage,
        [switch]$Alpha,
        [switch]$Beta,
        [switch]$RC
    )
    if ($Alpha) { return "alpha" }
    if ($Beta) { return "beta" }
    if ($RC) { return "rc" }
    return $Stage
}

function Resolve-BumpAction {
    param (
        [string]$Bump,
        [switch]$Major,
        [switch]$Minor,
        [switch]$Patch,
        [switch]$Prerelease
    )
    if ($Major) { return "Major" }
    if ($Minor) { return "Minor" }
    if ($Patch) { return "Patch" }
    if ($Prerelease) { return "Prerelease" }
    return $Bump
}

function Assert-ParameterRules {
    param (
        [switch]$CurrentVersion,
        [switch]$CreateVersionTag,
        [string]$NewVersion,
        [string]$Bump,
        [switch]$IncrementMajor,
        [switch]$IncrementMinor,
        [switch]$IncrementPatch,
        [switch]$IncrementPrerelease,
        [string]$PrereleaseStage,
        [switch]$SetAlpha,
        [switch]$SetBeta,
        [switch]$SetReleaseCandidate,
        [switch]$ReleaseVersion
    )

    if ($CurrentVersion) {
        if (-not $CreateVersionTag) {
            Write-Host -ForegroundColor Red "Error: The -CurrentVersion switch can only be used in combination with -CreateVersionTag."
            exit 1
        }
        $bumpFlagsSpecified = (
            (-not [string]::IsNullOrWhiteSpace($NewVersion)) -or
            (-not [string]::IsNullOrWhiteSpace($Bump)) -or
            $IncrementMajor -or $IncrementMinor -or $IncrementPatch -or $IncrementPrerelease -or
            (-not [string]::IsNullOrWhiteSpace($PrereleaseStage)) -or
            $SetAlpha -or $SetBeta -or $SetReleaseCandidate -or $ReleaseVersion
        )
        if ($bumpFlagsSpecified) {
            Write-Host -ForegroundColor Red "Error: The -CurrentVersion switch prevents version bumping and cannot be combined with version increment parameters."
            exit 1
        }
    }
}

function Test-ExplicitBumpInputSpecified {
    param (
        [string]$NewVersion,
        [string]$Bump,
        [switch]$IncrementMajor,
        [switch]$IncrementMinor,
        [switch]$IncrementPatch,
        [switch]$IncrementPrerelease,
        [string]$PrereleaseStage,
        [switch]$SetAlpha,
        [switch]$SetBeta,
        [switch]$SetReleaseCandidate,
        [switch]$ReleaseVersion,
        [switch]$CreateVersionTag
    )

    return (
        (-not [string]::IsNullOrWhiteSpace($NewVersion)) -or
        (-not [string]::IsNullOrWhiteSpace($Bump)) -or
        $IncrementMajor -or $IncrementMinor -or $IncrementPatch -or $IncrementPrerelease -or
        (-not [string]::IsNullOrWhiteSpace($PrereleaseStage)) -or
        $SetAlpha -or $SetBeta -or $SetReleaseCandidate -or $ReleaseVersion -or
        $CreateVersionTag
    )
}

# ==============================================================================
# Helper Functions: Git Queries, Cleanliness & Tag Resolution
# ==============================================================================

function Get-MostRecentTag {
    $tag = git describe --tags --abbrev=0 2>$null
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($tag)) {
        return $tag.Trim()
    }
    return $null
}

function Test-GitTagExists {
    param ([string]$Tag)
    if ([string]::IsNullOrWhiteSpace($Tag)) { return $false }
    $found = git tag -l $Tag.Trim() 2>$null
    return ($found -and ($found -contains $Tag.Trim() -or $found -eq $Tag.Trim()))
}

function Test-ValidGitTagName {
    param ([string]$Tag)
    if ([string]::IsNullOrWhiteSpace($Tag)) { return $false }
    $trimmed = $Tag.Trim()
    if ($trimmed -match '^-') { return $false }
    git check-ref-format --allow-onelevel $trimmed 2>&1 | Out-Null
    return ($LASTEXITCODE -eq 0)
}

function Resolve-TagCollision {
    param (
        [string]$TagName,
        [switch]$NonInteractive
    )

    $currentTag = $TagName
    while (Test-GitTagExists -Tag $currentTag) {
        Write-Host ""
        Write-Host -ForegroundColor Yellow "Tag Collision: Git tag '$currentTag' already exists in the repository."
        if ($NonInteractive) {
            Write-Host -ForegroundColor Red "Cannot resolve tag collision in non-interactive mode. Aborting."
            return $null
        }

        Write-Host "Please select an option to resolve the collision:"
        Write-Host "  (E) Edit the tag string manually"
        Write-Host "  (T) Prepend 'test-' to the existing tag name ('test-$currentTag')"
        Write-Host "  (Q) Quit the script and make no change (default)"
        Write-Host -NoNewline -ForegroundColor Cyan "Select option [E/T/Q] (default: Q): "
        $choice = Read-Host

        if ([string]::IsNullOrWhiteSpace($choice) -or $choice -imatch '^\s*q(uit)?\s*$') {
            return $null
        }
        elseif ($choice -imatch '^\s*t(est)?\s*$') {
            $currentTag = "test-$currentTag"
        }
        elseif ($choice -imatch '^\s*e(dit)?\s*$') {
            Write-Host -NoNewline -ForegroundColor Cyan "  Enter new tag name (press ENTER to cancel): "
            $customInput = Read-Host
            if ([string]::IsNullOrWhiteSpace($customInput)) {
                return $null
            }
            $candidate = $customInput.Trim()
            if (-not (Test-ValidGitTagName -Tag $candidate)) {
                Write-Host -ForegroundColor Red "Error: '$candidate' is not a valid Git tag name."
                continue
            }
            $currentTag = $candidate
        }
        else {
            Write-Host -ForegroundColor Yellow "Unrecognized option '$choice'. Defaulting to Quit."
            return $null
        }
    }

    return $currentTag
}

function Get-GitBranchInfo {
    param ([switch]$NoPush)

    $currentBranch = (git branch --show-current 2>$null).Trim()
    if ([string]::IsNullOrWhiteSpace($currentBranch)) {
        $currentBranch = "HEAD (detached)"
    }
    $canPush = (-not $NoPush) -and ($currentBranch -ne "HEAD (detached)")

    return [PSCustomObject]@{
        BranchName    = $currentBranch
        CanPushBranch = $canPush
    }
}

function Assert-GitWorkingTreeClean {
    param ([switch]$IsDryRun)

    $uncommitted = git status --porcelain=v1
    if (-not [string]::IsNullOrWhiteSpace($uncommitted)) {
        if ($IsDryRun) {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "Warning: Working directory has uncommitted changes. A clean repository is required for live execution:"
            $uncommitted -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | ForEach-Object {
                Write-Host -ForegroundColor Yellow "  $_"
            }
            Write-Host ""
        }
        else {
            Write-Host ""
            Write-Host -ForegroundColor Red "Error: Working directory has uncommitted changes. Repository must be clean before setting version or managing tags:"
            Write-Host ""
            $uncommitted -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) } | ForEach-Object {
                Write-Host -ForegroundColor Yellow "  $_"
            }
            Write-Host ""
            Write-Host "Please commit, stash, or revert these changes before setting version or pushing tags."
            exit 1
        }
    }
}

# ==============================================================================
# Helper Functions: Version Parsing, Formatting & Calculations
# ==============================================================================

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
    }
    catch {
        # Fallback to regex if ConvertFrom-Json fails
    }

    if ($raw -match '"productVersion"\s*:\s*"([^"]+)"') {
        return $Matches[1].Trim()
    }
    throw "Could not determine 'productVersion' from source of truth: $Path"
}

function Parse-VersionString {
    param ([string]$Tag)
    $clean = $Tag.Trim() -replace '^[vV]', ''
    $obj = [PSCustomObject]@{
        Raw              = $Tag
        Major            = 0
        Minor            = 0
        Patch            = 0
        Prerelease       = $null
        Stage            = $null
        StageNum         = $null
        StageDelimiter   = ""
        HasBuildMetadata = $false
        ValidationError  = $null
        IsValid          = $false
    }

    # Explicitly check for and disallow build metadata (+buildmetadata)
    if ($clean -match '\+') {
        $obj.HasBuildMetadata = $true
        $obj.ValidationError = "Build metadata ('+...') is not permitted in project versions for TextoVista/Windows compatibility."
        return $obj
    }

    # Match core SemVer X.Y.Z with optional prerelease
    if ($clean -notmatch '^(?<major>0|[1-9]\d*)\.(?<minor>0|[1-9]\d*)\.(?<patch>0|[1-9]\d*)(?:-(?<prerelease>.+))?$') {
        $obj.ValidationError = "Version '$Tag' does not match semantic version format (expected 'X.Y.Z' or 'vX.Y.Z', e.g., '1.0.0' or 'v1.0.0-rc1')."
        return $obj
    }

    $obj.Major = [int]$Matches.major
    $obj.Minor = [int]$Matches.minor
    $obj.Patch = [int]$Matches.patch

    # If no prerelease, valid full release version
    if (-not $Matches.prerelease) {
        $obj.IsValid = $true
        return $obj
    }

    $prerelease = $Matches.prerelease
    # Allowed prerelease formats: alpha, beta, rc with optional . or - or empty delimiter and optional 0-999 number
    if ($prerelease -notmatch '^(?<stage>alpha|beta|rc)(?:(?<delim>[.-])?(?<num>0|[1-9]\d*))?$') {
        $obj.ValidationError = "Prerelease stage '$prerelease' is not permitted. Only 'alpha', 'beta', or 'rc' stages (e.g., '-alpha.1', '-alpha-1', '-beta2', '-rc1') with numbers 0-999 are allowed."
        return $obj
    }

    $stage = $Matches.stage.ToLower()
    $delim = if ($Matches.delim) { $Matches.delim } else { "" }
    $num = if (-not [string]::IsNullOrWhiteSpace($Matches.num)) { [int]$Matches.num } else { 0 }

    if ($num -gt 999) {
        $obj.ValidationError = "Prerelease number '$num' exceeds maximum allowed value of 999."
        return $obj
    }

    $obj.Prerelease = $prerelease
    $obj.Stage = $stage
    $obj.StageNum = $num
    $obj.StageDelimiter = $delim
    $obj.IsValid = $true
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
    # Final releases use fixed 65432 so they always compare as newer than any prerelease.
    # Prereleases: alpha -> 1000 + n, beta -> 2000 + n, rc -> 3000 + n (where n is 0-999; defaults to 0 if omitted)
    # Supports delimiter styles: -alphaN, -alpha.N, and -alpha-N
    $stages = @('alpha', 'beta', 'rc')
    $fileVersionSuffix = 65432
    if ($vObj.Prerelease) {
        if ($vObj.Stage) {
            $stageIndex = $stages.IndexOf($vObj.Stage) + 1
            if ($stageIndex -gt 0) {
                $num = if ($null -ne $vObj.StageNum) { [int]$vObj.StageNum } else { 0 }
                $fileVersionSuffix = ($stageIndex * 1000) + $num
            }
            else {
                $fileVersionSuffix = 1000
            }
        }
        else {
            $fileVersionSuffix = 1000
        }
    }

    $fileVersion = "$($vObj.Major).$($vObj.Minor).$($vObj.Patch).$fileVersionSuffix"

    return [PSCustomObject]@{
        SemVer      = $semver
        TagName     = $tagName
        FileVersion = $fileVersion
    }
}

function Get-DefaultTagMessage {
    param (
        [string]$TagName,
        [string]$CustomMessage = ""
    )

    if (-not [string]::IsNullOrWhiteSpace($CustomMessage)) {
        return $CustomMessage
    }

    $parsed = Parse-VersionString -Tag $TagName
    if ($parsed.IsValid) {
        $cleanVer = $TagName.Trim() -replace '^[vV]', ''
        return "Version: $cleanVer"
    }

    return "Version: $TagName"
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
            throw "Invalid explicit version/tag provided: '$ExplicitTag'. $($parsed.ValidationError)"
        }
        return $parsed
    }

    $major = $CurrentObj.Major
    $minor = $CurrentObj.Minor
    $patch = $CurrentObj.Patch
    $prerelease = $null
    $stageName = $null
    $stageNum = $null
    $delim = ""

    # If -Release is requested, strip prerelease from current version
    if ($IsRelease) {
        return [PSCustomObject]@{
            Major          = $major
            Minor          = $minor
            Patch          = $patch
            Prerelease     = $null
            Stage          = $null
            StageNum       = $null
            StageDelimiter = ""
            IsValid        = $true
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
        }
        elseif ($CurrentObj.Prerelease) {
            $effectiveBump = "Prerelease"
        }
        else {
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
            $delim = if ($CurrentObj.StageDelimiter) { $CurrentObj.StageDelimiter } else { "" }
            if ($CurrentObj.Prerelease -and $CurrentObj.Stage) {
                $stageName = if ($RequestedStage) { $RequestedStage.ToLower() } else { $CurrentObj.Stage }
                $stageNum = if ($RequestedStage -and ($RequestedStage.ToLower() -ne $CurrentObj.Stage)) {
                    1
                }
                else {
                    $CurrentObj.StageNum + 1
                }
            }
            else {
                $patch++
                $stageName = if ($RequestedStage) { $RequestedStage.ToLower() } else { "alpha" }
                $stageNum = 1
            }
            if ($stageNum -gt 999) {
                throw "Cannot increment prerelease version beyond 999 for stage '$stageName'."
            }
            $prerelease = "${stageName}${delim}${stageNum}"
        }
    }

    # If a specific stage was passed with Major/Minor/Patch
    if ($RequestedStage -and ($effectiveBump -ne "Prerelease")) {
        $stageName = $RequestedStage.ToLower()
        $stageNum = 1
        $delim = if ($CurrentObj.StageDelimiter) { $CurrentObj.StageDelimiter } else { "" }
        $prerelease = "${stageName}${delim}${stageNum}"
    }

    return [PSCustomObject]@{
        Major          = $major
        Minor          = $minor
        Patch          = $patch
        Prerelease     = $prerelease
        Stage          = $stageName
        StageNum       = $stageNum
        StageDelimiter = $delim
        IsValid        = $true
    }
}

# ==============================================================================
# Helper Functions: File Update Engine
# ==============================================================================

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
                        Old        = $origLines[$i]
                        New        = $updLines[$i]
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
        }
        else {
            Write-Host -ForegroundColor DarkGray "  No changes needed: $($plan.RelPath)"
        }
    }
}

function Sync-FrontendPackageLock {
    param (
        [string]$ProjectRoot
    )

    $frontendDir = Join-Path $ProjectRoot "frontend"
    if (-not (Test-Path -Path $frontendDir -PathType Container)) {
        Write-Host -ForegroundColor Red "Error: 'frontend' directory not found at: $frontendDir"
        exit 1
    }

    $npmCmd = Get-Command "npm" -ErrorAction SilentlyContinue
    if (-not $npmCmd) {
        Write-Host -ForegroundColor Red "Error: 'npm' command not found in PATH. Node.js/npm is required to update package-lock.json."
        exit 1
    }

    Write-Host -ForegroundColor Cyan "Syncing frontend/package-lock.json via 'npm i --package-lock-only'..."
    Push-Location $frontendDir
    try {
        $npmOutput = & npm i --package-lock-only 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: 'npm i --package-lock-only' failed with exit code $LASTEXITCODE."
            if ($npmOutput) {
                $npmOutput | ForEach-Object { Write-Host -ForegroundColor Red "  $_" }
            }
            exit 1
        }
        Write-Host -ForegroundColor Green "  frontend/package-lock.json synced successfully."
    }
    catch {
        Write-Host -ForegroundColor Red "Error executing 'npm i --package-lock-only': $_"
        exit 1
    }
    finally {
        Pop-Location
    }
}

# ==============================================================================
# Helper Functions: Reporting & Display
# ==============================================================================

function Show-VersionInformation {
    param (
        $CurrentFormatted,
        [string]$MostRecentGitTag,
        [string]$ProjectRoot,
        $TargetFiles
    )

    Write-Host ""
    Write-Host -ForegroundColor Cyan "TextoVista Version Information:"
    Write-Host -NoNewline "  Source of Truth (wails.json) : "; Write-Host -ForegroundColor Green "$($CurrentFormatted.SemVer)"
    Write-Host -NoNewline "  Windows File Version         : "; Write-Host -ForegroundColor Green "$($CurrentFormatted.FileVersion)"
    if ($MostRecentGitTag) {
        Write-Host -NoNewline "  Most Recent Git Tag          : "; Write-Host -ForegroundColor Cyan "$MostRecentGitTag"
    }
    else {
        Write-Host -NoNewline "  Most Recent Git Tag          : "; Write-Host -ForegroundColor Yellow "No tags found"
    }

    $uncommittedCheck = git status --porcelain=v1
    Write-Host -NoNewline "  Working Tree Status          : "
    if ([string]::IsNullOrWhiteSpace($uncommittedCheck)) {
        Write-Host -ForegroundColor Green "Clean"
    }
    else {
        $dirtyCount = ($uncommittedCheck -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }).Count
        Write-Host -ForegroundColor Yellow "Dirty ($dirtyCount uncommitted change(s))"
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
}

function Show-PlannedOperations {
    param (
        $CurrentFormatted,
        [string]$MostRecentGitTag,
        [string]$NewSemVer,
        [string]$NewFileVer,
        [string]$FinalCommitMessage,
        [string]$CurrentBranch,
        [bool]$NoPush,
        [bool]$CreateVersionTag,
        [string]$ProposedTagName,
        [string]$FinalTagMessage,
        [bool]$PushTag,
        [bool]$CurrentVersion,
        $DiffPlans
    )

    Write-Host ""
    Write-Host -ForegroundColor Cyan "=========================================================="
    if ($CurrentVersion) {
        Write-Host -ForegroundColor Cyan "             TextoVista Version Tag Plan                  "
    }
    else {
        Write-Host -ForegroundColor Cyan "             TextoVista Version Bump Plan                 "
    }
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host "  Source of Truth (wails.json) : $($CurrentFormatted.SemVer) (Win: $($CurrentFormatted.FileVersion))"
    if ($MostRecentGitTag) {
        Write-Host "  Most Recent Git Tag          : $MostRecentGitTag"
    }

    if ($CurrentVersion) {
        Write-Host -NoNewline "  Target Version               : "; Write-Host -ForegroundColor Green "$NewSemVer (Existing version unchanged)"
        Write-Host "  Version Bump Operations      : Skipped (-CurrentVersion specified; project files will not be modified)"
    }
    else {
        Write-Host -NoNewline "  Target Version               : "; Write-Host -ForegroundColor Green "$NewSemVer (Win: $NewFileVer)"
        Write-Host -NoNewline "  Commit Message               : "; Write-Host -ForegroundColor Yellow "$FinalCommitMessage"
        Write-Host -NoNewline "  Git Branch                   : "; Write-Host "$CurrentBranch"
        Write-Host -NoNewline "  Lockfile Sync                : "; Write-Host -ForegroundColor Cyan "npm i --package-lock-only (in frontend/)"
        Write-Host -NoNewline "  Push Commit to Remote        : "
        if ($NoPush) {
            Write-Host -ForegroundColor Yellow "No (-NoPush specified: Local commit only)"
        }
        elseif ($CurrentBranch -eq "HEAD (detached)") {
            Write-Host -ForegroundColor Yellow "No (Detached HEAD: Local commit only)"
        }
        else {
            Write-Host -ForegroundColor Green "Yes (origin/$CurrentBranch)"
        }
    }

    Write-Host -NoNewline "  Create Git Tag Locally       : "
    if ($CreateVersionTag) {
        Write-Host -ForegroundColor Green "Yes (-CreateVersionTag specified: Proposed tag '$ProposedTagName')"
    }
    else {
        Write-Host -ForegroundColor DarkGray "No (Default: tag will not be created)"
    }
    if ($CreateVersionTag) {
        Write-Host -NoNewline "  Tag Message                  : "; Write-Host -ForegroundColor Yellow "$FinalTagMessage"
    }

    Write-Host -NoNewline "  Push Git Tag to Remote       : "
    if ($PushTag) {
        Write-Host -ForegroundColor Green "Yes (-PushTag specified)"
    }
    else {
        Write-Host -ForegroundColor DarkGray "No (Local tag only if created)"
    }

    if (-not $CurrentVersion) {
        Write-Host ""
        Write-Host -ForegroundColor Cyan "Target File Diffs:"
        foreach ($plan in $DiffPlans) {
            if ($plan.IsModified) {
                Write-Host "  File: $($plan.RelPath)" -ForegroundColor Magenta
                foreach ($d in $plan.Diffs) {
                    Write-Host "    - $($d.Old.Trim())" -ForegroundColor Red
                    Write-Host "    + $($d.New.Trim())" -ForegroundColor Green
                }
            }
            else {
                Write-Host "  File: $($plan.RelPath) (already at target version)" -ForegroundColor DarkGray
            }
        }
    }
    Write-Host ""
}

function Show-DryRunPlan {
    param (
        [bool]$CurrentVersion,
        $DiffPlans,
        [string]$FinalCommitMessage,
        [bool]$CanPushBranch,
        [string]$CurrentBranch,
        [bool]$CreateVersionTag,
        [string]$ProposedTagName,
        [string]$FinalTagMessage,
        [bool]$PushTag,
        [string]$MostRecentGitTag
    )

    Write-Host -ForegroundColor Yellow "[DryRun / WhatIf Mode] Simulation complete. No files or Git references modified."
    Write-Host ""
    Write-Host -ForegroundColor Cyan "Planned Git and build commands that would execute:"
    if (-not $CurrentVersion) {
        Write-Host "  (cd frontend && npm i --package-lock-only)"
        $relFiles = @($DiffPlans | Where-Object { $_.IsModified } | ForEach-Object { $_.RelPath })
        if (Test-Path "frontend/package-lock.json") {
            $relFiles += "frontend/package-lock.json"
        }
        $relList = ($relFiles | Select-Object -Unique) -join " "
        if (-not [string]::IsNullOrWhiteSpace($relList)) {
            Write-Host "  git add $relList"
            Write-Host "  git commit -m `"$FinalCommitMessage`""
            if ($CanPushBranch) {
                Write-Host "  git push origin $CurrentBranch"
            }
        }
    }
    if ($CreateVersionTag) {
        Write-Host "  git tag -a `"$ProposedTagName`" -m `"$FinalTagMessage`""
        if ($PushTag) {
            Write-Host "  git push origin $ProposedTagName"
        }
    }
    elseif ($PushTag -and $MostRecentGitTag) {
        Write-Host "  git push origin $MostRecentGitTag (after user prompt)"
    }
    Write-Host ""
}

# ==============================================================================
# Helper Functions: Interactive Prompts & Git Execution
# ==============================================================================

function Invoke-StandaloneTagPush {
    param (
        [string]$MostRecentTag,
        [switch]$IsDryRun,
        [switch]$NonInteractive
    )

    Write-Host ""
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host -ForegroundColor Cyan "             TextoVista Remote Tag Push                   "
    Write-Host -ForegroundColor Cyan "=========================================================="
    if (-not $MostRecentTag) {
        Write-Host -ForegroundColor Red "Error: No Git tags found in the repository to push."
        exit 1
    }

    Write-Host -NoNewline "  Most Recent Git Tag : "; Write-Host -ForegroundColor Green "$MostRecentTag"
    Write-Host -NoNewline "  Action              : "; Write-Host -ForegroundColor Yellow "Push existing tag '$MostRecentTag' to remote 'origin'"
    Write-Host ""

    if ($IsDryRun) {
        Write-Host -ForegroundColor Yellow "[DryRun / WhatIf Mode] Simulation complete."
        Write-Host -ForegroundColor Cyan "Planned Git command:"
        Write-Host "  git push origin $MostRecentTag"
        Write-Host ""
        return
    }

    $confirmedPush = $false
    if ($NonInteractive) {
        $confirmedPush = $true
    }
    else {
        Write-Host -NoNewline -ForegroundColor Cyan "Push most recent tag '$MostRecentTag' to remote 'origin'? [Y]es / [N]o (default: Y): "
        $resp = Read-Host
        if ([string]::IsNullOrWhiteSpace($resp) -or $resp -imatch '^\s*y(es)?\s*$') {
            $confirmedPush = $true
        }
    }

    if ($confirmedPush) {
        Write-Host -ForegroundColor Cyan "Pushing tag '$MostRecentTag' to remote 'origin'..."
        git push origin "$MostRecentTag"
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to push tag '$MostRecentTag' to remote."
            exit 1
        }
        Write-Host -ForegroundColor Green "Tag '$MostRecentTag' successfully pushed to remote!"
    }
    else {
        Write-Host -ForegroundColor Yellow "Tag push cancelled by user."
    }
}

function Confirm-VersionBumpPrompt {
    param (
        [string]$NewSemVer,
        [string]$NewFileVer,
        [string]$ProposedTagName,
        [string]$FinalCommitMessage,
        $DiffPlans,
        [string]$ProjectRoot,
        [switch]$NonInteractive
    )

    if ($NonInteractive) {
        return [PSCustomObject]@{
            Confirmed          = $true
            NewSemVer          = $NewSemVer
            NewFileVer         = $NewFileVer
            ProposedTagName    = $ProposedTagName
            FinalCommitMessage = $FinalCommitMessage
            DiffPlans          = $DiffPlans
        }
    }

    Write-Host -NoNewline -ForegroundColor Cyan "Apply version bump to '$NewSemVer', stage files, and commit? [Y]es / [N]o / [E]dit version or message (default: Y): "
    $response = Read-Host
    if ([string]::IsNullOrWhiteSpace($response) -or $response -imatch '^\s*y(es)?\s*$') {
        return [PSCustomObject]@{
            Confirmed          = $true
            NewSemVer          = $NewSemVer
            NewFileVer         = $NewFileVer
            ProposedTagName    = $ProposedTagName
            FinalCommitMessage = $FinalCommitMessage
            DiffPlans          = $DiffPlans
        }
    }
    elseif ($response -imatch '^\s*e(dit)?\s*$') {
        $validVerEntered = $false
        while (-not $validVerEntered) {
            $customVer = Read-Host "  Enter version to set (press ENTER to keep '$NewSemVer', 'q' to abort)"
            if ([string]::IsNullOrWhiteSpace($customVer)) {
                $validVerEntered = $true
                break
            }
            if ($customVer -imatch '^\s*q(uit)?\s*$') {
                Write-Host -ForegroundColor Yellow "Operation cancelled by user. No changes made."
                return $null
            }
            $updatedSemVer = $customVer.Trim() -replace '^[vV]', ''
            $parsedCustom = Parse-VersionString -Tag $updatedSemVer
            if ($parsedCustom.IsValid) {
                $nextFormatted = Format-BuildVersions $parsedCustom
                $NewSemVer = $nextFormatted.SemVer
                $NewFileVer = $nextFormatted.FileVersion
                $ProposedTagName = $nextFormatted.TagName
                $DiffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $NewSemVer -FileVersion $NewFileVer
                $validVerEntered = $true
            }
            else {
                if ($parsedCustom.HasBuildMetadata) {
                    Write-Host ""
                    Write-Host -ForegroundColor Yellow "Warning: Build metadata ('+...') is not permitted in project versions."
                    Write-Host -ForegroundColor Yellow "TextoVista requires versions without build metadata for Windows installer compatibility."
                }
                else {
                    Write-Host ""
                    Write-Host -ForegroundColor Red "Error: $($parsedCustom.ValidationError)"
                }
            }
        }
        $customMsg = Read-Host "  Enter commit message (press ENTER to keep '$FinalCommitMessage')"
        if (-not [string]::IsNullOrWhiteSpace($customMsg)) {
            $FinalCommitMessage = $customMsg
        }
        Write-Host ""
        Write-Host -ForegroundColor Cyan "Updated Target Version: $NewSemVer | Commit Message: $FinalCommitMessage"
        $confirm2 = Read-Host "Proceed with these updated values? [Y/n]"
        if ($confirm2 -inotmatch '^\s*(y(es)?|$)\s*$') {
            Write-Host -ForegroundColor Yellow "Operation cancelled by user. No changes made."
            return $null
        }
        return [PSCustomObject]@{
            Confirmed          = $true
            NewSemVer          = $NewSemVer
            NewFileVer         = $NewFileVer
            ProposedTagName    = $ProposedTagName
            FinalCommitMessage = $FinalCommitMessage
            DiffPlans          = $DiffPlans
        }
    }
    else {
        Write-Host -ForegroundColor Yellow "Operation cancelled by user. No changes made."
        return $null
    }
}

function Invoke-VersionBumpCommit {
    param (
        $DiffPlans,
        [string]$CommitMessage,
        [string]$CurrentBranch,
        [bool]$CanPushBranch
    )

    Write-Host -ForegroundColor Cyan "Staging files in Git..."
    $modifiedFiles = @($DiffPlans | Where-Object { $_.IsModified } | ForEach-Object { $_.RelPath })
    $lockfilePath = "frontend/package-lock.json"
    if (Test-Path -Path $lockfilePath -PathType Leaf) {
        $lockfileStatus = git status --porcelain -- $lockfilePath 2>$null
        if (-not [string]::IsNullOrWhiteSpace($lockfileStatus)) {
            $modifiedFiles += $lockfilePath
        }
    }
    $modifiedFiles = @($modifiedFiles | Select-Object -Unique)
    if ($modifiedFiles.Count -gt 0) {
        git add $modifiedFiles 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to stage files in Git."
            exit 1
        }

        Write-Host -ForegroundColor Cyan "Committing changes: $CommitMessage"
        git commit -m "$CommitMessage" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to commit changes to Git."
            exit 1
        }
        Write-Host -ForegroundColor Green "  Commit created successfully."

        # Push Version Bump Commit to Remote
        if ($CanPushBranch) {
            Write-Host -ForegroundColor Cyan "Pushing version bump commit to remote 'origin/$CurrentBranch'..."
            git push origin "$CurrentBranch" 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host -ForegroundColor Yellow "Warning: Failed to push branch '$CurrentBranch' to remote. You may need to push manually."
            }
            else {
                Write-Host -ForegroundColor Green "  Version bump commit pushed successfully to remote."
            }
        }
        else {
            Write-Host -ForegroundColor Yellow "  Skipped remote commit push (NoPush specified or detached HEAD)."
        }
    }
    else {
        Write-Host -ForegroundColor Yellow "  No file changes to commit (files already at target version)."
    }
}

function Test-ChangelogEntry {
    param (
        [string]$ChangelogPath,
        [string]$TagName
    )

    if (-not (Test-Path -Path $ChangelogPath -PathType Leaf)) {
        return [PSCustomObject]@{
            Exists     = $false
            HasEntry   = $false
            HasContent = $false
            Message    = "CHANGELOG.md was not found at '$ChangelogPath'."
        }
    }

    $lines = Get-Content -Path $ChangelogPath -Encoding utf8
    $escapedTag = [regex]::Escape($TagName)
    $headerRegex = "^##\s+\[?$escapedTag\]?(\s+-\s+.*)?$"
    $start = $null
    $end = $lines.Count

    for ($i = 0; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match $headerRegex) {
            $start = $i + 1
            continue
        }
        if ($null -ne $start -and $lines[$i] -match '^##\s+') {
            $end = $i
            break
        }
    }

    if ($null -eq $start) {
        return [PSCustomObject]@{
            Exists     = $true
            HasEntry   = $false
            HasContent = $false
            Message    = "No matching section header found for tag '$TagName' in CHANGELOG.md (expected header like '## $TagName - YYYY-MM-DD')."
        }
    }

    $bodyLines = if ($start -lt $end) { $lines[$start..($end - 1)] } else { @() }
    $body = ($bodyLines -join "`n").Trim()

    if ([string]::IsNullOrWhiteSpace($body)) {
        return [PSCustomObject]@{
            Exists     = $true
            HasEntry   = $true
            HasContent = $false
            Message    = "Section header for tag '$TagName' was found in CHANGELOG.md, but contains no release notes content."
        }
    }

    return [PSCustomObject]@{
        Exists     = $true
        HasEntry   = $true
        HasContent = $true
        Message    = "Valid release notes entry found for tag '$TagName'."
    }
}

function Invoke-TagOperations {
    param (
        [bool]$CreateVersionTag,
        [string]$ProposedTagName,
        [string]$TagMessage,
        [bool]$PushTag,
        [switch]$NonInteractive,
        [string]$ProjectRoot = ""
    )

    if ([string]::IsNullOrWhiteSpace($ProjectRoot)) {
        $ProjectRoot = $script:ProjectRoot
    }

    $tagToApply = $ProposedTagName

    if ($CreateVersionTag) {
        $tagMessageToUse = Get-DefaultTagMessage -TagName $tagToApply -CustomMessage $TagMessage

        # Prompt user to verify or change the proposed tag
        $confirmedTag = $false
        if ($NonInteractive) {
            $confirmedTag = $true
        }
        else {
            Write-Host ""
            Write-Host -NoNewline -ForegroundColor Cyan "Create local Git tag '$ProposedTagName' (message: '$tagMessageToUse')? [Y]es / [N]o / [E]dit tag name (default: Y): "
            $tagResp = Read-Host
            if ([string]::IsNullOrWhiteSpace($tagResp) -or $tagResp -imatch '^\s*y(es)?\s*$') {
                $confirmedTag = $true
            }
            elseif ($tagResp -imatch '^\s*e(dit)?\s*$') {
                $validTagEntered = $false
                while (-not $validTagEntered) {
                    $customTagInput = Read-Host "  Enter tag name (press ENTER to keep '$ProposedTagName', 'q' to skip)"
                    if ([string]::IsNullOrWhiteSpace($customTagInput)) {
                        $validTagEntered = $true
                        $confirmedTag = $true
                        break
                    }
                    if ($customTagInput -imatch '^\s*q(uit)?\s*$') {
                        Write-Host -ForegroundColor Yellow "Skipped Git tag creation."
                        $confirmedTag = $false
                        break
                    }
                    $candidateTag = $customTagInput.Trim()
                    if (-not (Test-ValidGitTagName -Tag $candidateTag)) {
                        Write-Host ""
                        Write-Host -ForegroundColor Red "Error: '$candidateTag' is not a valid Git tag name (contains invalid characters or illegal format)."
                        continue
                    }

                    $parsedTag = Parse-VersionString -Tag $candidateTag
                    if ($parsedTag.IsValid) {
                        $tagToApply = if ($candidateTag -match '^[vV]') { $candidateTag } else { "v$candidateTag" }
                        $validTagEntered = $true
                        $confirmedTag = $true
                    }
                    else {
                        Write-Host ""
                        if ($parsedTag.HasBuildMetadata) {
                            Write-Host -ForegroundColor Yellow "Warning: Tag '$candidateTag' contains build metadata ('+...'), which does not match normal project SemVer format."
                        }
                        else {
                            Write-Host -ForegroundColor Yellow "Warning: Tag '$candidateTag' does not match normal project SemVer format (e.g., 'vX.Y.Z' or 'vX.Y.Z-rc1')."
                            if (-not [string]::IsNullOrWhiteSpace($parsedTag.ValidationError)) {
                                Write-Host -ForegroundColor Yellow "  Details: $($parsedTag.ValidationError)"
                            }
                        }
                        Write-Host -NoNewline -ForegroundColor Cyan "  Use non-SemVer tag '$candidateTag' anyway? [Y]es / [N]o (default: Y): "
                        $useNonSemVer = Read-Host
                        if ([string]::IsNullOrWhiteSpace($useNonSemVer) -or $useNonSemVer -imatch '^\s*y(es)?\s*$') {
                            $tagToApply = $candidateTag
                            $validTagEntered = $true
                            $confirmedTag = $true
                        }
                        else {
                            Write-Host -ForegroundColor Yellow "  Tag discarded. Please enter a new tag name."
                        }
                    }
                }
            }
            else {
                Write-Host -ForegroundColor Yellow "Skipped Git tag creation."
            }
        }

        if ($confirmedTag) {
            if (Test-GitTagExists -Tag $tagToApply) {
                $resolvedTagCollision = Resolve-TagCollision -TagName $tagToApply -NonInteractive:$NonInteractive
                if (-not $resolvedTagCollision) {
                    Write-Host -ForegroundColor Yellow "Skipped Git tag creation due to unresolved tag collision."
                    $confirmedTag = $false
                }
                else {
                    $tagToApply = $resolvedTagCollision
                }
            }
        }

        if ($confirmedTag) {
            # Check CHANGELOG.md for semantic version tags (vX.Y.Z or vX.Y.Z-prerelease)
            $parsedTagForChangelog = Parse-VersionString -Tag $tagToApply
            if ($parsedTagForChangelog.IsValid) {
                $changelogFile = Join-Path $ProjectRoot "CHANGELOG.md"
                $changelogCheck = Test-ChangelogEntry -ChangelogPath $changelogFile -TagName $tagToApply
                if (-not ($changelogCheck.HasEntry -and $changelogCheck.HasContent)) {
                    Write-Host ""
                    Write-Host -ForegroundColor Red "ERROR: Cannot assign version tag '$tagToApply' because CHANGELOG.md is not up to date."
                    Write-Host -ForegroundColor Yellow "  $($changelogCheck.Message)"
                    Write-Host ""
                    Write-Host -ForegroundColor Yellow "Please update CHANGELOG.md with release notes for '$tagToApply' and commit the change."
                    Write-Host -ForegroundColor Cyan   "Once CHANGELOG.md is updated, you can create and push the version tag using:"
                    if ($PushTag) {
                        Write-Host -ForegroundColor Cyan "  ./scripts/Set-Version.ps1 -CreateVersionTag -CurrentVersion -PushTag"
                    }
                    else {
                        Write-Host -ForegroundColor Cyan "  ./scripts/Set-Version.ps1 -CreateVersionTag -CurrentVersion"
                    }
                    Write-Host ""
                    $confirmedTag = $false
                }
            }
            else {
                Write-Host -ForegroundColor Yellow "  Notice: Proposed tag '$tagToApply' does not use normal semantic 'vX.Y.Z' format; skipping CHANGELOG.md check."
            }
        }

        if ($confirmedTag) {
            $tagMessageToUse = Get-DefaultTagMessage -TagName $tagToApply -CustomMessage $TagMessage
            Write-Host -ForegroundColor Cyan "Creating local Git tag: $tagToApply (message: '$tagMessageToUse')"
            git tag -a "$tagToApply" -m "$tagMessageToUse" 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host -ForegroundColor Red "Error: Failed to create Git tag '$tagToApply'."
                exit 1
            }
            Write-Host -ForegroundColor Green "  Tag '$tagToApply' created locally."

            # Check if tag should be pushed to remote
            if ($PushTag) {
                Write-Host -ForegroundColor Cyan "Pushing tag '$tagToApply' to remote 'origin'..."
                git push origin "$tagToApply" 2>&1 | Out-Null
                if ($LASTEXITCODE -ne 0) {
                    Write-Host -ForegroundColor Yellow "Warning: Failed to push tag '$tagToApply' to remote. You may need to push manually."
                }
                else {
                    Write-Host -ForegroundColor Green "  Tag '$tagToApply' pushed successfully to remote."
                }
            }
            else {
                Write-Host -ForegroundColor Yellow "  Tag '$tagToApply' exists locally only. (To push to remote, run: git push origin $tagToApply or rerun with -PushTag)."
            }
            return $true
        }
        else {
            return $false
        }
    }
    elseif ($PushTag) {
        $latestTag = Get-MostRecentTag
        Write-Host ""
        if (-not $latestTag) {
            Write-Host -ForegroundColor Yellow "Warning: -PushTag was requested without -CreateVersionTag, but no Git tags were found in repository."
        }
        else {
            $confirmLatestPush = $false
            if ($NonInteractive) {
                $confirmLatestPush = $true
            }
            else {
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
                }
                else {
                    Write-Host -ForegroundColor Green "  Tag '$latestTag' pushed successfully to remote."
                }
            }
            else {
                Write-Host -ForegroundColor Yellow "Skipped remote tag push."
            }
        }
    }
}

# ==============================================================================
# Main Controller Logic
# ==============================================================================

Push-Location $ProjectRoot
try {
    $isDryRun = $DryRun -or $PSCmdlet.MyInvocation.BoundParameters.ContainsKey('WhatIf')
    $stage = Resolve-PrereleaseStage -Stage $PrereleaseStage -Alpha:$SetAlpha -Beta:$SetBeta -RC:$SetReleaseCandidate
    $bumpAction = Resolve-BumpAction -Bump $Bump -Major:$IncrementMajor -Minor:$IncrementMinor -Patch:$IncrementPatch -Prerelease:$IncrementPrerelease

    Assert-ParameterRules `
        -CurrentVersion:$CurrentVersion `
        -CreateVersionTag:$CreateVersionTag `
        -NewVersion $NewVersion `
        -Bump $Bump `
        -IncrementMajor:$IncrementMajor `
        -IncrementMinor:$IncrementMinor `
        -IncrementPatch:$IncrementPatch `
        -IncrementPrerelease:$IncrementPrerelease `
        -PrereleaseStage $PrereleaseStage `
        -SetAlpha:$SetAlpha `
        -SetBeta:$SetBeta `
        -SetReleaseCandidate:$SetReleaseCandidate `
        -ReleaseVersion:$ReleaseVersion

    $hasExplicitBumpInput = Test-ExplicitBumpInputSpecified `
        -NewVersion $NewVersion `
        -Bump $Bump `
        -IncrementMajor:$IncrementMajor `
        -IncrementMinor:$IncrementMinor `
        -IncrementPatch:$IncrementPatch `
        -IncrementPrerelease:$IncrementPrerelease `
        -PrereleaseStage $PrereleaseStage `
        -SetAlpha:$SetAlpha `
        -SetBeta:$SetBeta `
        -SetReleaseCandidate:$SetReleaseCandidate `
        -ReleaseVersion:$ReleaseVersion `
        -CreateVersionTag:$CreateVersionTag

    # 0. Validate -NewVersion if specified
    if (-not [string]::IsNullOrWhiteSpace($NewVersion)) {
        $parsedExplicit = Parse-VersionString -Tag $NewVersion
        while (-not $parsedExplicit.IsValid) {
            if ($parsedExplicit.HasBuildMetadata) {
                Write-Host ""
                Write-Host -ForegroundColor Yellow "Warning: Build metadata ('+...') is not permitted in project versions."
                Write-Host -ForegroundColor Yellow "TextoVista requires versions without build metadata for Windows installer compatibility."
            }
            else {
                Write-Host ""
                Write-Host -ForegroundColor Red "Error: $($parsedExplicit.ValidationError)"
            }

            if ($NonInteractive) {
                Write-Host -ForegroundColor Red "Invalid explicit version '$NewVersion' in non-interactive mode. Aborting."
                exit 1
            }

            Write-Host -NoNewline -ForegroundColor Cyan "Enter a valid semantic version (e.g., '1.0.0', '1.0.0-rc.1', or '1.0.0-rc-1', or press ENTER/'q' to abort): "
            $promptVer = Read-Host
            if ([string]::IsNullOrWhiteSpace($promptVer) -or $promptVer -imatch '^\s*q(uit)?\s*$') {
                Write-Host -ForegroundColor Yellow "Operation cancelled by user. Aborting."
                exit 1
            }
            $NewVersion = $promptVer.Trim()
            $parsedExplicit = Parse-VersionString -Tag $NewVersion
        }
    }

    # 1. Read Current Version from wails.json (Source of Truth)
    $sourceVersionStr = Get-CurrentVersionFromWails -Path $WailsConfigPath
    $currentObj = Parse-VersionString -Tag $sourceVersionStr
    if (-not $currentObj.IsValid) {
        Write-Host -ForegroundColor Red "Error: Failed to parse version '$sourceVersionStr' from $WailsConfigPath as valid SemVer: $($currentObj.ValidationError)"
        exit 1
    }
    $currentFormatted = Format-BuildVersions $currentObj
    $mostRecentGitTag = Get-MostRecentTag

    # 2. Handle -ShowCurrentVersion
    if ($ShowCurrentVersion) {
        Show-VersionInformation `
            -CurrentFormatted $currentFormatted `
            -MostRecentGitTag $mostRecentGitTag `
            -ProjectRoot $ProjectRoot `
            -TargetFiles $TargetFiles
        return
    }

    # 3. Verify working tree is clean (no uncommitted changes allowed)
    Assert-GitWorkingTreeClean -IsDryRun:$isDryRun

    # 4. Handle Standalone -PushTag Mode
    if ($PushTag -and (-not $CreateVersionTag) -and (-not $hasExplicitBumpInput)) {
        Invoke-StandaloneTagPush `
            -MostRecentTag $mostRecentGitTag `
            -IsDryRun:$isDryRun `
            -NonInteractive:$NonInteractive
        return
    }

    # 5. Verify all target files exist
    if (-not (Test-TargetFilesExist -BaseDir $ProjectRoot)) {
        Write-Host -ForegroundColor Red "Aborting: One or more target files are missing."
        exit 1
    }

    if ((-not $CurrentVersion) -and (-not (Get-Command "npm" -ErrorAction SilentlyContinue))) {
        Write-Host -ForegroundColor Red "Error: 'npm' command not found in PATH. Node.js/npm is required to sync frontend/package-lock.json."
        exit 1
    }

    # 6. Calculate target version using wails.json version as baseline
    if ($CurrentVersion) {
        $nextObj = $currentObj
    }
    else {
        $nextObj = Calculate-NextVersion `
            -CurrentObj $currentObj `
            -ExplicitTag $NewVersion `
            -RequestedBump $bumpAction `
            -RequestedStage $stage `
            -IsRelease:($ReleaseVersion -eq $true)
    }

    $nextFormatted = Format-BuildVersions $nextObj
    $newSemVer = $nextFormatted.SemVer
    $newFileVer = $nextFormatted.FileVersion
    $proposedTagName = $nextFormatted.TagName

    # 7. Check if proposed tag already exists when -CreateVersionTag is specified
    if ($CreateVersionTag -and (Test-GitTagExists -Tag $proposedTagName)) {
        $resolvedTag = Resolve-TagCollision -TagName $proposedTagName -NonInteractive:$NonInteractive
        if (-not $resolvedTag) {
            Write-Host ""
            Write-Host -ForegroundColor Yellow "Tag collision unresolved. Operation cancelled by user. No changes made."
            return
        }
        $proposedTagName = $resolvedTag
        Write-Host -ForegroundColor Green "Resolved tag name to apply: '$proposedTagName'"
    }

    # 8. Prepare commit message and tag message
    $finalCommitMessage = if ([string]::IsNullOrWhiteSpace($Message)) {
        "chore(release): bump version to $newSemVer"
    }
    else {
        $Message
    }

    $finalTagMessage = Get-DefaultTagMessage -TagName $proposedTagName -CustomMessage $TagMessage

    # 9. Check current branch and push capability
    $branchInfo = Get-GitBranchInfo -NoPush:$NoPush

    # 10. Compute file diff plan
    $diffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $newSemVer -FileVersion $newFileVer

    # 11. Display Planned Operations Summary
    Show-PlannedOperations `
        -CurrentFormatted $currentFormatted `
        -MostRecentGitTag $mostRecentGitTag `
        -NewSemVer $newSemVer `
        -NewFileVer $newFileVer `
        -FinalCommitMessage $finalCommitMessage `
        -CurrentBranch $branchInfo.BranchName `
        -NoPush:$NoPush `
        -CreateVersionTag:$CreateVersionTag `
        -ProposedTagName $proposedTagName `
        -FinalTagMessage $finalTagMessage `
        -PushTag:$PushTag `
        -CurrentVersion:$CurrentVersion `
        -DiffPlans $diffPlans

    # 12. Handle -WhatIf / -DryRun
    if ($isDryRun) {
        Show-DryRunPlan `
            -CurrentVersion:$CurrentVersion `
            -DiffPlans $diffPlans `
            -FinalCommitMessage $finalCommitMessage `
            -CanPushBranch:$branchInfo.CanPushBranch `
            -CurrentBranch $branchInfo.BranchName `
            -CreateVersionTag:$CreateVersionTag `
            -ProposedTagName $proposedTagName `
            -FinalTagMessage $finalTagMessage `
            -PushTag:$PushTag `
            -MostRecentGitTag $mostRecentGitTag
        return
    }

    # 13. Execute Version Bump Operations (Skipped when -CurrentVersion is used)
    if (-not $CurrentVersion) {
        $bumpPlan = Confirm-VersionBumpPrompt `
            -NewSemVer $newSemVer `
            -NewFileVer $newFileVer `
            -ProposedTagName $proposedTagName `
            -FinalCommitMessage $finalCommitMessage `
            -DiffPlans $diffPlans `
            -ProjectRoot $ProjectRoot `
            -NonInteractive:$NonInteractive

        if (-not $bumpPlan) {
            return
        }

        $newSemVer = $bumpPlan.NewSemVer
        $newFileVer = $bumpPlan.NewFileVer
        $proposedTagName = $bumpPlan.ProposedTagName
        $finalCommitMessage = $bumpPlan.FinalCommitMessage
        $diffPlans = $bumpPlan.DiffPlans

        Write-Host -ForegroundColor Cyan "Applying file updates..."
        Apply-FileUpdates -DiffPlans $diffPlans

        Sync-FrontendPackageLock -ProjectRoot $ProjectRoot

        Invoke-VersionBumpCommit `
            -DiffPlans $diffPlans `
            -CommitMessage $finalCommitMessage `
            -CurrentBranch $branchInfo.BranchName `
            -CanPushBranch:$branchInfo.CanPushBranch
    }

    # 14. Tag Handling Logic
    $tagSuccess = Invoke-TagOperations `
        -CreateVersionTag:$CreateVersionTag `
        -ProposedTagName $proposedTagName `
        -TagMessage $TagMessage `
        -PushTag:$PushTag `
        -NonInteractive:$NonInteractive `
        -ProjectRoot $ProjectRoot

    Write-Host ""
    if ($CreateVersionTag -and -not $tagSuccess) {
        Write-Host -ForegroundColor Yellow "Set-Version finished: Version tag creation was skipped or aborted."
    }
    else {
        Write-Host -ForegroundColor Green "Set-Version completed successfully!"
    }
    Write-Host ""
}
finally {
    Pop-Location
}
