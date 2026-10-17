<#
    .SYNOPSIS
        Creates a new Git tag and updates version numbers across project files.
    .DESCRIPTION
        Calculates and applies the next semantic version and Windows 4-part numeric
        version to wails.json, package.json, info.json, and project.nsi. Commits
        the changes, creates an annotated Git tag, and pushes to remote.
    .PARAMETER TagName
        The explicit version or Git tag to set (e.g., 'v0.6.0' or '0.6.0').
    .PARAMETER Message
        The commit and tag message. Defaults to 'Version X.Y.Z (vX.Y.Z)'.
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
    .PARAMETER ShowCurrentVersion
        Show current Git tag and project file versions without making changes. Alias: -Show, -Current.
    .PARAMETER DryRun
        Preview all file changes and Git operations without modifying anything. Alias for -WhatIf.
    .PARAMETER Push
        Automatically push commit and tag to remote Git repository without interactive prompt.
    .PARAMETER NoPush
        Commit and tag locally, but do not push to the remote repository (skips push prompt).
    .PARAMETER NonInteractive
        Run without interactive confirmation prompts. Alias: -Force, -Yes.
#>

#Requires -Version 7.0

[CmdletBinding(SupportsShouldProcess = $true)]
param (
    [Parameter(Position = 0, HelpMessage = "The explicit version or Git tag to create.")]
    [Alias("Version")]
    [string]$TagName = "",

    [Parameter(HelpMessage = "The message for the new Git tag and commit.")]
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

    [Parameter(HelpMessage = "Show current versions without making changes.")]
    [Alias("show", "current")]
    [switch]$ShowCurrentVersion,

    [Parameter(HelpMessage = "Preview planned changes without writing to disk or Git.")]
    [switch]$DryRun,

    [Parameter(HelpMessage = "Automatically push commit and tag to remote Git repository.")]
    [switch]$Push,

    [Parameter(HelpMessage = "Do not push commit or tag to remote Git repository.")]
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

# --- Pure Functions: Version Parsing & Formatting ---

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
    # 4-digit scheme (< 65535 UINT16 ceiling):
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
            throw "Invalid explicit version/tag provided: '$ExplicitTag'. Must follow SemVer (e.g., 'v0.6.0' or '0.6.0-beta1')."
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
        $content = Get-Content -Raw -Path $fullPath
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
            Write-Host -ForegroundColor Yellow "  No changes needed: $($plan.RelPath)"
        }
    }
}

# --- Main Controller Logic ---

Push-Location $ProjectRoot
try {
    # 1. Resolve parameters & normalize switches
    if ($Push -and $NoPush) {
        Write-Host -ForegroundColor Red "Error: Cannot specify both -Push and -NoPush."
        exit 1
    }

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

    # 2. Get current tag & parse current version
    $currentTag = Get-MostRecentTag
    if ($currentTag) {
        $currentObj = Parse-VersionString -Tag $currentTag
    } else {
        $currentObj = Parse-VersionString -Tag "v0.0.0"
    }

    $currentFormatted = Format-BuildVersions $currentObj

    # 3. Handle -ShowCurrentVersion
    if ($ShowCurrentVersion) {
        Write-Host ""
        Write-Host -ForegroundColor Cyan "TextoVista Version Information:"
        if ($currentTag) {
            Write-Host -NoNewline "  Most Recent Git Tag  : "; Write-Host -ForegroundColor Green "$currentTag"
        } else {
            Write-Host -NoNewline "  Most Recent Git Tag  : "; Write-Host -ForegroundColor Yellow "No tags found"
        }
        Write-Host -NoNewline "  Current SemVer       : "; Write-Host -ForegroundColor Green "$($currentFormatted.SemVer)"
        Write-Host -NoNewline "  Current Windows File : "; Write-Host -ForegroundColor Green "$($currentFormatted.FileVersion)"
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

    # 4. Verify working tree is clean (allowing target files and the script itself)
    $uncommitted = git status --porcelain=v1
    if (-not [string]::IsNullOrWhiteSpace($uncommitted)) {
        $dirtyLines = $uncommitted -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
        $otherChanges = @()
        foreach ($line in $dirtyLines) {
            $file = $line.Substring(3).Trim()
            $isTarget = $false
            if ($file -match 'New-GitTag\.ps1$') {
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
            Write-Host "Please commit, stash, or revert these changes before tagging."
            exit 1
        }
    }

    # 5. Verify all target files exist
    if (-not (Test-TargetFilesExist -BaseDir $ProjectRoot)) {
        Write-Host -ForegroundColor Red "Aborting: One or more target files are missing."
        exit 1
    }

    # 6. Calculate target version
    $nextObj = Calculate-NextVersion `
        -CurrentObj $currentObj `
        -ExplicitTag $TagName `
        -RequestedBump $bumpAction `
        -RequestedStage $stage `
        -IsRelease:($ReleaseVersion -eq $true)

    $nextFormatted = Format-BuildVersions $nextObj
    $newTagName = $nextFormatted.TagName
    $newSemVer = $nextFormatted.SemVer
    $newFileVer = $nextFormatted.FileVersion

    # 7. Prepare commit message
    if ([string]::IsNullOrWhiteSpace($Message)) {
        $finalMessage = "Version $newSemVer ($newTagName)"
    } else {
        if ($Message -notmatch [regex]::Escape($newTagName)) {
            $finalMessage = "$Message ($newTagName)"
        } else {
            $finalMessage = $Message
        }
    }

    # 8. Check remote and branch
    $currentBranch = (git branch --show-current 2>$null).Trim()
    if ([string]::IsNullOrWhiteSpace($currentBranch)) {
        $currentBranch = "HEAD (detached)"
    }
    $canPush = (-not $NoPush) -and ($currentBranch -ne "HEAD (detached)")

    # 9. Compute file diff plan
    $diffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $newSemVer -FileVersion $newFileVer

    # 10. Display Planned Operations Summary
    Write-Host ""
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host -ForegroundColor Cyan "             TextoVista Tag & Release Plan                "
    Write-Host -ForegroundColor Cyan "=========================================================="
    Write-Host "  Current Version      : $($currentFormatted.TagName) (SemVer: $($currentFormatted.SemVer), Win: $($currentFormatted.FileVersion))"
    Write-Host -NoNewline "  Target Version       : "; Write-Host -ForegroundColor Green "$newTagName (SemVer: $newSemVer, Win: $newFileVer)"
    Write-Host -NoNewline "  Git Tag Name         : "; Write-Host -ForegroundColor Yellow "$newTagName"
    Write-Host -NoNewline "  Commit/Tag Message   : "; Write-Host -ForegroundColor Yellow "$finalMessage"
    Write-Host -NoNewline "  Git Branch           : "; Write-Host "$currentBranch"
    Write-Host -NoNewline "  Push to Remote       : "
    if ($Push) {
        Write-Host -ForegroundColor Green "Yes (-Push specified: origin/$currentBranch + tag $newTagName)"
    } elseif ($NoPush) {
        Write-Host -ForegroundColor Yellow "No (-NoPush specified: Local only)"
    } elseif ($currentBranch -eq "HEAD (detached)") {
        Write-Host -ForegroundColor Yellow "No (Detached HEAD: Local only)"
    } elseif ($NonInteractive) {
        Write-Host -ForegroundColor Yellow "No (Local only in non-interactive mode unless -Push is specified)"
    } else {
        Write-Host -ForegroundColor Cyan "Prompt after local commit & tag (default: No)"
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
        Write-Host "  git add $relList"
        Write-Host "  git commit -m `"$finalMessage`""
        Write-Host "  git tag -a `"$newTagName`" -m `"$finalMessage`""
        if ($Push) {
            Write-Host "  git push origin $currentBranch"
            Write-Host "  git push origin $newTagName"
        } elseif ($canPush -and -not $NonInteractive) {
            Write-Host "  (Interactive prompt will ask whether to push to origin after commit & tag)"
        }
        Write-Host ""
        return
    }

    # 12. Interactive Confirmation Prompt (if not -NonInteractive)
    if (-not $NonInteractive) {
        Write-Host -NoNewline -ForegroundColor Cyan "Apply these version changes, commit, and create tag '$newTagName'? [Y]es / [N]o / [E]dit message or tag (default: Y): "
        $response = Read-Host
        if ([string]::IsNullOrWhiteSpace($response) -or $response -imatch '^\s*y(es)?\s*$') {
            # User confirmed
        } elseif ($response -imatch '^\s*e(dit)?\s*$') {
            $customTag = Read-Host "  Enter tag name (press ENTER to keep '$newTagName')"
            if (-not [string]::IsNullOrWhiteSpace($customTag)) {
                $newTagName = $customTag
                $parsedCustom = Parse-VersionString -Tag $newTagName
                if ($parsedCustom.IsValid) {
                    $nextFormatted = Format-BuildVersions $parsedCustom
                    $newSemVer = $nextFormatted.SemVer
                    $newFileVer = $nextFormatted.FileVersion
                    $diffPlans = Get-FileDiffPlan -BaseDir $ProjectRoot -SemVer $newSemVer -FileVersion $newFileVer
                }
            }
            $customMsg = Read-Host "  Enter commit/tag message (press ENTER to keep '$finalMessage')"
            if (-not [string]::IsNullOrWhiteSpace($customMsg)) {
                $finalMessage = $customMsg
            }
            Write-Host ""
            Write-Host -ForegroundColor Cyan "Updated Target: $newTagName | Message: $finalMessage"
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

    # 14. Git Stage & Commit
    Write-Host -ForegroundColor Cyan "Staging files in Git..."
    $modifiedFiles = $diffPlans | Where-Object { $_.IsModified } | ForEach-Object { $_.RelPath }
    if ($modifiedFiles.Count -gt 0) {
        git add $modifiedFiles 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to stage files in Git."
            exit 1
        }

        Write-Host -ForegroundColor Cyan "Committing changes: $finalMessage"
        git commit -m "$finalMessage" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Red "Error: Failed to commit changes to Git."
            exit 1
        }
        Write-Host -ForegroundColor Green "  Commit created successfully."
    } else {
        Write-Host -ForegroundColor Yellow "  No file changes to commit."
    }

    # 15. Create Annotated Tag
    Write-Host -ForegroundColor Cyan "Creating Git tag: $newTagName"
    git tag -a "$newTagName" -m "$finalMessage" 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host -ForegroundColor Red "Error: Failed to create Git tag '$newTagName'."
        exit 1
    }
    Write-Host -ForegroundColor Green "  Tag '$newTagName' created locally."

    # 16. Push to Remote (Prompt or Switch)
    $shouldPush = $false

    if ($currentBranch -eq "HEAD (detached)") {
        Write-Host -ForegroundColor Yellow "Skipped remote push (detached HEAD). Commit and tag exist locally."
    } elseif ($NoPush) {
        Write-Host -ForegroundColor Yellow "Skipped remote push (-NoPush specified). Commit and tag exist locally."
    } elseif ($Push) {
        $shouldPush = $true
    } elseif ($NonInteractive) {
        Write-Host -ForegroundColor Yellow "Skipped remote push in non-interactive mode. (Use -Push to push automatically)."
    } else {
        Write-Host ""
        Write-Host -NoNewline -ForegroundColor Cyan "Push branch '$currentBranch' and tag '$newTagName' to remote 'origin'? [y/N]: "
        $pushResponse = Read-Host
        if ($pushResponse -imatch '^\s*y(es)?\s*$') {
            $shouldPush = $true
        }
    }

    if ($shouldPush) {
        Write-Host -ForegroundColor Cyan "Pushing branch '$currentBranch' to remote..."
        git push origin "$currentBranch" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Yellow "Warning: Failed to push branch '$currentBranch' to remote. You may need to push manually."
        } else {
            Write-Host -ForegroundColor Green "  Branch pushed successfully."
        }

        Write-Host -ForegroundColor Cyan "Pushing tag '$newTagName' to remote..."
        git push origin "$newTagName" 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Host -ForegroundColor Yellow "Warning: Failed to push tag '$newTagName' to remote. You may need to push manually."
        } else {
            Write-Host -ForegroundColor Green "  Tag pushed successfully."
        }
    } else {
        if (-not $NoPush -and ($currentBranch -ne "HEAD (detached)") -and -not $NonInteractive) {
            Write-Host -ForegroundColor Yellow "Skipped remote push. Commit and tag exist locally."
        }
        Write-Host ""
        Write-Host "To push to remote manually later, run:"
        Write-Host "  git push origin $currentBranch"
        Write-Host "  git push origin $newTagName"
    }

    Write-Host ""
    Write-Host -ForegroundColor Green "Successfully completed release tag $newTagName!"
    Write-Host ""
}
finally {
    Pop-Location
}
