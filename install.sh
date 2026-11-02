#!/bin/bash

set -u

# -----------------------------------------------------------------------------
# YukkiMusic Dependency & Environment Installer
# -----------------------------------------------------------------------------

NTGCALLS_VERSION="v3.0.5"
GO_REQUIRED="1.27.0"
GO_TARGET="1.27.0"
GOTDBOT_TOOLS_VERSION="v0.10.1"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
INSTALL_LOG="$(mktemp "${TMPDIR:-/tmp}/yukki_install_XXXXXX.log" 2>/dev/null || mktemp /tmp/yukki_install_XXXXXX 2>/dev/null || echo "/tmp/yukki_install_$$.log")"
TMP_DIR=""
SUDO_KEEPALIVE_PID=""
SUDO_AUTHENTICATED=0
USE_SUDO=false

# Terminal Colors
RED="\033[1;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
CYAN="\033[1;36m"
BOLD="\033[1m"
RESET="\033[0m"

# State
WARNINGS=0
PKG_MANAGER_UPDATED=0
SHELL_PROFILE_UPDATED=0
UPDATED_PROFILE=""

INSTALL_ALL=true
INSTALL_GO=false
INSTALL_DENO=false
INSTALL_FFMPEG=false
INSTALL_YTDLP=false
INSTALL_NTGCALLS=false
INSTALL_TDLIB=false
SKIP_SUMMARY=false
QUIET_MODE=false

OS_TYPE=""
ARCH_TYPE=""
PKG_MGR=""
DOWNLOAD_TOOL=""

# -----------------------------------------------------------------------------
# Cleanup & Traps
# -----------------------------------------------------------------------------
cleanup() {
    local exit_code=$?
    trap - EXIT INT TERM HUP
    if [[ -n "$SUDO_KEEPALIVE_PID" ]] && kill -0 "$SUDO_KEEPALIVE_PID" 2>/dev/null; then
        kill "$SUDO_KEEPALIVE_PID" 2>/dev/null || true
    fi
    if [[ -n "$TMP_DIR" && -d "$TMP_DIR" ]]; then
        rm -rf "$TMP_DIR"
    fi
    exit "$exit_code"
}

setup_workspace() {
    TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/yukki_tmp_XXXXXX" 2>/dev/null || mktemp -d /tmp/yukki_tmp_XXXXXX)"
    trap cleanup EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    trap 'exit 129' HUP
}

# -----------------------------------------------------------------------------
# Output & Logging Helpers
# -----------------------------------------------------------------------------
print_step() {
    [[ "$QUIET_MODE" == true ]] && return
    echo -e "\n${BLUE}▶ $1${RESET}"
}

print_success() {
    [[ "$QUIET_MODE" == true ]] && echo "✓ $1" && return
    echo -e "${GREEN}✓ $1${RESET}"
}

print_warning() {
    [[ "$QUIET_MODE" == true ]] && echo "⚠ $1" && return
    echo -e "${YELLOW}⚠ $1${RESET}"
}

print_info() {
    [[ "$QUIET_MODE" == true ]] && return
    echo -e "${CYAN}ℹ $1${RESET}"
}

print_soft_error() {
    echo -e "${RED}✗ $1${RESET}"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1" >> "$INSTALL_LOG"
    WARNINGS=$((WARNINGS + 1))
}

print_error() {
    echo -e "\n${RED}${BOLD}✗ CRITICAL: $1${RESET}"
    echo -e "${RED}${BOLD}Please $2 manually.${RESET}\n"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] CRITICAL: $1" >> "$INSTALL_LOG"
    exit 1
}

log_info() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1" >> "$INSTALL_LOG"
}

run_cmd_arr() {
    local desc="$1"
    shift
    log_info "Executing: $desc ($*)"
    if "$@" >> "$INSTALL_LOG" 2>&1; then
        return 0
    else
        local exit_status=$?
        log_info "Failed: $desc (exit status: $exit_status)"
        return "$exit_status"
    fi
}

# -----------------------------------------------------------------------------
# Privilege Management (Lazy Evaluation)
# -----------------------------------------------------------------------------
sudo_preauth() {
    [[ $EUID -eq 0 ]] && { USE_SUDO=false; return 0; }
    [[ $SUDO_AUTHENTICATED -eq 1 ]] && return 0

    if command -v sudo >/dev/null 2>&1; then
        if sudo -n true 2>/dev/null || (print_info "Requesting sudo privileges for installation..." && sudo -v); then
            (while kill -0 "$$" 2>/dev/null; do sudo -n true 2>/dev/null; sleep 50; done) >/dev/null 2>&1 &
            SUDO_KEEPALIVE_PID=$!
            USE_SUDO=true
            SUDO_AUTHENTICATED=1
            return 0
        else
            print_warning "Sudo authentication failed; proceeding with current user permissions"
            USE_SUDO=false
            SUDO_AUTHENTICATED=1
            return 1
        fi
    fi
    return 1
}

run_as_root() {
    local desc="$1"
    shift
    if [[ $EUID -eq 0 ]]; then
        run_cmd_arr "$desc" "$@"
    elif command -v sudo >/dev/null 2>&1; then
        sudo_preauth
        if [[ "$USE_SUDO" == true ]]; then
            run_cmd_arr "$desc" sudo "$@"
        else
            run_cmd_arr "$desc" "$@"
        fi
    else
        run_cmd_arr "$desc" "$@"
    fi
}

# -----------------------------------------------------------------------------
# System & Package Manager Detection
# -----------------------------------------------------------------------------
detect_system() {
    print_step "Detecting system..."

    local os arch
    os="$(uname -s)"
    arch="$(uname -m)"

    case "$os" in
        Linux)                OS_TYPE="linux" ;;
        Darwin)               OS_TYPE="macos" ;;
        MINGW*|MSYS*|CYGWIN*) OS_TYPE="windows" ;;
        *)  print_error "Unsupported OS: $os" "use Linux, macOS, or Windows" ;;
    esac

    case "$arch" in
        x86_64|amd64)   ARCH_TYPE="amd64" ;;
        aarch64|arm64)  ARCH_TYPE="arm64" ;;
        *) print_error "Unsupported architecture: $arch" "use x86_64 or arm64" ;;
    esac

    case "$OS_TYPE" in
        linux)
            if   command -v apt-get >/dev/null 2>&1; then PKG_MGR="apt"
            elif command -v apk     >/dev/null 2>&1; then PKG_MGR="apk"
            elif command -v dnf     >/dev/null 2>&1; then PKG_MGR="dnf"
            elif command -v yum     >/dev/null 2>&1; then PKG_MGR="yum"
            elif command -v pacman  >/dev/null 2>&1; then PKG_MGR="pacman"
            else PKG_MGR="unknown"
            fi
            ;;
        macos)
            if command -v brew >/dev/null 2>&1; then PKG_MGR="brew"
            else PKG_MGR="unknown"
            fi
            ;;
        windows)
            PKG_MGR="windows"
            ;;
    esac

    print_success "System: $OS_TYPE ($ARCH_TYPE), Package manager: $PKG_MGR"
    log_info "Detected system: $OS_TYPE ($ARCH_TYPE), pkg_mgr=$PKG_MGR"
}

refresh_package_manager() {
    [[ $PKG_MANAGER_UPDATED -eq 1 ]] && return 0
    log_info "Updating package manager repositories..."

    case "$PKG_MGR" in
        apt)    run_as_root "Updating apt repositories" apt-get update -y ;;
        apk)    run_as_root "Updating apk repositories" apk update ;;
        dnf)    run_as_root "Checking dnf updates" dnf check-update || true ;;
        yum)    run_as_root "Checking yum updates" yum check-update || true ;;
        pacman) run_as_root "Synchronizing pacman db" pacman -Sy --noconfirm ;;
        brew)   run_cmd_arr "Updating Homebrew" brew update ;;
    esac

    PKG_MANAGER_UPDATED=1
}

install_package() {
    local package="$1"
    local display_name="${2:-$package}"

    print_info "Installing $display_name..."
    refresh_package_manager

    local cmd=()
    case "$PKG_MGR" in
        apt)    cmd=(apt-get install -y --no-install-recommends "$package") ;;
        apk)    cmd=(apk add --no-cache "$package") ;;
        dnf)    cmd=(dnf install -y "$package") ;;
        yum)    cmd=(yum install -y "$package") ;;
        pacman) cmd=(pacman -S --noconfirm --needed "$package") ;;
        brew)   cmd=(brew install "$package") ;;
        *)      print_soft_error "Unsupported package manager: $PKG_MGR"; return 1 ;;
    esac

    local runner="run_as_root"
    [[ "$PKG_MGR" == "brew" ]] && runner="run_cmd_arr"

    if "$runner" "Installing $package via $PKG_MGR" "${cmd[@]}"; then
        print_success "$display_name installed via $PKG_MGR"
        return 0
    fi

    print_soft_error "Failed to install $display_name"
    return 1
}

# -----------------------------------------------------------------------------
# Base Prerequisites & File Downloader
# -----------------------------------------------------------------------------
ensure_prerequisites() {
    print_step "Checking base prerequisites..."

    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        print_warning "No download tool found, installing curl..."
        install_package "curl" "curl" || print_error "Could not install curl" "install curl or wget manually"
    fi

    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_TOOL="curl"
    else
        DOWNLOAD_TOOL="wget"
    fi
    print_success "Download tool: $DOWNLOAD_TOOL"

    if ! command -v tar >/dev/null 2>&1; then
        install_package "tar" "tar" || true
    fi

    if ! command -v unzip >/dev/null 2>&1; then
        install_package "unzip" "unzip" || true
    fi
}

download_file() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry_count=0

    while [[ $retry_count -lt $max_retries ]]; do
        case "$DOWNLOAD_TOOL" in
            curl) run_cmd_arr "Downloading $url" curl -sSL -f -o "$output" "$url" && return 0 ;;
            wget) run_cmd_arr "Downloading $url" wget -q -O "$output" "$url" && return 0 ;;
            *)    return 1 ;;
        esac

        retry_count=$((retry_count + 1))
        if [[ $retry_count -lt $max_retries ]]; then
            print_info "Download failed, retrying... ($retry_count/$max_retries)"
            sleep 2
        fi
    done

    return 1
}

update_path() {
    local new_path="$1"
    local desc="${2:-$1}"

    if [[ ":$PATH:" != *":$new_path:"* ]]; then
        export PATH="$new_path:$PATH"
    fi

    local profile=""
    if [[ -n "${SHELL:-}" && "$SHELL" == *"zsh"* ]]; then
        profile="$HOME/.zshrc"
    elif [[ -f "$HOME/.bashrc" ]]; then
        profile="$HOME/.bashrc"
    elif [[ -f "$HOME/.profile" ]]; then
        profile="$HOME/.profile"
    fi

    if [[ -n "$profile" ]]; then
        touch "$profile" 2>/dev/null || true
        if ! grep -qF -- "$new_path" "$profile" 2>/dev/null; then
            echo "export PATH=\"$new_path:\$PATH\"" >> "$profile"
            print_info "Added $desc to PATH in $profile"
            SHELL_PROFILE_UPDATED=1
            UPDATED_PROFILE="$profile"
        fi
    fi
}

version_ge() {
    awk -v v1="$1" -v v2="$2" 'BEGIN {
        split(v1, a, "."); split(v2, b, ".");
        for (i = 1; i <= 3; i++) {
            n1 = a[i] + 0; n2 = b[i] + 0;
            if (n1 != n2) exit (n1 < n2);
        }
        exit 0;
    }'
}

find_libtdjson() {
    local f
    for f in "$SCRIPT_DIR"/libtdjson.so* "$SCRIPT_DIR"/libtdjson.dylib* "$SCRIPT_DIR"/tdjson.dll; do
        if [[ -f "$f" ]]; then
            echo "$f"
            return 0
        fi
    done
    return 1
}

# -----------------------------------------------------------------------------
# Component Check & Installation
# -----------------------------------------------------------------------------
should_install() {
    local component="$1"

    [[ "$INSTALL_ALL" == true ]] && return 0

    case "$component" in
        go)       [[ "$INSTALL_GO"       == true ]] && return 0 ;;
        deno)     [[ "$INSTALL_DENO"     == true ]] && return 0 ;;
        ffmpeg)   [[ "$INSTALL_FFMPEG"   == true ]] && return 0 ;;
        ytdlp)    [[ "$INSTALL_YTDLP"    == true ]] && return 0 ;;
        ntgcalls) [[ "$INSTALL_NTGCALLS" == true ]] && return 0 ;;
        tdlib)    [[ "$INSTALL_TDLIB"    == true ]] && return 0 ;;
    esac

    return 1
}

check_install_go() {
    should_install go || return 0

    print_step "Checking Go..."

    local current_go=""
    if command -v go >/dev/null 2>&1; then
        current_go=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
        if [[ -n "$current_go" ]] && version_ge "$current_go" "$GO_REQUIRED"; then
            print_success "Go $current_go (sufficient)"
            return 0
        fi
        print_warning "Go $current_go is outdated (need $GO_REQUIRED+), upgrading..."
    fi

    case "$OS_TYPE" in
        linux)
            if [[ "$PKG_MGR" == "apk" ]]; then
                print_info "Installing Go via apk..."
                if install_package "go" "Go"; then
                    return 0
                fi
            fi

            print_info "Installing Go $GO_TARGET via official tarball..."
            local archive="go${GO_TARGET}.linux-${ARCH_TYPE}.tar.gz"
            local url="https://go.dev/dl/${archive}"
            local dl_path="$TMP_DIR/${archive}"

            if download_file "$url" "$dl_path"; then
                run_as_root "Removing old Go" rm -rf /usr/local/go
                run_as_root "Extracting Go" tar -C /usr/local -xzf "$dl_path"
                update_path "/usr/local/go/bin" "Go"

                if /usr/local/go/bin/go version >/dev/null 2>&1; then
                    local ver
                    ver=$(/usr/local/go/bin/go version | awk '{print $3}' | sed 's/go//')
                    print_success "Go $ver installed via tarball"
                    return 0
                fi
            fi

            # Fallback to snap if tarball extraction failed
            if command -v snap >/dev/null 2>&1; then
                print_info "Attempting Go install via snap fallback..."
                if run_as_root "Installing Go via snap" snap install go --classic; then
                    update_path "/snap/bin" "Snap"
                    local snap_ver
                    snap_ver=$(/snap/bin/go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
                    print_success "Go $snap_ver installed via snap"
                    return 0
                fi
            fi

            print_soft_error "Go installation failed"
            return 1
            ;;

        macos)
            if command -v brew >/dev/null 2>&1; then
                print_info "Installing/upgrading Go via Homebrew..."
                run_cmd_arr "brew install go" brew install go || run_cmd_arr "brew upgrade go" brew upgrade go
                local final_ver
                final_ver=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
                if [[ -n "$final_ver" ]] && version_ge "$final_ver" "$GO_REQUIRED"; then
                    print_success "Go $final_ver installed via Homebrew"
                    return 0
                fi
            fi
            print_soft_error "Homebrew Go installation failed"
            return 1
            ;;

        windows)
            print_warning "Please install Go $GO_REQUIRED+ manually on Windows from https://go.dev/dl/"
            WARNINGS=$((WARNINGS + 1))
            return 1
            ;;
    esac
}

check_install_deno() {
    should_install deno || return 0

    print_step "Checking Deno..."

    if command -v deno >/dev/null 2>&1; then
        local ver
        ver=$(deno --version 2>/dev/null | head -n1 | awk '{print $2}')
        print_success "Deno already installed ($ver)"
        return 0
    fi

    print_warning "Deno not found, installing..."

    if [[ "$OS_TYPE" == "windows" ]]; then
        if command -v powershell >/dev/null 2>&1; then
            if run_cmd_arr "Installing Deno for Windows" powershell -Command "irm https://deno.land/install.ps1 | iex"; then
                update_path "$HOME/.deno/bin" "Deno"
                print_success "Deno installed"
                return 0
            fi
        fi
    else
        local deno_script="$TMP_DIR/deno_install.sh"
        if download_file "https://deno.land/install.sh" "$deno_script"; then
            chmod +x "$deno_script"
            if run_cmd_arr "Running Deno installer" sh "$deno_script"; then
                update_path "$HOME/.deno/bin" "Deno"
                if command -v deno >/dev/null 2>&1 || [[ -x "$HOME/.deno/bin/deno" ]]; then
                    print_success "Deno installed"
                    return 0
                fi
            fi
        fi
    fi

    print_soft_error "Deno installation failed"
    return 1
}

check_install_ffmpeg() {
    should_install ffmpeg || return 0

    print_step "Checking FFmpeg & FFprobe..."

    local has_ffmpeg=false
    local has_ffprobe=false

    command -v ffmpeg >/dev/null 2>&1 && has_ffmpeg=true
    command -v ffprobe >/dev/null 2>&1 && has_ffprobe=true

    if [[ "$has_ffmpeg" == true && "$has_ffprobe" == true ]]; then
        local ver
        ver=$(ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}')
        print_success "FFmpeg & FFprobe installed ($ver)"
        return 0
    fi

    print_warning "FFmpeg/FFprobe missing or incomplete, installing..."

    if [[ "$OS_TYPE" == "windows" ]]; then
        local url="https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
        local zip="$TMP_DIR/ffmpeg.zip"
        local install_dir="$HOME/ffmpeg"

        if download_file "$url" "$zip"; then
            mkdir -p "$install_dir"
            if run_cmd_arr "Extracting FFmpeg" unzip -q "$zip" -d "$TMP_DIR/"; then
                local extracted
                extracted=$(find "$TMP_DIR" -maxdepth 1 -type d -name "ffmpeg-*-essentials_build" | head -n1)
                if [[ -n "$extracted" ]]; then
                    cp -r "$extracted/bin/"* "$install_dir/"
                    update_path "$install_dir" "FFmpeg"
                    command -v ffmpeg >/dev/null 2>&1 && { print_success "FFmpeg & FFprobe installed"; return 0; }
                fi
            fi
        fi
        print_soft_error "FFmpeg installation failed on Windows"
        return 1
    fi

    if [[ "$OS_TYPE" == "linux" ]] && { [[ "$PKG_MGR" == "yum" ]] || [[ "$PKG_MGR" == "dnf" ]]; }; then
        run_as_root "Installing EPEL" "$PKG_MGR" install -y epel-release || true
        local el_ver
        el_ver=$(rpm -E %rhel 2>/dev/null || echo "9")
        run_as_root "Installing RPM Fusion" "$PKG_MGR" install -y \
            "https://mirrors.rpmfusion.org/free/el/rpmfusion-free-release-${el_ver}.noarch.rpm" 2>/dev/null || true
    fi

    install_package "ffmpeg" "FFmpeg"

    if command -v ffmpeg >/dev/null 2>&1; then
        if command -v ffprobe >/dev/null 2>&1; then
            print_success "FFmpeg & FFprobe installed"
            return 0
        else
            print_warning "FFmpeg installed, but ffprobe was not found"
            WARNINGS=$((WARNINGS + 1))
            return 0
        fi
    fi

    print_soft_error "FFmpeg installation failed"
    return 1
}

check_install_ytdlp() {
    should_install ytdlp || return 0

    print_step "Checking yt-dlp..."

    if command -v yt-dlp >/dev/null 2>&1; then
        local ver
        ver=$(yt-dlp --version 2>/dev/null || echo "installed")
        print_success "yt-dlp installed ($ver)"
        return 0
    fi

    print_warning "yt-dlp not found, downloading standalone binary..."

    local url=""
    local target_name="yt-dlp"

    case "$OS_TYPE" in
        linux)
            if [[ "$ARCH_TYPE" == "amd64" ]]; then
                url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"
            else
                url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64"
            fi
            ;;
        macos)
            url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
            ;;
        windows)
            url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
            target_name="yt-dlp.exe"
            ;;
    esac

    local dl_path="$TMP_DIR/$target_name"

    if [[ -n "$url" ]] && download_file "$url" "$dl_path"; then
        chmod +x "$dl_path"

        if [[ "$OS_TYPE" == "windows" ]]; then
            mkdir -p "$HOME/bin"
            mv "$dl_path" "$HOME/bin/$target_name"
            update_path "$HOME/bin" "yt-dlp"
            print_success "yt-dlp installed to $HOME/bin"
            return 0
        fi

        run_as_root "Creating /usr/local/bin" mkdir -p /usr/local/bin
        if run_as_root "Installing yt-dlp to /usr/local/bin" cp "$dl_path" /usr/local/bin/yt-dlp; then
            run_as_root "Setting yt-dlp permissions" chmod 0755 /usr/local/bin/yt-dlp
            print_success "yt-dlp binary installed to /usr/local/bin"
            return 0
        else
            mkdir -p "$HOME/.local/bin"
            cp "$dl_path" "$HOME/.local/bin/yt-dlp"
            chmod 0755 "$HOME/.local/bin/yt-dlp"
            update_path "$HOME/.local/bin" "local bin"
            print_success "yt-dlp binary installed to ~/.local/bin"
            return 0
        fi
    fi

    print_soft_error "yt-dlp installation failed"
    return 1
}

install_ntgcalls() {
    should_install ntgcalls || return 0

    print_step "Installing ntgcalls ($NTGCALLS_VERSION)..."

    local url=""
    case "$OS_TYPE" in
        linux)
            if [[ "$ARCH_TYPE" == "amd64" ]]; then
                url="https://github.com/pytgcalls/ntgcalls/releases/download/$NTGCALLS_VERSION/ntgcalls.linux-x86_64-static_libs.zip"
            else
                url="https://github.com/pytgcalls/ntgcalls/releases/download/$NTGCALLS_VERSION/ntgcalls.linux-arm64-static_libs.zip"
            fi
            ;;
        macos)
            if [[ "$ARCH_TYPE" == "arm64" ]]; then
                url="https://github.com/pytgcalls/ntgcalls/releases/download/$NTGCALLS_VERSION/ntgcalls.macos-arm64-static_libs.zip"
            else
                print_error "ntgcalls prebuilt binaries unavailable for macOS x86_64" "build ntgcalls from source"
            fi
            ;;
        windows)
            url="https://github.com/pytgcalls/ntgcalls/releases/download/$NTGCALLS_VERSION/ntgcalls.windows-x86_64-shared_libs.zip"
            ;;
    esac

    [[ -z "$url" ]] && print_error "Could not determine ntgcalls URL" "check system compatibility"

    local zip_path="$TMP_DIR/ntgcalls.zip"
    local extract_dir="$TMP_DIR/ntg_extracted"

    if download_file "$url" "$zip_path"; then
        mkdir -p "$extract_dir"
        if run_cmd_arr "Extracting ntgcalls" unzip -q "$zip_path" -d "$extract_dir"; then
            mkdir -p "$SCRIPT_DIR/ntgcalls"

            if [[ "$OS_TYPE" == "windows" ]]; then
                cp -f "$extract_dir/include/ntgcalls.h" "$SCRIPT_DIR/ntgcalls/" 2>/dev/null || true
                find "$extract_dir" -type f \( -name "*.dll" -o -name "*.lib" -o -name "*.a" -o -name "*.exp" \) -exec cp -f {} "$SCRIPT_DIR/" \; 2>/dev/null || true
                print_success "ntgcalls ($NTGCALLS_VERSION) installed in $SCRIPT_DIR"
                return 0
            else
                local lib_file
                lib_file=$(find "$extract_dir/lib" -maxdepth 1 -type f \( -name "*.a" -o -name "*.so*" -o -name "*.dylib" -o -name "*.lib" \) 2>/dev/null | head -n1)

                if [[ -f "$extract_dir/include/ntgcalls.h" && -n "$lib_file" ]]; then
                    if cp -f "$extract_dir/include/ntgcalls.h" "$SCRIPT_DIR/ntgcalls/" 2>/dev/null && \
                       cp -f "$lib_file" "$SCRIPT_DIR/$(basename "$lib_file")" 2>/dev/null; then
                        print_success "ntgcalls ($NTGCALLS_VERSION) installed in $SCRIPT_DIR"
                        return 0
                    elif run_as_root "Installing ntgcalls header" cp -f "$extract_dir/include/ntgcalls.h" "$SCRIPT_DIR/ntgcalls/" && \
                         run_as_root "Installing ntgcalls library" cp -f "$lib_file" "$SCRIPT_DIR/$(basename "$lib_file")"; then
                        print_success "ntgcalls ($NTGCALLS_VERSION) installed in $SCRIPT_DIR"
                        return 0
                    fi
                fi
            fi
        fi
    fi

    print_soft_error "ntgcalls installation failed"
    return 1
}

install_tdlib() {
    should_install tdlib || return 0

    print_step "Checking libtdjson (TDLib)..."

    local existing_td
    if existing_td=$(find_libtdjson); then
        print_success "libtdjson already present ($(basename "$existing_td"))"
        return 0
    fi

    if command -v go >/dev/null 2>&1; then
        print_info "Downloading libtdjson via gotdbot tools..."
        if (cd "$SCRIPT_DIR" && run_cmd_arr "Fetching libtdjson" go run "github.com/AshokShau/gotdbot/scripts/tools@$GOTDBOT_TOOLS_VERSION"); then
            if existing_td=$(find_libtdjson); then
                print_success "libtdjson installed ($(basename "$existing_td"))"
                return 0
            fi
        fi
    fi

    print_warning "libtdjson not found. You can run: (cd $SCRIPT_DIR && go run github.com/AshokShau/gotdbot/scripts/tools@$GOTDBOT_TOOLS_VERSION)"
    WARNINGS=$((WARNINGS + 1))
    return 1
}

# -----------------------------------------------------------------------------
# Post-Install Automation
# -----------------------------------------------------------------------------
setup_environment_file() {
    [[ "$INSTALL_ALL" == false ]] && return 0
    if [[ ! -f "$SCRIPT_DIR/.env" && -f "$SCRIPT_DIR/sample.env" ]]; then
        cp "$SCRIPT_DIR/sample.env" "$SCRIPT_DIR/.env"
        print_info "Created .env from sample.env. Please edit .env with your Telegram API credentials."
    fi
}

# -----------------------------------------------------------------------------
# Summary & Status Reporting
# -----------------------------------------------------------------------------
_get_ver() {
    case "$1" in
        go)       go version 2>/dev/null | awk '{print $3}' | sed 's/go//' ;;
        deno)     deno --version 2>/dev/null | head -n1 | awk '{print $2}' ;;
        ffmpeg)   ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}' ;;
        ffprobe)  ffprobe -version 2>/dev/null | head -n1 | awk '{print $3}' ;;
        ytdlp)    yt-dlp --version 2>/dev/null | head -n1 ;;
        ntgcalls) echo "$NTGCALLS_VERSION" ;;
        tdlib)
            local td
            td=$(find_libtdjson) && basename "$td" || echo "ready"
            ;;
    esac
}

_is_installed() {
    case "$1" in
        go)       command -v go >/dev/null 2>&1 ;;
        deno)     command -v deno >/dev/null 2>&1 ;;
        ffmpeg)   command -v ffmpeg >/dev/null 2>&1 ;;
        ffprobe)  command -v ffprobe >/dev/null 2>&1 ;;
        ytdlp)    command -v yt-dlp >/dev/null 2>&1 ;;
        ntgcalls)
            [[ -f "$SCRIPT_DIR/ntgcalls/ntgcalls.h" ]] && \
                find "$SCRIPT_DIR" -maxdepth 1 \( -name "libntgcalls.*" -o -name "ntgcalls.*" \) 2>/dev/null | grep -q .
            ;;
        tdlib)    find_libtdjson >/dev/null 2>&1 ;;
    esac
}

print_summary() {
    [[ "$QUIET_MODE"   == true ]] && return
    [[ "$SKIP_SUMMARY" == true ]] && return

    echo -e "\n${BLUE}${BOLD}==================================================${RESET}"
    echo -e "${BLUE}${BOLD}             INSTALLATION SUMMARY               ${RESET}"
    echo -e "${BLUE}${BOLD}==================================================${RESET}"
    printf "${CYAN}${BOLD}%-15s | %-12s | %-15s${RESET}\n" "Component" "Status" "Version"
    echo -e "--------------------------------------------------"

    _print_row() {
        local label="$1" key="$2"
        local status ver color

        if should_install "$key"; then
            if _is_installed "$key"; then
                status="Installed"
                ver=$(_get_ver "$key")
                [[ -z "$ver" ]] && ver="ready"
                color=$GREEN
            else
                status="Failed"
                ver="-"
                color=$RED
            fi
        else
            status="Skipped"
            ver="-"
            color=$YELLOW
        fi

        printf "%-15s | ${color}%-12s${RESET} | %-15s\n" "$label" "$status" "$ver"
    }

    _print_row "Go"        "go"
    _print_row "Deno"      "deno"
    _print_row "FFmpeg"    "ffmpeg"
    _print_row "FFprobe"   "ffmpeg"
    _print_row "yt-dlp"    "ytdlp"
    _print_row "ntgcalls"  "ntgcalls"
    _print_row "libtdjson" "tdlib"

    echo -e "--------------------------------------------------"
    echo -e "${CYAN}Detailed log: $INSTALL_LOG${RESET}"

    if [[ $SHELL_PROFILE_UPDATED -eq 1 ]]; then
        echo -e "\n${YELLOW}${BOLD}Note on PATH:${RESET} Environment paths were updated in your profile."
        echo -e "${YELLOW}To use newly installed tools in your current terminal session, run:${RESET}"
        echo -e "  ${CYAN}source ${UPDATED_PROFILE:-~/.bashrc}${RESET}  # (or restart your terminal)\n"
    fi
}

# -----------------------------------------------------------------------------
# CLI Arguments Parser
# -----------------------------------------------------------------------------
show_help() {
    printf "%b\n" \
"${CYAN}${BOLD}Usage:${RESET} $0 [OPTIONS]

${CYAN}${BOLD}Component Options:${RESET}
  -a, --all               Install all components and setup environment (default)
  -g, --go                Install Go 1.27+ only
  -d, --deno              Install Deno runtime only
  -f, --ffmpeg            Install FFmpeg and FFprobe only
  -y, --yt-dlp, --ytdlp   Install yt-dlp standalone binary only
  -n, --ntgcalls          Install ntgcalls CGo headers & static library only
  -t, --tdlib, --libtdjson Install libtdjson (gotdbot) shared library only

${CYAN}${BOLD}Behavioral Options:${RESET}
  -q, --quiet             Quiet mode (minimal output, ideal for Docker/CI)
  --skip-summary          Skip final summary table
  -h, --help              Show this help message
"
    exit 0
}

parse_arguments() {
    local any_component=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)                 show_help ;;
            -a|--all)
                INSTALL_ALL=true
                INSTALL_GO=false
                INSTALL_DENO=false
                INSTALL_FFMPEG=false
                INSTALL_YTDLP=false
                INSTALL_NTGCALLS=false
                INSTALL_TDLIB=false
                any_component=false
                ;;
            -g|--go)                   INSTALL_GO=true;       any_component=true ;;
            -d|--deno)                 INSTALL_DENO=true;     any_component=true ;;
            -f|--ffmpeg)               INSTALL_FFMPEG=true;   any_component=true ;;
            -y|--yt-dlp|--ytdlp)      INSTALL_YTDLP=true;    any_component=true ;;
            -n|--ntgcalls)             INSTALL_NTGCALLS=true; any_component=true ;;
            -t|--tdlib|--libtdjson)    INSTALL_TDLIB=true;    any_component=true ;;
            -q|--quiet)                QUIET_MODE=true ;;
            --skip-summary)            SKIP_SUMMARY=true ;;
            # Backward-compatibility no-ops
            -p|--python|--pip|--break-system-packages|--user|--check|-b|--build|--no-env)
                log_info "Option $1 is deprecated and ignored"
                ;;
            *)
                echo -e "${RED}Unknown option: $1${RESET}"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
        shift
    done

    [[ "$any_component" == true && "$INSTALL_ALL" == true ]] && INSTALL_ALL=false
}

# -----------------------------------------------------------------------------
# Main Entry Point
# -----------------------------------------------------------------------------
main() {
    parse_arguments "$@"
    setup_workspace

    log_info "Installation started with arguments: $*"

    detect_system

    print_step "Starting installation..."
    [[ "$QUIET_MODE" == false ]] && echo -e "${CYAN}Installing selected components...${RESET}\n"

    ensure_prerequisites
    check_install_go
    check_install_deno
    check_install_ffmpeg
    check_install_ytdlp
    install_ntgcalls
    install_tdlib

    setup_environment_file

    print_summary

    log_info "Installation completed with $WARNINGS warning(s)"

    if [[ $WARNINGS -eq 0 ]]; then
        [[ "$QUIET_MODE" == false ]] && echo -e "\n${GREEN}${BOLD}✓ Installation completed successfully! ✓${RESET}"
        exit 0
    else
        echo -e "\n${YELLOW}${BOLD}⚠ Installation completed with $WARNINGS warning(s) ⚠${RESET}"
        echo -e "${YELLOW}⚠ Some components had warnings. Review log: $INSTALL_LOG${RESET}\n"
        exit 1
    fi
}

main "$@"
