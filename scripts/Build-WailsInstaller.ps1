<#
    .SYNOPSIS
        Builds the Wails application and updates the version in wails.json.
    .DESCRIPTION
        This script builds the Wails application, updates the version in
        wails.json based on the current Git commit, and generates SHA256
        and SHA1 hashes for the installer files.

        To preserve the current state of the repository, it reverts changes to
        the wails.json and project.nsi files after building. It also checks
        if the repository is clean before proceeding with the build. If there are
        uncommitted changes, the script will not proceed and will display an error message.

        By default, it shows the current version and file version (based on the
        latest Git commit) that would be used for the build, but without actually
        performing the build.

        You can use the -Build, -NSIS, and -Zip switches to perform the build,
        create and package an NSIS installer into a ZIP file, or bundle the executable
        into a portable ZIP distribution.
    .PARAMETER Build
        Run the build process without generating the NSIS installer.
    .PARAMETER NSIS
        Implies -Build. Create the NSIS installer after building, compress it into a
        ZIP file, and generate SHA256 and SHA1 hashes for the ZIP.
    .PARAMETER Zip
        Implies -Build. Package the executable into a portable ZIP distribution
        (TextoVista-<version>-amd64-portable.zip) and generate SHA256 and SHA1 hashes.
    .PARAMETER RunAllTests
        Run the full frontend test suite when invoking Run-AllTests.ps1 before building.
    .PARAMETER NoTests
        Bypass Run-AllTests.ps1 before building.
#>

#Requires -Version 7.0

[CmdletBinding()]
param (
    [Parameter(Mandatory = $false, HelpMessage = "Run the build process without generating the NSIS installer.")]
    [Alias("b","exe")]
    [switch]$Build,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -Build) Create the NSIS installer and compress it into a ZIP file with hashes.")]
    [Alias("i","installer")]
    [switch]$NSIS,
    [Parameter(Mandatory = $false, HelpMessage = "(Implies -Build) Create a portable ZIP package containing the executable with hashes.")]
    [Alias("z","portable")]
    [switch]$Zip,
    [Parameter(Mandatory = $false, HelpMessage = "Run the full frontend test suite before building.")]
    [switch]$RunAllTests,
    [Parameter(Mandatory = $false, HelpMessage = "Bypass running tests before building.")]
    [Alias("NoTest")]
    [switch]$NoTests
)

# Set up script and project paths
$ScriptFullName = $MyInvocation.MyCommand.Path
$ScriptRoot = Split-Path -Parent $ScriptFullName
$ScriptName = Split-Path -Leaf $ScriptFullName
if ($ScriptRoot -match '[\\/]scripts[\\/]?$') {
    $tmpProjectRoot = $ScriptRoot -replace '[\\/]scripts[\\/]?', ''
} else {
    $tmpProjectRoot = $ScriptRoot
}
if (Test-Path -Path "$tmpProjectRoot\wails.json") {
    $ProjectRoot = $tmpProjectRoot
} else {
    Write-Host -ForegroundColor Red "Could not find wails.json in the expected project root: $tmpProjectRoot"
    exit 1
}

# Set the default behavior to show the current version and file version
# without performing the build.
$ShowVersionOnly = $true
if ($Build -or $NSIS -or $Zip) {
    $ShowVersionOnly = $false
}

$FileList = @(
    "wails.json",
    "frontend/package.json",
    "frontend/package-lock.json",
    "build/windows/info.json",
    "build/windows/installer/project.nsi"
)

$TestCacheSchemaVersion = 1
$TestCacheDirectory = Join-Path (Join-Path $ProjectRoot "tmp") "test-cache"
$TestCachePath = Join-Path $TestCacheDirectory "build-wailsinstaller-tests.json"
$TestImpactPaths = @(
    ':(glob)**/*.go',
    'go.mod',
    'go.sum',
    'frontend/src/',
    'frontend/tests/',
    'frontend/package.json',
    'frontend/package-lock.json',
    'frontend/playwright.config.ts',
    'frontend/vite.config.ts',
    'frontend/vitest.setup.ts',
    'frontend/index.html',
    'frontend/tsconfig.json',
    'frontend/tsconfig.node.json',
    'wailsjs/',
    'scripts/Run-AllTests.ps1',
    'scripts/Run-FrontendTests.ps1'
)

function Get-JsonContent {
    <#
    .SYNOPSIS
        Reads a JSON file and converts it into a PowerShell object.
    .DESCRIPTION
        This function reads the content of a JSON file and converts it into a PowerShell object.
    .PARAMETER Path
        The path to the JSON file to read.
    #>
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName = $true, Position = 0)]
        [string]$Path
    )

    process {
        if (-not (Test-Path -Path $Path -PathType Leaf)) {
            Write-Error "File not found at path: $Path"
            return
        }
        try {
            # Read the file and convert it from JSON into a PowerShell object.
            # PowerShell automatically creates an ordered object when parsing from JSON text.
            return Get-Content -Raw -Path $Path | ConvertFrom-Json
        }
        catch {
            Write-Error "Failed to read or parse JSON from '$Path'. Error: $_"
        }
    }
}

function Set-JsonContent {
    <#
    .SYNOPSIS
        Writes a PowerShell object to a JSON file.
    .DESCRIPTION
        This function takes a PowerShell object and writes it to a JSON file.
    .PARAMETER Path
        The path to the JSON file to write.
    #>
    [CmdletBinding()]
    param (
        [Parameter(Mandatory = $true, Position = 0)]
        [string]$Path,

        [Parameter(Mandatory = $true, ValueFromPipeline = $true, Position = 1)]
        [psobject]$Value
    )

    process {
        try {
            # Convert the incoming object to a JSON string and write it to the specified file.
            # The -Depth parameter prevents truncation of nested objects.
            $Value | ConvertTo-Json -Depth 100 | Out-File -FilePath $Path -Encoding utf8 -NoNewline
        }
        catch {
            Write-Error "Failed to write JSON to '$Path'. Error: $_"
        }
    }
}

function Get-RequestedFrontendTestSuite {
    if ($RunAllTests) {
        return "all"
    }

    return "fast"
}

function Test-CachedFrontendSuiteSatisfiesRequest {
    param (
        [Parameter(Mandatory = $true)]
        [string]$CachedSuite,
        [Parameter(Mandatory = $true)]
        [string]$RequestedSuite
    )

    if ([string]::IsNullOrWhiteSpace($CachedSuite)) {
        return $false
    }

    if ($RequestedSuite -eq "all") {
        return ($CachedSuite -eq "all")
    }

    return ($CachedSuite -eq "fast" -or $CachedSuite -eq "all")
}

function Test-GitCommitExists {
    param (
        [Parameter(Mandatory = $true)]
        [string]$Commit
    )

    if ([string]::IsNullOrWhiteSpace($Commit)) {
        return $false
    }

    & git rev-parse --verify --quiet "$Commit^{commit}" 2>$null | Out-Null
    return ($LASTEXITCODE -eq 0)
}

function Get-ShortGitCommit {
    param (
        [Parameter(Mandatory = $true)]
        [string]$Commit
    )

    if ([string]::IsNullOrWhiteSpace($Commit)) {
        return $Commit
    }

    $shortCommit = (& git rev-parse --short=8 $Commit 2>$null)
    if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($shortCommit)) {
        return $shortCommit.Trim()
    }

    if ($Commit.Length -le 8) {
        return $Commit
    }

    return $Commit.Substring(0, 8)
}

function Get-TestImpactingChangesSinceCommit {
    param (
        [Parameter(Mandatory = $true)]
        [string]$Commit
    )

    $diffArgs = @("diff", "--name-only", $Commit, "HEAD", "--") + $TestImpactPaths
    $diffOutput = & git @diffArgs 2>$null

    if ($LASTEXITCODE -ne 0) {
        throw "git diff failed while checking test-impacting changes from commit '$Commit'."
    }

    return @($diffOutput | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
}

function Get-TestCacheRecord {
    if (-not (Test-Path -Path $TestCachePath -PathType Leaf)) {
        return $null
    }

    try {
        return Get-Content -Raw -Path $TestCachePath | ConvertFrom-Json
    }
    catch {
        Write-Host -ForegroundColor Yellow "Test cache is unreadable and will be ignored: $TestCachePath"
        return $null
    }
}

function Update-TestCacheRecord {
    param (
        [Parameter(Mandatory = $true)]
        [string]$TestedCommit,
        [Parameter(Mandatory = $true)]
        [string]$FrontendSuite
    )

    if (-not (Test-Path -Path $TestCacheDirectory -PathType Container)) {
        New-Item -ItemType Directory -Path $TestCacheDirectory -Force | Out-Null
    }

    $cacheRecord = [pscustomobject][ordered]@{
        SchemaVersion = $TestCacheSchemaVersion
        Runner = "Run-AllTests.ps1"
        Platform = "windows"
        FrontendSuite = $FrontendSuite
        TestedCommit = $TestedCommit
        TestedAtUtc = (Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
        Success = $true
    }

    try {
        Set-JsonContent -Path $TestCachePath -Value $cacheRecord
    }
    catch {
        Write-Host -ForegroundColor Yellow "Failed to update test cache at '$TestCachePath'."
    }
}

function Get-TestExecutionPlan {
    param (
        [Parameter(Mandatory = $true)]
        [string]$RequestedSuite
    )

    $currentCommit = (git rev-parse HEAD 2>$null).Trim()
    if ([string]::IsNullOrWhiteSpace($currentCommit)) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Could not determine the current Git commit. Running tests before build."
        }
    }

    $cacheRecord = Get-TestCacheRecord
    if ($null -eq $cacheRecord) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "No successful test cache was found. Running tests before build."
        }
    }

    if (($cacheRecord.SchemaVersion -as [int]) -ne $TestCacheSchemaVersion) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Test cache schema version changed. Running tests before build."
        }
    }

    if (-not $cacheRecord.Success) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Last cached test run was not successful. Running tests before build."
        }
    }

    if ($cacheRecord.Runner -ne "Run-AllTests.ps1" -or $cacheRecord.Platform -ne "windows") {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Test cache was created by a different runner or platform. Running tests before build."
        }
    }

    if (-not (Test-CachedFrontendSuiteSatisfiesRequest -CachedSuite $cacheRecord.FrontendSuite -RequestedSuite $RequestedSuite)) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Cached frontend test scope does not satisfy the current build request. Running tests before build."
        }
    }

    if (-not (Test-GitCommitExists -Commit $cacheRecord.TestedCommit)) {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Cached test commit is no longer available locally. Running tests before build."
        }
    }

    $shortCurrentCommit = Get-ShortGitCommit -Commit $currentCommit
    $shortTestedCommit = Get-ShortGitCommit -Commit $cacheRecord.TestedCommit

    if ($cacheRecord.TestedCommit -eq $currentCommit) {
        return [pscustomobject]@{
            ShouldRun = $false
            Reason = "Skipping tests: commit $shortCurrentCommit already has a successful cached test run."
        }
    }

    try {
        $changedFiles = Get-TestImpactingChangesSinceCommit -Commit $cacheRecord.TestedCommit
    }
    catch {
        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Could not compare test-impacting changes since the cached test run. Running tests before build."
        }
    }

    if ($changedFiles.Count -gt 0) {
        $preview = @($changedFiles | Select-Object -First 5)
        $previewText = $preview -join ", "
        if ($changedFiles.Count -gt $preview.Count) {
            $previewText += ", ..."
        }

        return [pscustomobject]@{
            ShouldRun = $true
            Reason = "Test-impacting changes were detected since commit ${shortTestedCommit}: $previewText"
        }
    }

    return [pscustomobject]@{
        ShouldRun = $false
        Reason = "Skipping tests: no test-impacting changes were detected since successful tests on commit $shortTestedCommit."
    }
}

function Get-DateStamp {
    <#
    .SYNOPSIS
        Generates a timestamp based on the current date and time.
    .DESCRIPTION
        This function creates a timestamp by calculating the number of 10-minute intervals
        that have elapsed since the beginning of the year.
    .OUTPUTS
        Returns an integer representing the number of 10-minute intervals since the start of the year.
        Integer value will be in the range of 1 to approximately 52560 (for a full year).
    #>
    $DateNow = Get-Date -AsUTC
    $TicksPerDay = 24 * 6   # number of 10 minute intervals in a day
    $DayOfYear = $DateNow.DayOfYear - 1
    $HourTicksToday = $DateNow.Hour * 6
    $TicksThisHour = [int]($DateNow.Minute / 10)
    $result = ($DayOfYear * $TicksPerDay) + $HourTicksToday + $TicksThisHour
    return [int]$result
}

function Get-VersionHash {
    <#
    .SYNOPSIS
        Retrieves the version hash from a Git tag name structured as a semantic version.
    .DESCRIPTION
        This function extracts the version information from the specified Git tag name.
        It assumes the tag follows the semantic versioning format:
        vMAJOR.MINOR.PATCH[-PRERELEASE][+AHEAD-HASH].
    .PARAMETER TagName
        The name of the Git tag to parse for version information.
    #>
    param (
        [Parameter(Mandatory = $true, HelpMessage = "The tag name to parse for the version hash.")]
        [string]$TagName
    )

    $thisVersionHash = [ordered]@{
        Major = $null
        Minor = $null
        Patch = $null
        Prerelease = $null
        Ahead = $null
        Hash = $null
        IsValid = $false
    }

    if (-not $TagName) {
        Write-Host "No tag name provided. Please specify a tag name."
        return [pscustomobject]$thisVersionHash
    }

    $baseTag = $TagName
    $ahead = $null
    $hash = $null

    if ($TagName -match '^(?<base>.+?)(?:-(?<ahead>\d+)-g(?<hash>[0-9a-fA-F]+))?$') {
        $baseTag = $Matches.base
        $ahead = $Matches.ahead
        $hash = $Matches.hash
    }

    if ($baseTag -match '^v?(?<major>\d+)\.(?<minor>\d+)\.(?<patch>\d+)(?:-(?<prerelease>[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*))?$') {
        $thisVersionHash["Major"] = $Matches.major
        $thisVersionHash["Minor"] = $Matches.minor
        $thisVersionHash["Patch"] = $Matches.patch
        $thisVersionHash["Prerelease"] = $Matches.prerelease
        $thisVersionHash["Ahead"] = $ahead
        $thisVersionHash["Hash"] = $hash
        $thisVersionHash["IsValid"] = $true
    }

    return [pscustomobject]$thisVersionHash
}

function Get-BuildVersionInfo {
    <#
    .SYNOPSIS
        Converts a parsed Git tag into semantic and numeric build versions.
    .DESCRIPTION
        This function maps release and prerelease tags to the semantic version used by the app
        and to the four-part numeric file version required by NSIS.
    .PARAMETER VersionHash
        The parsed version object returned by Get-VersionHash.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [psobject]$VersionHash
    )

    $version = "$($VersionHash.Major).$($VersionHash.Minor).$($VersionHash.Patch)"
    $fileVersionSuffix = 0
    $releaseClasses = @("alpha", "beta", "rc", "patch")

    if (-not [string]::IsNullOrWhiteSpace($VersionHash.Prerelease)) {
        $version += "-$($VersionHash.Prerelease)"
        if ($VersionHash.Prerelease -match '^(?<stage>alpha|beta|rc|patch)(?<number>\d+)?$') {
            $stageName = $Matches.stage
            $stageNumber = if ([string]::IsNullOrWhiteSpace($Matches.number)) { 0 } else { [int]$Matches.number }
            $stageIndex = $releaseClasses.IndexOf($stageName) + 1
            $fileVersionSuffix = ($stageIndex * 10000) + ($stageNumber * 100)
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($VersionHash.Ahead)) {
        $version += "+$($VersionHash.Ahead)"
        $fileVersionSuffix += [int]$VersionHash.Ahead
    }

    return [pscustomobject]@{
        Version = $version
        FileVersion = "$($VersionHash.Major).$($VersionHash.Minor).$($VersionHash.Patch).$fileVersionSuffix"
    }
}

function Get-MostRecentTag {
    <#
    .SYNOPSIS
        Retrieves the most recent Git tag from the repository.
    .DESCRIPTION
        This function uses Git commands to find and return the most recent tag
        in the current repository.
    #>
    $currentTag = git describe --tags --abbrev=0 2>$null
    if ($currentTag) {
        # Write-Host "Most recent tag: $currentTag"
        return $currentTag
    } else {
        Write-Host -ForegroundColor Yellow "No tags found in the repository."
        return $null
    }
}

function Confirm-RepositoryIsClean {
    <#
    .SYNOPSIS
        Checks if the Git repository is clean (no uncommitted changes).
    .DESCRIPTION
        This function checks the status of the Git repository and determines if there are any
        uncommitted changes. It returns $true if the repository is clean, and $false otherwise.
    #>
    param(
        [Parameter(Mandatory = $false)]
        [Alias("i","ignore")]
        [switch]$IgnoreFileList,
        [Parameter(Mandatory = $false)]
        [Alias("q","s","silent")]
        [switch]$Quiet
    )
    $status = git status --porcelain=v1
    if ($status) {
        # $isClean = $true
        $statusList = $status -split "`n"
        $tmpStatusList = @()
        foreach ($line in $statusList) {
            if ( -not ($line -match $ScriptName) ) {
                # $isClean = $false
                $tmpStatusList += $line
            }
        }

        $newStatusList = @()
        if ($IgnoreFileList) {
            foreach ($line in $tmpStatusList) {
                $found = $false
                foreach ($file in $FileList) {
                    $fileRE = [regex]::Escape($file)
                    # Write-StdErr "Checking line: $line ($fileRE)"
                    if ($line -match "^\s*[A-Z]\s+$fileRE") {
                        $found = $true
                        # git add "$file" 2>&1 | Out-Null
                        break
                    }
                }
                if (-not $found) {
                    $newStatusList += $line
                }
            }
        } else {
            $newStatusList += $tmpStatusList
        }

        if ($newStatusList.Count -eq 0) { return $true }

        if (-not $Quiet) {
            Write-Host ""
            Write-Host -ForegroundColor Red "WARNING: Repository is not clean. Please commit or stash your changes before building."
            Write-Host ""
            Write-Host "Uncommitted changes:"
            Write-Host ""
            $newStatusList | ForEach-Object { Write-Host -ForegroundColor Yellow "  $_" }
            Write-Host ""
            Write-Host "Suggestions:"
            Write-Host "  - Commit your changes: git commit -m 'Your commit message'"
            Write-Host "  - Create a new branch: git checkout -b new-branch-name "
            Write-Host "       and commit your changes to the branch"
            Write-Host "  - Stash your changes: git stash --all"
            Write-Host "  - Discard your changes: git reset --hard HEAD"
            Write-Host ""
            Write-Host -ForegroundColor Cyan "NOTE: Script ignores changes to the script itself ($ScriptName)"
            Write-Host ""
        }
        return $false
    }
    return $true
}

function Restore-RepositoryToCleanState {
    <#
    .SYNOPSIS
        Restores the Git repository to a clean state by discarding uncommitted changes.
    .DESCRIPTION
        This function checks for uncommitted changes in the specified files and restores them
        to their last committed state.

        As written, it will only restore changes to the wails.json and project.nsi files.
        If you want to restore other files, add them to the $FileList array.
    #>
    Push-Location $ProjectRoot -StackName "restoreproject"
    Write-Host -ForegroundColor Yellow "Restoring repository to a clean state..."
    foreach ($file in $FileList) {
        if ( (git status --porcelain=v1 | Select-String "$file") ) {
            Write-Host -NoNewLine -ForegroundColor Green "  Restoring: " ; Write-Host -ForegroundColor Yellow "$file"
            git restore "$file" 2>&1 | Out-Null
        }
    }
    Pop-Location -StackName "restoreproject"
}

function Update-ProjectNSI {
    <#
    .SYNOPSIS
        Updates the NSIS project file with the specified file version.
    .DESCRIPTION
        This function modifies the NSIS project file to set the correct file version.
    .PARAMETER ProjectNSIPath
        The path to the NSIS project file to update.
    .PARAMETER FileVersion
        The new file version to set in the project file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$ProjectNSIPath,
        [Parameter(Mandatory = $true)]
        [string]$FileVersion
    )

    if (-not (Test-Path -Path $ProjectNSIPath)) {
        Write-Host -ForegroundColor Red "Project NSI file not found at path: $ProjectNSIPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating NSI project file: $ProjectNSIPath"

    $NSIData = Get-Content -Path $ProjectNSIPath
    $NewNSIData = @()
    $FileChanged = $false

    foreach ($line in $NSIData) {
        if ($line -match '^(?<key>VIFileVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                $newval1 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIFileVersion\s+)"([^"]+)"\s*$', $newval1
                Write-Host -ForegroundColor Green "  Updating VIFileVersion to: $FileVersion"
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIProductVersion\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIProductVersion to: $FileVersion"
                $newval2 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIProductVersion\s+)"([^"]+)"\s*$', $newval2
                $FileChanged = $true
            }
        }
        if ($line -match '^(?<key>VIAddVersionKey\s+"FileVersion"\s+)"(?<value>[^"]+)"\s*$') {
            if ($Matches.value -ne $FileVersion) {
                Write-Host -ForegroundColor Green "  Updating VIAddVersionKey FileVersion to: $FileVersion"
                $newval3 = $Matches.key + '"' + $FileVersion + '"'
                $line = $line -replace '^(VIAddVersionKey\s+"FileVersion"\s+)"([^"]+)"\s*$', $newval3
                $FileChanged = $true
            }
        }
        $NewNSIData += $line
    }
    if (-not $FileChanged) {
        Write-Host -ForegroundColor Yellow "  No changes made to NSIS project file"
        return
    }
    Set-Content -Path $ProjectNSIPath -Encoding utf8 -Value $NewNSIData
}

function Update-WailsJSON {
    <#
    .SYNOPSIS
        Updates the wails.json file with the specified version.
    .DESCRIPTION
        This function modifies the wails.json file to set the correct version.
    .PARAMETER WailsJsonPath
        The path to the wails.json file to update.
    .PARAMETER Version
        The new version to set in the wails.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$WailsJsonPath,
        [Parameter(Mandatory = $true)]
        [string]$Version
    )

    if (-not (Test-Path -Path $WailsJsonPath)) {
        Write-Host -ForegroundColor Red "wails.json file not found at path: $WailsJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating wails.json with version value: $Version"
    $WailsData = Get-JsonContent -Path $WailsJsonPath

    if (-not $WailsData) {
        Write-Host -ForegroundColor Red "  Failed to read wails.json or it is empty."
        return
    }
    if ($WailsData.Info.productVersion -ne $Version) {
        $WailsData.Info.productVersion = $Version
        Set-JsonContent -Path $WailsJsonPath -Value $WailsData
        Write-Host -ForegroundColor Green "  Updating productVersion to: $Version"
    } else {
        Write-Host -ForegroundColor Yellow "  No changes made to wails.json, version is already set to: $Version"
    }
}

function Update-PackageJSON {
    <#
    .SYNOPSIS
        Updates the package.json file with the specified version.
    .DESCRIPTION
        This function modifies the package.json file to set the correct version.
    .PARAMETER PackageJsonPath
        The path to the package.json file to update.
    .PARAMETER Version
        The new version to set in the package.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$PackageJsonPath,
        [Parameter(Mandatory = $true)]
        [string]$Version
    )

    if (-not (Test-Path -Path $PackageJsonPath)) {
        Write-Host -ForegroundColor Red "package.json file not found at path: $PackageJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating package.json with version value: $Version"
    $PackageData = Get-JsonContent -Path $PackageJsonPath

    if (-not $PackageData) {
        Write-Host -ForegroundColor Red "  Failed to read package.json or it is empty."
        return
    }
    if ($PackageData.version -ne $Version) {
        $PackageData.version = $Version
        Set-JsonContent -Path $PackageJsonPath -Value $PackageData
        Write-Host -ForegroundColor Green "  Updating version to: $Version"
    } else {
        Write-Host -ForegroundColor Yellow "  No changes made to package.json, version is already set to: $Version"
    }
}

function Update-InfoJSON {
    <#
    .SYNOPSIS
        Updates the info.json file with the specified version.
    .DESCRIPTION
        This function modifies the info.json file to set the correct version.
    .PARAMETER InfoJsonPath
        The path to the info.json file to update.
    .PARAMETER Version
        The new version to set in the info.json file.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$InfoJsonPath,
        [Parameter(Mandatory = $true)]
        [string]$FileVersion
    )

    if (-not (Test-Path -Path $InfoJsonPath)) {
        Write-Host -ForegroundColor Red "info.json file not found at path: $InfoJsonPath"
        return
    }

    Write-Host -ForegroundColor Cyan "Updating info.json:"
    $VIData = Get-JsonContent -Path $InfoJsonPath

    if (-not $VIData) {
        Write-Host -ForegroundColor Red "  Failed to read info.json or it is empty."
        return
    }
    $VIData.fixed.file_version = $FileVersion
    if (Set-JsonContent -Path $InfoJsonPath -Value $VIData) {
        Write-Host -ForegroundColor Green "  Updated info.json"
    }
}

function New-FileHashes {
    param(
        [Parameter(Mandatory = $false, HelpMessage = "The directory path to create file hashes in.")]
        [string]$DirectoryPath = "./build/bin",
        [Parameter(Mandatory = $false, HelpMessage = "The file pattern to match.")]
        [string]$FilePattern = "*.exe"
    )
    Push-Location $DirectoryPath -StackName "NewFileHashes"
    Write-Host -ForegroundColor Cyan "Writing sha256 and sha1 hashes for executable files..."
    foreach ($file in Get-ChildItem -Path $DirectoryPath -Filter $FilePattern -File -ErrorAction SilentlyContinue) {
        $sha256Name = $file.Name + ".sha256"
        $sha1Name = $file.Name + ".sha1"
        $sha256Hash = (Get-FileHash $file -Algorithm SHA256).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha256Name -Encoding utf8
        $sha1Hash = (Get-FileHash $file -Algorithm SHA1).Hash # | ForEach-Object { $_.Hash } | Out-File -FilePath $sha1Name -Encoding utf8
        "$sha256Hash  $($file.Name)" | Out-File -FilePath $sha256Name -Encoding utf8
        "$sha1Hash  $($file.Name)" | Out-File -FilePath $sha1Name -Encoding utf8
    }
    Pop-Location -StackName "NewFileHashes"
}

function Invoke-AllTests {
    $allTestsScript = Join-Path $ScriptRoot "Run-AllTests.ps1"
    $powerShellExe = Join-Path $PSHOME "pwsh.exe"
    $requestedFrontendSuite = Get-RequestedFrontendTestSuite

    if (-not (Test-Path -Path $allTestsScript -PathType Leaf)) {
        Write-Host -ForegroundColor Red "Could not find test runner script: $allTestsScript"
        return $false
    }

    if (-not (Test-Path -Path $powerShellExe -PathType Leaf)) {
        Write-Host -ForegroundColor Red "Could not find pwsh.exe under PSHOME: $powerShellExe"
        return $false
    }

    $allTestArgs = @("-Silent")

    if ($RunAllTests) {
        Write-Host -ForegroundColor Yellow "Running ALL tests (please wait)..."
        $allTestArgs += "-RunAllTests"
    } else {
        Write-Host -ForegroundColor Yellow "Running tests (please wait)..."
    }

    & $powerShellExe -NoProfile -File $allTestsScript @allTestArgs
    $testExitCode = $LASTEXITCODE

    if ($testExitCode -eq 0) {
        $testedCommit = (git rev-parse HEAD 2>$null).Trim()
        if (-not [string]::IsNullOrWhiteSpace($testedCommit)) {
            Update-TestCacheRecord -TestedCommit $testedCommit -FrontendSuite $requestedFrontendSuite
        }
        Write-Host -ForegroundColor Green "  --[ALL TESTS PASSED]--"
        return $true
    }

    Write-Host -ForegroundColor Red "  **[FAILED]**"
    return $false
}

function Test-ExecutableIsUpToDate {
    <#
    .SYNOPSIS
        Checks if the compiled executable is newer than all relevant project source files.
    .DESCRIPTION
        Compares the LastWriteTimeUtc of the specified executable against all Go, frontend,
        and build configuration files. Returns $true if the binary exists and no sources
        have been modified since the binary was compiled.
    .PARAMETER ExecutablePath
        The path to the executable to inspect.
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$ExecutablePath
    )

    if (-not (Test-Path -Path $ExecutablePath -PathType Leaf)) {
        return $false
    }

    $exeTime = (Get-Item -Path $ExecutablePath).LastWriteTimeUtc

    $sourceScanPaths = @(
        "app.go",
        "main.go",
        "go.mod",
        "go.sum",
        "wails.json",
        "frontend/src",
        "frontend/index.html",
        "frontend/package.json",
        "frontend/vite.config.ts"
    )

    foreach ($relPath in $sourceScanPaths) {
        $fullPath = Join-Path $ProjectRoot $relPath
        if (-not (Test-Path -Path $fullPath)) {
            continue
        }

        if (Test-Path -Path $fullPath -PathType Leaf) {
            if ((Get-Item -Path $fullPath).LastWriteTimeUtc -gt $exeTime) {
                return $false
            }
        } else {
            $items = Get-ChildItem -Path $fullPath -Recurse -File -ErrorAction SilentlyContinue
            foreach ($item in $items) {
                if ($item.LastWriteTimeUtc -gt $exeTime) {
                    return $false
                }
            }
        }
    }

    return $true
}

function Invoke-WailsBuild {
    <#
    .SYNOPSIS
        Invokes the Wails build process with the specified options.
    .DESCRIPTION
        This function prepares and executes the Wails build command with the appropriate
        flags and environment variables.
    #>
    $NSISOption = ""
    $ProjectNSI = Join-Path $ProjectRoot "build" "windows" "installer" "project.nsi"
    $WailsJsonPath = Join-Path $ProjectRoot "wails.json"
    $PackageJsonPath = Join-Path $ProjectRoot "frontend" "package.json"
    $InfoJsonPath = Join-Path $ProjectRoot "build" "windows" "info.json"
    $binDir = Join-Path $ProjectRoot "build" "bin"
    $exePath = Join-Path $binDir "texto-vista.exe"

    if ($NSIS) {
        $NSISOption = "-nsis"
        $Build = $true
    }
    if ($Zip) {
        $Build = $true
    }

    $Date = $(Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")
    $Version = ""
    $FileVersion = ""
    $Commit = $(git rev-parse --short HEAD)
    $CurrentCommitTag = $(git describe --tags HEAD 2> $null)
    $CurrentRepoTag = Get-MostRecentTag
    $CurrentCommitVersion = Get-VersionHash -TagName $CurrentCommitTag
    $CurrentRepoVersion = Get-VersionHash -TagName $CurrentRepoTag

    if ($CurrentCommitVersion.IsValid) {
        $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $CurrentCommitVersion
        $Version = $BuildVersionInfo.Version
        $FileVersion = $BuildVersionInfo.FileVersion
    } elseif ($CurrentRepoVersion.IsValid) {
        $BuildVersionInfo = Get-BuildVersionInfo -VersionHash $CurrentRepoVersion
        $Version = $BuildVersionInfo.Version
        $FileVersion = $BuildVersionInfo.FileVersion
    } else {
        $Version = "0.0.0-dev+${Commit}"
        $ds = Get-DateStamp
        $FileVersion = "0.0.0.${ds}"
    }

    $isRepoClean = Confirm-RepositoryIsClean -Quiet -IgnoreFileList

    if ($ShowVersionOnly) {
        $Build = $false
        Write-Host -ForegroundColor Cyan "`nCurrent Repository Information:"
        Write-Host -NoNewline -ForegroundColor Yellow "  Most Recent Repo Tag: " ; Write-Host -ForegroundColor Green "$CurrentRepoTag"
        Write-Host -NoNewline -ForegroundColor Yellow "   Current Repo Status: "
        if ($isRepoClean) {
            Write-Host -ForegroundColor Green "Clean (Safe to build)"
        } else {
            Write-Host -ForegroundColor Red "Dirty (Uncommitted changes found)"
        }
        Write-Host -ForegroundColor Cyan "`nVersion Values For Next Build:"
        Write-Host -NoNewLine -ForegroundColor Yellow "      Semantic Version: " ; Write-Host -ForegroundColor Green "$Version"
        Write-Host -NoNewLine -ForegroundColor Yellow "       Numeric Version: " ; Write-Host -ForegroundColor Green "$FileVersion"
        return
    }

    if (-not (Confirm-RepositoryIsClean -IgnoreFileList)) {
        return
    }

    if ($Build) {
        # Check if the existing binary is up to date when only packaging a portable ZIP
        $skipCompilation = $false
        if ($Zip -and (-not $NSIS)) {
            if (Test-ExecutableIsUpToDate -ExecutablePath $exePath) {
                $skipCompilation = $true
                Write-Host -ForegroundColor Green "Executable 'texto-vista.exe' is already up to date with sources. Skipping Wails compilation."
            }
        }

        if (-not $skipCompilation) {
            if (-not $NoTests) {
                $testExecutionPlan = Get-TestExecutionPlan -RequestedSuite (Get-RequestedFrontendTestSuite)
                if ($testExecutionPlan.ShouldRun) {
                    Write-Host -ForegroundColor Yellow $testExecutionPlan.Reason
                    if (-not (Invoke-AllTests)) {
                        return
                    }
                } else {
                    Write-Host -ForegroundColor Cyan $testExecutionPlan.Reason
                }
            }

            if ($NSIS) {
                Update-ProjectNSI -ProjectNSIPath $ProjectNSI -FileVersion $FileVersion
            }
            Update-WailsJSON -WailsJsonPath $WailsJsonPath -Version $Version
            Update-PackageJSON -PackageJsonPath $PackageJsonPath -Version $Version
            Update-InfoJSON -InfoJsonPath $InfoJsonPath -FileVersion $FileVersion

            Write-Host -NoNewLine -ForegroundColor Cyan "Building Wails application with version value: "
            Write-Host -NoNewLine -ForegroundColor Black -BackgroundColor White "$Version"
            Write-Host "`n`n$('=' * 78)"
            wails build -clean -ldflags "-X main.Version=${Version} -X main.Date=${Date} -X main.Commit=${Commit} -s -w" ${NSISOption}
            Write-Host "$('=' * 78)`n"

            # Create SHA1 and SHA256 hashes for the executable files
            New-FileHashes -DirectoryPath $binDir -FilePattern "*.exe"

            Restore-RepositoryToCleanState
        }

        if ($NSIS) {
            $installer = Get-ChildItem -Path $binDir -Filter "*-installer.exe" | Sort-Object LastWriteTime -Descending | Select-Object -First 1
            if (-not $installer) {
                Write-Host -ForegroundColor Red "No NSIS installer executable found in $binDir."
            } else {
                $zipPath = Join-Path $installer.DirectoryName ($installer.BaseName + ".zip")
                Write-Host -ForegroundColor Cyan "Compressing NSIS installer into ZIP: $zipPath"
                Compress-Archive -Path $installer.FullName -DestinationPath $zipPath -Force
                Write-Host -ForegroundColor Cyan "Writing sha256 and sha1 hashes for NSIS ZIP file..."
                $sha256Hash = (Get-FileHash $zipPath -Algorithm SHA256).Hash
                $sha1Hash = (Get-FileHash $zipPath -Algorithm SHA1).Hash
                $zipName = Split-Path -Leaf $zipPath
                "$sha256Hash  $zipName" | Out-File -FilePath ($zipPath + ".sha256") -Encoding utf8
                "$sha1Hash  $zipName" | Out-File -FilePath ($zipPath + ".sha1") -Encoding utf8
            }
        }

        if ($Zip) {
            if (-not (Test-Path -Path $exePath -PathType Leaf)) {
                Write-Host -ForegroundColor Red "No executable found to package in $exePath."
            } else {
                $portableZipName = "TextoVista-${Version}-amd64-portable.zip"
                $portableZipPath = Join-Path $binDir $portableZipName
                Write-Host -ForegroundColor Cyan "Compressing portable executable into ZIP: $portableZipPath"
                Compress-Archive -Path $exePath -DestinationPath $portableZipPath -Force
                Write-Host -ForegroundColor Cyan "Writing sha256 and sha1 hashes for portable ZIP file..."
                $sha256Hash = (Get-FileHash $portableZipPath -Algorithm SHA256).Hash
                $sha1Hash = (Get-FileHash $portableZipPath -Algorithm SHA1).Hash
                "$sha256Hash  $portableZipName" | Out-File -FilePath ($portableZipPath + ".sha256") -Encoding utf8
                "$sha1Hash  $portableZipName" | Out-File -FilePath ($portableZipPath + ".sha1") -Encoding utf8
            }
        }
    }
}

Push-Location $ProjectRoot -StackName "ProjectRoot"
Invoke-WailsBuild
Pop-Location -StackName "ProjectRoot"