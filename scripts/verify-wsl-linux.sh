#!/bin/bash

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Verify WSL/Linux (Ubuntu) build prerequisites and run repo validation steps"
    echo ""
    echo "Options:"
    echo "  --skip-go-tests         Skip 'go test ./...'"
    echo "  --skip-frontend         Skip frontend dependency install and build"
    echo "  --skip-wails-build      Skip 'wails build -tags webkit2_41'"
    echo "  -h, --help              Display this help message"
    echo ""
    echo "Examples:"
    echo "  $0"
    echo "  $0 --skip-wails-build"
}

skip_go_tests=0
skip_frontend=0
skip_wails_build=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-go-tests)
            skip_go_tests=1
            shift
            ;;
        --skip-frontend)
            skip_frontend=1
            shift
            ;;
        --skip-wails-build)
            skip_wails_build=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1" >&2
            usage
            exit 1
            ;;
    esac
done

script_full_path="$(readlink -f "${BASH_SOURCE[0]}")"
script_root="$(dirname "$script_full_path")"

if [[ "$script_root" == */scripts ]]; then
    tmp_project_root="$(dirname "$script_root")"
else
    tmp_project_root="$script_root"
fi

if [[ -f "$tmp_project_root/wails.json" ]]; then
    project_root="$tmp_project_root"
else
    echo "Error: Could not find wails.json in the expected project root: $tmp_project_root" >&2
    exit 1
fi

frontend_dir="$project_root/frontend"
go_version="1.27.0"
node_major_version="24"
wails_version="v2.16.0"
changes_made=()

required_apt_packages=(
    build-essential
    ca-certificates
    curl
    pkg-config
    libgtk-3-dev
    libwebkit2gtk-4.1-dev
)

required_commands=(
    go
    node
    npm
    wails
    pkg-config
)

print_section() {
    echo ""
    echo "== $1 =="
}

fail() {
    echo "Error: $1" >&2
    exit 1
}

record_change() {
    changes_made+=("$1")
}

print_change_summary() {
    if [[ ${#changes_made[@]} -eq 0 ]]; then
        return 0
    fi

    print_section "Setup Changes"
    printf 'Upgrade performed: %s\n' "${changes_made[@]}"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

extract_semver() {
    local version_text="$1"

    printf '%s\n' "$version_text" | grep -Eo 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n 1
}

version_is_less_than() {
    local current_version="$1"
    local target_version="$2"

    current_version="${current_version#v}"
    target_version="${target_version#v}"

    [[ "$(printf '%s\n%s\n' "$current_version" "$target_version" | sort -V | head -n 1)" != "$target_version" && "$current_version" != "$target_version" ]]
}

installed_go_version() {
    if ! command_exists go; then
        return 1
    fi

    go version | awk '{print $3}' | sed 's/^go//'
}

installed_node_version() {
    if ! command_exists node; then
        return 1
    fi

    node --version 2>/dev/null | sed 's/^v//'
}

installed_wails_version() {
    if ! command_exists wails; then
        return 1
    fi

    extract_semver "$(wails version 2>/dev/null)"
}

node_target_version_floor() {
    echo "${node_major_version}.0.0"
}

report_current_version() {
    local label="$1"
    local current_version="$2"
    local target_version="$3"

    if [[ -n "$current_version" ]]; then
        echo "$label version detected: $current_version"
    else
        echo "$label version could not be determined"
    fi

    echo "$label target version: $target_version"
}

verify_ubuntu_wsl_distro() {
    local distro_id
    local distro_like
    local distro_name

    if [[ ! -r /etc/os-release ]]; then
        fail "Could not read /etc/os-release to verify the WSL distro"
    fi

    distro_id=""
    distro_like=""
    distro_name=""

    # shellcheck disable=SC1091
    . /etc/os-release

    distro_id="${ID:-}"
    distro_like="${ID_LIKE:-}"
    distro_name="${PRETTY_NAME:-${NAME:-unknown}}"

    if [[ "$distro_id" == "ubuntu" || " $distro_like " == *" ubuntu "* ]]; then
        echo "Ubuntu-based distro detected: $distro_name"
        return 0
    fi

    fail "This script supports Ubuntu-based WSL distros only. Detected: $distro_name"
}

run_privileged() {
    if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
        run_checked "$@"
    elif command_exists sudo; then
        echo "> sudo $*"
        sudo "$@" || fail "Command failed: sudo $*"
    else
        fail "Root privileges are required for: $*"
    fi
}

refresh_shell_path() {
    export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"
    hash -r
}

persist_go_path() {
    local bashrc_path="$HOME/.bashrc"
    local path_line='export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH'

    if [[ -f "$bashrc_path" ]] && grep -Fqx "$path_line" "$bashrc_path"; then
        return 0
    fi

    echo "$path_line" >> "$bashrc_path"
    echo "Added Go and Wails PATH setup to $bashrc_path"
}

go_linux_arch() {
    local detected_arch
    detected_arch="$(dpkg --print-architecture 2>/dev/null || uname -m)"

    case "$detected_arch" in
        amd64|x86_64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            fail "Unsupported architecture for Go bootstrap: $detected_arch"
            ;;
    esac
}

install_go_toolchain() {
    local go_arch
    local go_archive
    local go_url
    local temp_dir
    local action_label="$1"

    go_arch="$(go_linux_arch)"
    go_archive="go${go_version}.linux-${go_arch}.tar.gz"
    go_url="https://go.dev/dl/${go_archive}"
    temp_dir="$(mktemp -d)"

    print_section "Installing Go"
    run_checked curl -fL "$go_url" -o "$temp_dir/$go_archive"
    run_privileged rm -rf /usr/local/go
    run_privileged tar -C /usr/local -xzf "$temp_dir/$go_archive"
    rm -rf "$temp_dir"
    persist_go_path
    refresh_shell_path
    record_change "Go ${action_label} -> ${go_version}"
}

install_nodejs_runtime() {
    local nodesource_url
    local action_label="$1"

    nodesource_url="https://deb.nodesource.com/setup_${node_major_version}.x"

    print_section "Installing Node.js"
    if [[ ${EUID:-$(id -u)} -eq 0 ]]; then
        run_checked bash -lc "curl -fsSL '$nodesource_url' | bash -"
    elif command_exists sudo; then
        run_checked bash -lc "curl -fsSL '$nodesource_url' | sudo -E bash -"
    else
        fail "Root privileges are required to configure the NodeSource repository"
    fi

    run_privileged env DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs
    record_change "Node.js ${action_label} -> ${node_major_version}.x"
}

install_wails_cli() {
    local action_label="$1"

    print_section "Installing Wails CLI"
    refresh_shell_path
    run_checked go install github.com/wailsapp/wails/v2/cmd/wails@"$wails_version"
    refresh_shell_path
    record_change "Wails ${action_label} -> ${wails_version}"
}

ensure_go_available() {
    local current_go_version
    local action_label

    refresh_shell_path

    if command_exists go; then
        current_go_version="$(installed_go_version)"
        report_current_version "Go" "$current_go_version" "$go_version"

        if [[ -n "$current_go_version" ]] && ! version_is_less_than "$current_go_version" "$go_version"; then
            return 0
        fi

        echo "Go is older than required or version detection failed; upgrading"
        action_label="upgraded"
    else
        echo "Go is not installed; installing"
        action_label="installed"
    fi

    install_go_toolchain "$action_label"

    current_go_version="$(installed_go_version)"
    report_current_version "Go" "$current_go_version" "$go_version"

    if [[ -n "$current_go_version" ]] && ! version_is_less_than "$current_go_version" "$go_version"; then
        return 0
    fi

    fail "Go is still not available at the required version after attempted install"
}

ensure_node_available() {
    local current_node_version
    local node_target_floor
    local action_label

    node_target_floor="$(node_target_version_floor)"

    if command_exists node && command_exists npm; then
        current_node_version="$(installed_node_version)"
        report_current_version "Node.js" "$current_node_version" "$node_target_floor"

        if [[ -n "$current_node_version" ]] && ! version_is_less_than "$current_node_version" "$node_target_floor"; then
            return 0
        fi

        echo "Node.js is older than required or version detection failed; upgrading"
        action_label="upgraded"
    else
        echo "Node.js/npm are not installed; installing"
        action_label="installed"
    fi

    install_nodejs_runtime "$action_label"

    if ! command_exists npm; then
        fail "npm is still not available after attempted install"
    fi

    current_node_version="$(installed_node_version)"
    report_current_version "Node.js" "$current_node_version" "$node_target_floor"

    if [[ -n "$current_node_version" ]] && ! version_is_less_than "$current_node_version" "$node_target_floor"; then
        return 0
    fi

    fail "Node.js is still not available at the required version after attempted install"
}

ensure_wails_available() {
    local current_wails_version
    local action_label

    refresh_shell_path

    if command_exists wails; then
        current_wails_version="$(installed_wails_version)"
        report_current_version "Wails" "$current_wails_version" "$wails_version"

        if [[ -n "$current_wails_version" ]] && ! version_is_less_than "$current_wails_version" "$wails_version"; then
            return 0
        fi

        echo "Wails is older than required or version detection failed; upgrading"
        action_label="upgraded"
    else
        echo "Wails is not installed; installing"
        action_label="installed"
    fi

    ensure_go_available
    install_wails_cli "$action_label"

    current_wails_version="$(installed_wails_version)"
    report_current_version "Wails" "$current_wails_version" "$wails_version"

    if [[ -n "$current_wails_version" ]] && ! version_is_less_than "$current_wails_version" "$wails_version"; then
        return 0
    fi

    fail "Wails CLI is still not available at the required version after attempted install"
}

bootstrap_toolchain() {
    ensure_go_available
    ensure_node_available
    ensure_wails_available
}

ensure_apt_get_available() {
    if ! command_exists apt-get; then
        fail "apt-get is required to install missing Ubuntu packages"
    fi
}

check_command() {
    local command_name="$1"

    if ! command_exists "$command_name"; then
        fail "Required command not found: $command_name"
    fi
}

check_dpkg_package() {
    local package_name="$1"

    if ! dpkg -s "$package_name" >/dev/null 2>&1; then
        return 1
    fi

    return 0
}

install_missing_apt_packages() {
    local missing_packages=()
    local package_name

    ensure_apt_get_available

    for package_name in "$@"; do
        if check_dpkg_package "$package_name"; then
            echo "Installed package: $package_name"
        else
            missing_packages+=("$package_name")
        fi
    done

    if [[ ${#missing_packages[@]} -eq 0 ]]; then
        echo "All required Ubuntu packages are already installed"
        return 0
    fi

    echo "Missing Ubuntu packages: ${missing_packages[*]}"

    print_section "Installing Missing Ubuntu Packages"
    run_privileged apt-get update
    run_privileged env DEBIAN_FRONTEND=noninteractive apt-get install -y "${missing_packages[@]}"

    print_section "Retesting Installed Ubuntu Packages"
    for package_name in "${missing_packages[@]}"; do
        if check_dpkg_package "$package_name"; then
            echo "Installed package: $package_name"
            record_change "Ubuntu package installed -> $package_name"
        else
            fail "Required package is still not installed after attempted install: $package_name"
        fi
    done
}

run_checked() {
    echo "> $*"
    "$@" || fail "Command failed: $*"
}

original_dir="$PWD"
cd "$project_root" || fail "Could not change to project root directory: $project_root"
trap 'cd "$original_dir" >/dev/null 2>&1 || true' EXIT
refresh_shell_path

print_section "Environment"
uname -a

if grep -qi microsoft /proc/version 2>/dev/null; then
    echo "WSL environment detected"
    verify_ubuntu_wsl_distro
else
    echo "Warning: This does not appear to be WSL. Continuing anyway."
fi

if [[ "$project_root" == /mnt/* ]]; then
    echo "Warning: Repo is running from a mounted Windows path: $project_root"
    echo "Warning: For better file watching and dev-server performance, prefer a Linux path such as ~/src/markdown-reader-dev"
fi

print_section "Required Ubuntu Packages"
install_missing_apt_packages "${required_apt_packages[@]}"

print_section "Toolchain Bootstrap"
bootstrap_toolchain

print_section "Required Commands"
for command_name in "${required_commands[@]}"; do
    check_command "$command_name"
    echo "Found command: $command_name -> $(command -v "$command_name")"
done

echo "Go version: $(go version)"
echo "Node version: $(node --version 2>/dev/null || echo 'node not found')"
echo "npm version: $(npm --version)"
echo "Wails version: $(wails version)"

print_section "pkg-config Checks"
echo "gtk+-3.0: $(pkg-config --modversion gtk+-3.0)"
echo "webkit2gtk-4.1: $(pkg-config --modversion webkit2gtk-4.1)"

if [[ -f "$frontend_dir/package-lock.json" ]]; then
    echo "frontend/package-lock.json found; npm ci will be used"
else
    echo "frontend/package-lock.json not found; npm install will be used"
fi

if [[ $skip_go_tests -eq 0 ]]; then
    print_section "Go Tests"
    run_checked go test ./...
else
    print_section "Go Tests"
    echo "Skipped"
fi

if [[ $skip_frontend -eq 0 ]]; then
    print_section "Frontend Install"
    cd "$frontend_dir" || fail "Could not change to frontend directory: $frontend_dir"

    if [[ -f package-lock.json ]]; then
        run_checked npm ci
    else
        run_checked npm install
    fi

    print_section "Frontend Build"
    run_checked npm run build
    cd "$project_root" || fail "Could not change back to project root directory: $project_root"
else
    print_section "Frontend Install/Build"
    echo "Skipped"
fi

if [[ $skip_wails_build -eq 0 ]]; then
    print_section "Wails Build"
    run_checked wails build -tags webkit2_41
else
    print_section "Wails Build"
    echo "Skipped"
fi

print_change_summary

print_section "Verification Complete"
echo "WSL/Linux verification completed successfully"