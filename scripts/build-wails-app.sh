#!/usr/bin/env bash

set -euo pipefail

usage() {
    cat <<'EOF'
Usage: build-wails-app.sh [OPTIONS]

Build the Linux TextoVista artifacts from inside WSL/Ubuntu.

Options:
  -b, --build        Build the Linux executable only
  -t, --tar          Implies --build and creates a tar.gz package
  -d, --deb          Implies --build and creates a .deb package (unimplemented)
  --keep-stage-dir   Keep the temporary tar/deb packaging stage directory
                       after a successful build
  -h, --help         Show this help text

With no options, the script reports repository cleanliness and the next build version.
EOF
}

SCRIPT_FULL_PATH="$(readlink -f "${BASH_SOURCE[0]}")"
SCRIPT_ROOT="$(dirname "$SCRIPT_FULL_PATH")"
SCRIPT_NAME="$(basename "$SCRIPT_FULL_PATH")"

if [[ "$SCRIPT_ROOT" == */scripts ]]; then
    TMP_PROJECT_ROOT="$(dirname "$SCRIPT_ROOT")"
else
    TMP_PROJECT_ROOT="$SCRIPT_ROOT"
fi

if [[ -f "$TMP_PROJECT_ROOT/wails.json" ]]; then
    PROJECT_ROOT="$TMP_PROJECT_ROOT"
else
    echo "Error: Could not find wails.json in the expected project root: $TMP_PROJECT_ROOT" >&2
    exit 1
fi

MANAGED_FILES=(
    "wails.json"
    "frontend/package.json"
    "frontend/package-lock.json"
)

SHOW_VERSION_ONLY=1
DO_BUILD=0
DO_TAR=0
DO_DEB=0
KEEP_STAGE_DIR=0
RESTORE_NEEDED=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        -b|--build|-build)
            DO_BUILD=1
            SHOW_VERSION_ONLY=0
            ;;
        -t|--tar|-tar)
            DO_TAR=1
            DO_BUILD=1
            SHOW_VERSION_ONLY=0
            ;;
        -d|--deb|-deb)
            DO_DEB=1
            DO_BUILD=1
            SHOW_VERSION_ONLY=0
            ;;
        --keep-stage-dir)
            KEEP_STAGE_DIR=1
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Error: Unknown option $1" >&2
            usage >&2
            exit 1
            ;;
    esac
    shift
done

DATE_STAMP=""
PARSED_MAJOR=""
PARSED_MINOR=""
PARSED_PATCH=""
PARSED_PRERELEASE=""
PARSED_AHEAD=""
PARSED_HASH=""
PARSED_IS_VALID=0
BUILD_VERSION=""
BUILD_FILE_VERSION=""
CURRENT_REPO_TAG=""
CURRENT_COMMIT_TAG=""
CURRENT_COMMIT=""

cleanup() {
    if [[ "$RESTORE_NEEDED" -eq 1 ]]; then
        restore_repository_to_clean_state || true
    fi
}

trap cleanup EXIT

require_command() {
    local command_name="$1"
    if ! command -v "$command_name" >/dev/null 2>&1; then
        echo "Error: Required command not found: $command_name" >&2
        exit 1
    fi
}

get_date_stamp() {
    local day_of_year hour minute ticks_per_day hour_ticks minute_ticks
    day_of_year="$(date -u +%j)"
    hour="$(date -u +%H)"
    minute="$(date -u +%M)"
    ticks_per_day=$((24 * 6))
    hour_ticks=$((10#${hour} * 6))
    minute_ticks=$((10#${minute} / 10))
    DATE_STAMP=$((((10#${day_of_year} - 1) * ticks_per_day) + hour_ticks + minute_ticks))
}

get_most_recent_tag() {
    git describe --tags --abbrev=0 2>/dev/null || true
}

parse_version_hash() {
    local tag_name="$1"
    local base_tag ahead hash

    PARSED_MAJOR=""
    PARSED_MINOR=""
    PARSED_PATCH=""
    PARSED_PRERELEASE=""
    PARSED_AHEAD=""
    PARSED_HASH=""
    PARSED_IS_VALID=0

    if [[ -z "$tag_name" ]]; then
        return 1
    fi

    base_tag="$tag_name"
    ahead=""
    hash=""

    if [[ "$tag_name" =~ ^(.+)-([0-9]+)-g([0-9a-fA-F]+)$ ]]; then
        base_tag="${BASH_REMATCH[1]}"
        ahead="${BASH_REMATCH[2]}"
        hash="${BASH_REMATCH[3]}"
    fi

    if [[ "$base_tag" =~ ^v?([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9A-Za-z]+([.-][0-9A-Za-z]+)*))?$ ]]; then
        PARSED_MAJOR="${BASH_REMATCH[1]}"
        PARSED_MINOR="${BASH_REMATCH[2]}"
        PARSED_PATCH="${BASH_REMATCH[3]}"
        PARSED_PRERELEASE="${BASH_REMATCH[5]:-}"
        PARSED_AHEAD="$ahead"
        PARSED_HASH="$hash"
        PARSED_IS_VALID=1
        return 0
    fi

    return 1
}

get_build_version_info() {
    local version file_version_suffix stage_name stage_number stage_index
    local release_classes=(alpha beta rc patch)

    version="${PARSED_MAJOR}.${PARSED_MINOR}.${PARSED_PATCH}"
    file_version_suffix=0

    if [[ -n "$PARSED_PRERELEASE" ]]; then
        version+="-${PARSED_PRERELEASE}"
        if [[ "$PARSED_PRERELEASE" =~ ^(alpha|beta|rc|patch)([0-9]+)?$ ]]; then
            stage_name="${BASH_REMATCH[1]}"
            stage_number="${BASH_REMATCH[2]:-0}"
            stage_index=0

            local index
            for index in "${!release_classes[@]}"; do
                if [[ "${release_classes[$index]}" == "$stage_name" ]]; then
                    stage_index=$((index + 1))
                    break
                fi
            done

            file_version_suffix=$(((stage_index * 10000) + (10#${stage_number} * 100)))
        fi
    fi

    if [[ -n "$PARSED_AHEAD" ]]; then
        version+="+${PARSED_AHEAD}"
        file_version_suffix=$((file_version_suffix + 10#${PARSED_AHEAD}))
    fi

    BUILD_VERSION="$version"
    BUILD_FILE_VERSION="${PARSED_MAJOR}.${PARSED_MINOR}.${PARSED_PATCH}.${file_version_suffix}"
}

resolve_build_version_values() {
    CURRENT_COMMIT="$(git rev-parse --short HEAD)"
    CURRENT_COMMIT_TAG="$(git describe --tags HEAD 2>/dev/null || true)"
    CURRENT_REPO_TAG="$(get_most_recent_tag)"

    if parse_version_hash "$CURRENT_COMMIT_TAG"; then
        get_build_version_info
    elif parse_version_hash "$CURRENT_REPO_TAG"; then
        get_build_version_info
    else
        get_date_stamp
        BUILD_VERSION="0.0.0-dev+${CURRENT_COMMIT}"
        BUILD_FILE_VERSION="0.0.0.${DATE_STAMP}"
    fi
}

confirm_repository_is_clean() {
    local ignore_managed_files="$1"
    local quiet="$2"
    local line path_part file
    local dirty_lines=()

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        path_part="${line:3}"

        if [[ "$path_part" == *"$SCRIPT_NAME" ]]; then
            continue
        fi

        if [[ "$ignore_managed_files" == "1" ]]; then
            local skip_line=0
            for file in "${MANAGED_FILES[@]}"; do
                if [[ "$path_part" == "$file" || "$path_part" == *" -> $file" ]]; then
                    skip_line=1
                    break
                fi
            done
            if [[ "$skip_line" -eq 1 ]]; then
                continue
            fi
        fi

        dirty_lines+=("$line")
    done < <(git status --porcelain=v1)

    if [[ "${#dirty_lines[@]}" -eq 0 ]]; then
        return 0
    fi

    if [[ "$quiet" != "1" ]]; then
        echo
        echo "WARNING: Repository is not clean. Please commit or stash your changes before building." >&2
        echo >&2
        echo "Uncommitted changes:" >&2
        echo >&2
        printf '  %s\n' "${dirty_lines[@]}" >&2
        echo >&2
        echo "Suggestions:" >&2
        echo "  - Commit your changes: git commit -m 'Your commit message'" >&2
        echo "  - Create a new branch: git checkout -b new-branch-name" >&2
        echo "  - Stash your changes: git stash --all" >&2
        echo "  - Discard your changes: git reset --hard HEAD" >&2
        echo >&2
        echo "NOTE: Script ignores changes to ${SCRIPT_NAME}" >&2
        echo >&2
    fi

    return 1
}

restore_repository_to_clean_state() {
    local file
    echo "Restoring repository to a clean state..."
    for file in "${MANAGED_FILES[@]}"; do
        if git status --porcelain=v1 -- "$file" | grep -q .; then
            echo "  Restoring: $file"
            git restore "$file"
        fi
    done
    RESTORE_NEEDED=0
}

update_wails_json() {
    local version="$1"
    local file_path="$PROJECT_ROOT/wails.json"
    perl -0pi -e 's/"productVersion"\s*:\s*"[^"]+"/"productVersion": "'"$version"'"/' "$file_path"
}

update_package_json() {
    local version="$1"
    local file_path="$PROJECT_ROOT/frontend/package.json"
    perl -0pi -e 's/"version"\s*:\s*"[^"]+"/"version": "'"$version"'"/' "$file_path"
}

find_linux_binary() {
    local explicit_path="$PROJECT_ROOT/build/bin/texto-vista"
    if [[ -f "$explicit_path" ]]; then
        printf '%s\n' "$explicit_path"
        return 0
    fi

    explicit_path="$PROJECT_ROOT/build/bin/textovista"
    if [[ -f "$explicit_path" ]]; then
        printf '%s\n' "$explicit_path"
        return 0
    fi

    find "$PROJECT_ROOT/build/bin" -maxdepth 1 -type f \( -name 'texto-vista*' -o -name 'textovista*' \) \
        ! -name '*.sha1' ! -name '*.sha256' ! -name '*.tar.gz' ! -name '*.deb' ! -name '*.exe' \
        | sort \
        | head -n 1
}

write_file_hashes() {
    local file_path="$1"
    local file_name sha256_hash sha1_hash

    if [[ ! -f "$file_path" ]]; then
        echo "Error: Cannot write hashes for missing file: $file_path" >&2
        return 1
    fi

    file_name="$(basename "$file_path")"
    sha256_hash="$(sha256sum "$file_path" | awk '{print $1}')"
    sha1_hash="$(sha1sum "$file_path" | awk '{print $1}')"

    printf '%s  %s\n' "$sha256_hash" "$file_name" > "${file_path}.sha256"
    printf '%s  %s\n' "$sha1_hash" "$file_name" > "${file_path}.sha1"
}

create_tar_package() {
    local version="$1"
    local binary_path="$2"
    local stage_dir archive_path

    stage_dir="$PROJECT_ROOT/build/package/linux-tar-gz"
    archive_path="$PROJECT_ROOT/build/bin/textovista-${version}-linux-amd64.tar.gz"

    rm -rf "$stage_dir"
    mkdir -p "$stage_dir"

    cp "$binary_path" "$stage_dir/textovista"
    chmod +x "$stage_dir/textovista"

    tar -C "$stage_dir" -czf "$archive_path" .

    if [[ "$KEEP_STAGE_DIR" -eq 0 ]]; then
        rm -rf "$stage_dir"
    fi

    printf '%s\n' "$archive_path"
}

print_repository_information() {
    local repo_is_clean="$1"

    echo
    echo "Current Repository Information:"
    echo "  Most Recent Repo Tag: ${CURRENT_REPO_TAG:-<none>}"
    if [[ "$repo_is_clean" -eq 1 ]]; then
        echo "  Current Repo Status: Clean (Safe to build)"
    else
        echo "  Current Repo Status: Dirty (Uncommitted changes found)"
    fi
    echo
    echo "Version Values For Next Build:"
    echo "  Semantic Version: ${BUILD_VERSION}"
    echo "  Numeric Version:  ${BUILD_FILE_VERSION}"
}

invoke_wails_build() {
    local repo_is_clean=0
    local build_date binary_path archive_path

    require_command git
    resolve_build_version_values

    if confirm_repository_is_clean 1 1; then
        repo_is_clean=1
    fi

    if [[ "$SHOW_VERSION_ONLY" -eq 1 ]]; then
        print_repository_information "$repo_is_clean"
        if [[ "$repo_is_clean" -eq 0 ]]; then
            echo
            echo "Repository is not clean. Resolve the listed changes before running a distribution build." >&2
        fi
        return 0
    fi

    if ! confirm_repository_is_clean 1 0; then
        return 1
    fi

    require_command perl
    require_command sha1sum
    require_command sha256sum
    require_command tar
    require_command wails

    build_date="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

    echo "Updating build metadata to version: ${BUILD_VERSION}"
    update_wails_json "$BUILD_VERSION"
    update_package_json "$BUILD_VERSION"
    RESTORE_NEEDED=1

    echo
    echo "Building Wails application with version value: ${BUILD_VERSION}"
    echo
    wails build -clean -tags webkit2_41 -ldflags "-X main.Version=${BUILD_VERSION} -X main.Date=${build_date} -X main.Commit=${CURRENT_COMMIT} -s -w"
    echo

    binary_path="$(find_linux_binary)"
    if [[ -z "$binary_path" ]]; then
        echo "Error: No Linux binary found in $PROJECT_ROOT/build/bin after build." >&2
        return 1
    fi

    echo "Writing hashes for executable: $(basename "$binary_path")"
    write_file_hashes "$binary_path"

    if [[ "$DO_TAR" -eq 1 ]]; then
        archive_path="$(create_tar_package "$BUILD_VERSION" "$binary_path")"
        echo "Writing hashes for archive: $(basename "$archive_path")"
        write_file_hashes "$archive_path"
    fi

    if [[ "$DO_DEB" -eq 1 ]]; then
        echo "DEB packaging is not implemented yet. Stub only; no .deb package was created."
    fi
}

cd "$PROJECT_ROOT"
invoke_wails_build