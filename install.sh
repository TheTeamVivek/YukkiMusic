#!/bin/bash

set -u

RED="\033[1;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
CYAN="\033[1;36m"
BOLD="\033[1m"
RESET="\033[0m"

WARNINGS=0
SHELL_RELOAD_NEEDED=0
PKG_MANAGER_UPDATED=0
INSTALL_LOG="/tmp/install_$(date +%s).log"

INSTALL_ALL=true
INSTALL_DENO=false
INSTALL_GO=false
INSTALL_PYTHON=false
INSTALL_PIP=false
INSTALL_FFMPEG=false
INSTALL_YTDLP=false
INSTALL_NTGCALLS=false
SKIP_SUMMARY=false
QUIET_MODE=false

NTGCALLS_VERSION="v2.1.0"
GO_REQUIRED="1.25.7"
GO_TARGET="1.25.7"

OS_TYPE=""
ARCH_TYPE=""
DOWNLOAD_TOOL=""
PYTHON_CMD=""

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

show_help() {
    printf "%b\n" \
"${CYAN}${BOLD}Usage:${RESET} $0 [OPTIONS]

${CYAN}${BOLD}Options:${RESET}
  -h, --help              Show this help message
  -a, --all               Install all components (default)
  -g, --go                Install Go only
  -d, --deno              Install Deno only
  -p, --python            Install Python only
  --pip                   Install pip only
  -f, --ffmpeg            Install FFmpeg only
  -y, --yt-dlp            Install yt-dlp only
  -n, --ntgcalls          Install ntgcalls only
  -q, --quiet             Quiet mode (minimal output)
  --skip-summary          Skip final summary

${CYAN}${BOLD}Examples:${RESET}
  $0                      # Install everything
  $0 --deno               # Install only Deno
  $0 --ntgcalls           # Install only ntgcalls (useful for Docker)
  $0 --go --ffmpeg        # Install only Go and FFmpeg
  $0 --python --pip       # Install only Python and pip
  $0 --all --quiet        # Install everything in quiet mode
"
    exit 0
}

parse_arguments() {
    local any_component=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)            show_help ;;
            -a|--all)             INSTALL_ALL=true ;;
            -g|--go)              INSTALL_GO=true;       any_component=true ;;
            -d|--deno)            INSTALL_DENO=true;     any_component=true ;;
            -p|--python)          INSTALL_PYTHON=true;   any_component=true ;;
            --pip)                INSTALL_PIP=true;      any_component=true ;;
            -f|--ffmpeg)          INSTALL_FFMPEG=true;   any_component=true ;;
            -y|--yt-dlp|--ytdlp) INSTALL_YTDLP=true;   any_component=true ;;
            -n|--ntgcalls)        INSTALL_NTGCALLS=true; any_component=true ;;
            -q|--quiet)           QUIET_MODE=true ;;
            --skip-summary)       SKIP_SUMMARY=true ;;
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

should_install() {
    local component="$1"

    [[ "$INSTALL_ALL" == true ]] && return 0

    case "$component" in
        go)       [[ "$INSTALL_GO"       == true ]] && return 0 ;;
        deno)     [[ "$INSTALL_DENO"     == true ]] && return 0 ;;
        python)   [[ "$INSTALL_PYTHON"   == true ]] && return 0 ;;
        pip)      [[ "$INSTALL_PIP"      == true ]] && return 0 ;;
        ffmpeg)   [[ "$INSTALL_FFMPEG"   == true ]] && return 0 ;;
        ytdlp)    [[ "$INSTALL_YTDLP"   == true ]] && return 0 ;;
        ntgcalls) [[ "$INSTALL_NTGCALLS" == true ]] && return 0 ;;
    esac

    return 1
}

detect_system() {
    print_step "Detecting system..."

    local os arch
    os="$(uname -s)"
    arch="$(uname -m)"

    case "$os" in
        Linux)              OS_TYPE="linux" ;;
        Darwin)             OS_TYPE="macos" ;;
        MINGW*|MSYS*|CYGWIN*) OS_TYPE="windows" ;;
        *)  print_error "Unsupported OS: $os" "use Linux/macOS/Windows" ;;
    esac

    case "$arch" in
        x86_64|amd64)   ARCH_TYPE="amd64" ;;
        aarch64|arm64)  ARCH_TYPE="arm64" ;;
        *) print_error "Unsupported arch: $arch" "use x86_64/arm64" ;;
    esac

    print_success "System: $OS_TYPE ($ARCH_TYPE)"
    log_info "Detected system: $OS_TYPE ($ARCH_TYPE)"
}

# Usage: run_cmd_arr "description" cmd arg1 arg2 ...
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


run_as_root() {
    local desc="$1"
    shift
    if [[ $EUID -eq 0 ]]; then
        run_cmd_arr "$desc" "$@"
    elif command -v sudo >/dev/null 2>&1; then
        run_cmd_arr "$desc" sudo "$@"
    else
        run_cmd_arr "$desc" "$@"
    fi
}

mark_shell_reload() { SHELL_RELOAD_NEEDED=1; }

update_path() {
    local new_path="$1"
    local desc="${2:-$1}"

    if [[ ":$PATH:" != *":$new_path:"* ]]; then
        export PATH="$new_path:$PATH"
    fi

    local profile
    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" "$HOME/.bash_profile"; do
        if [[ -f "$profile" ]] && ! grep -qF -- "$new_path" "$profile" 2>/dev/null; then
            echo "export PATH=\"$new_path:\$PATH\"" >> "$profile"
            print_info "Added $desc to PATH in $profile"
            mark_shell_reload
        fi
    done
}

reload_shell_if_needed() {
    [[ $SHELL_RELOAD_NEEDED -ne 1 ]] && return

    print_info "Reloading shell environment..."
    local profile
    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
        if [[ -f "$profile" ]]; then
            # shellcheck disable=SC1090
            source "$profile" 2>/dev/null || true
            print_success "Environment reloaded from $profile"
            break
        fi
    done
    export PATH="/usr/local/go/bin:$HOME/.deno/bin:$HOME/.local/bin:$PATH"
}

refresh_package_manager() {
    [[ $PKG_MANAGER_UPDATED -eq 1 ]] && return 0

    log_info "Updating package manager..."

    case "$OS_TYPE" in
        linux)
            if   command -v apt    >/dev/null 2>&1; then run_as_root "Updating apt"    apt update
            elif command -v dnf    >/dev/null 2>&1; then run_as_root "Updating dnf"    dnf check-update
            elif command -v yum    >/dev/null 2>&1; then run_as_root "Updating yum"    yum check-update
            elif command -v pacman >/dev/null 2>&1; then run_as_root "Updating pacman" pacman -Sy
            fi
            ;;
        macos)
            if command -v brew >/dev/null 2>&1; then
                run_cmd_arr "Updating Homebrew" brew update
            fi
            ;;
    esac

    PKG_MANAGER_UPDATED=1
}

install_package() {
    local package="$1"
    local display_name="${2:-$package}"

    print_info "Installing $display_name..."
    refresh_package_manager

    case "$OS_TYPE" in
        linux)
            if   command -v apt    >/dev/null 2>&1; then
                run_as_root "Installing $package via apt"    apt install -y "$package" \
                    && { print_success "$display_name installed via apt";    return 0; }
            elif command -v dnf    >/dev/null 2>&1; then
                run_as_root "Installing $package via dnf"    dnf install -y "$package" \
                    && { print_success "$display_name installed via dnf";    return 0; }
            elif command -v yum    >/dev/null 2>&1; then
                run_as_root "Installing $package via yum"    yum install -y "$package" \
                    && { print_success "$display_name installed via yum";    return 0; }
            elif command -v pacman >/dev/null 2>&1; then
                run_as_root "Installing $package via pacman" pacman -S --noconfirm "$package" \
                    && { print_success "$display_name installed via pacman"; return 0; }
            fi
            ;;
        macos)
            if command -v brew >/dev/null 2>&1; then
                run_cmd_arr "Installing $package via Homebrew" brew install "$package" \
                    && { print_success "$display_name installed via Homebrew"; return 0; }
            fi
            ;;
    esac

    print_soft_error "Failed to install $display_name"
    return 1
}

ensure_download_tool() {
    print_step "Checking download tools..."

    if command -v curl >/dev/null 2>&1; then
        print_success "curl available"
        DOWNLOAD_TOOL="curl"
        return 0
    fi

    if command -v wget >/dev/null 2>&1; then
        print_success "wget available"
        DOWNLOAD_TOOL="wget"
        return 0
    fi

    print_warning "No download tool found, installing curl..."

    if install_package "curl"; then
        DOWNLOAD_TOOL="curl"
        return 0
    fi

    print_error "Could not install curl/wget" "install curl or wget"
}

download_file() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry_count=0

    while [[ $retry_count -lt $max_retries ]]; do
        case "$DOWNLOAD_TOOL" in
            curl) run_cmd_arr "Downloading $url" curl -sSL -o "$output" "$url" && return 0 ;;
            wget) run_cmd_arr "Downloading $url" wget -q -O "$output" "$url"   && return 0 ;;
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

version_ge() {
    local lowest
    lowest=$(printf '%s\n%s' "$1" "$2" | sort -V | head -n1)
    [[ "$lowest" == "$2" ]]
}

check_install_python() {
    should_install python || return 0

    print_step "Checking Python..."

    local cmd
    for cmd in python3 python; do
        if command -v "$cmd" >/dev/null 2>&1; then
            PYTHON_CMD="$cmd"
            local ver
            ver=$("$PYTHON_CMD" --version 2>&1 | awk '{print $2}')
            if version_ge "$ver" "3.8"; then
                print_success "Python found: $ver"
                export PYTHON_CMD
                return 0
            else
                print_warning "Python $ver is too old (need 3.8+)"
            fi
        fi
    done

    print_warning "Python not found or outdated, installing..."

    case "$OS_TYPE" in
        linux) install_package "python3" "Python3"       && PYTHON_CMD="python3" ;;
        macos) install_package "python@3.12" "Python"    && PYTHON_CMD="python3" ;;
    esac

    export PYTHON_CMD
    if [[ -n "$PYTHON_CMD" ]] && command -v "$PYTHON_CMD" >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

check_install_pip() {
    should_install pip || return 0

    print_step "Checking pip..."

    if [[ -z "$PYTHON_CMD" ]]; then
        print_warning "Python unavailable, skipping pip"
        return 1
    fi

    if run_cmd_arr "Checking pip" "$PYTHON_CMD" -m pip --version; then
        local ver
        ver=$("$PYTHON_CMD" -m pip --version 2>&1 | awk '{print $2}')
        print_success "pip installed ($ver)"
        return 0
    fi

    print_warning "pip not found, installing..."

    case "$OS_TYPE" in
        linux)
            install_package "python3-pip" "pip" && return 0
            ;;
        macos)
            run_cmd_arr "Installing pip via ensurepip" "$PYTHON_CMD" -m ensurepip \
                && { print_success "pip installed via ensurepip"; return 0; }
            ;;
    esac

    print_info "Trying get-pip.py..."
    if download_file "https://bootstrap.pypa.io/get-pip.py" "/tmp/get-pip.py"; then
        if run_cmd_arr "Installing pip via get-pip.py" "$PYTHON_CMD" /tmp/get-pip.py; then
            rm -f /tmp/get-pip.py
            print_success "pip installed via get-pip.py"
            return 0
        fi
        rm -f /tmp/get-pip.py
    fi

    print_soft_error "pip installation failed"
    return 1
}

check_install_go() {
    should_install go || return 0

    print_step "Checking Go..."

    local current_go=""
    if command -v go >/dev/null 2>&1; then
        current_go=$(go version | awk '{print $3}' | sed 's/go//')
        if version_ge "$current_go" "$GO_REQUIRED"; then
            print_success "Go $current_go (sufficient)"
            return 0
        fi
        print_warning "Go $current_go is too old (need $GO_REQUIRED+), reinstalling..."
    fi

    case "$OS_TYPE" in

        linux)
            if command -v snap >/dev/null 2>&1; then
                if [[ -n "$current_go" ]]; then
                    if snap list go >/dev/null 2>&1; then
                        run_as_root "Removing old Go snap" snap remove go
                    elif [[ -d /usr/local/go ]]; then
                        run_as_root "Removing old Go tarball" rm -rf /usr/local/go
                    elif command -v apt >/dev/null 2>&1 && dpkg -l golang-go >/dev/null 2>&1; then
                        run_as_root "Removing apt Go" apt remove -y golang-go golang
                    fi
                fi

                if snap list go >/dev/null 2>&1; then
                    print_info "Refreshing Go via snap..."
                    if run_as_root "Refreshing Go snap" snap refresh go; then
                        print_success "Go $(go version | awk '{print $3}') refreshed via snap"
                        return 0
                    fi
                else
                    print_info "Installing Go $GO_TARGET via snap..."
                    if run_as_root "Installing Go snap" snap install go --classic; then
                        print_success "Go $(go version | awk '{print $3}') installed via snap"
                        return 0
                    fi
                fi

                print_soft_error "Go snap install/refresh failed"
                return 1

            else
                print_info "Installing Go $GO_TARGET via tarball..."
                local archive="go${GO_TARGET}.linux-${ARCH_TYPE}.tar.gz"
                local url="https://go.dev/dl/${archive}"

                if download_file "$url" "/tmp/${archive}"; then
                    run_as_root "Removing old Go" rm -rf /usr/local/go
                    run_as_root "Extracting Go"   tar -C /usr/local -xzf "/tmp/${archive}"
                    rm -f "/tmp/${archive}"
                    update_path "/usr/local/go/bin" "Go"

                    if command -v go >/dev/null 2>&1; then
                        print_success "Go $(go version | awk '{print $3}') installed via tarball"
                        return 0
                    fi
                fi

                print_soft_error "Go tarball installation failed"
                return 1
            fi
            ;;

        macos)
            if command -v brew >/dev/null 2>&1; then
                print_info "Installing/upgrading Go via Homebrew..."
                # brew install is a no-op if already present; brew upgrade handles the rest.
                run_cmd_arr "brew install go" brew install go
                run_cmd_arr "brew upgrade go" brew upgrade go

                local final_ver
                final_ver=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
                if version_ge "$final_ver" "$GO_REQUIRED"; then
                    print_success "Go $final_ver installed via Homebrew"
                    return 0
                fi

                print_soft_error "Go $final_ver via Homebrew is still below $GO_REQUIRED"
                return 1
            fi

            print_soft_error "Homebrew not found — cannot install Go on macOS without it"
            return 1
            ;;

        windows)
            print_warning "Go $GO_REQUIRED or higher is required. Please install manually from https://go.dev/dl/"
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
            if run_cmd_arr "Installing Deno for Windows" \
                powershell -Command "irm https://deno.land/install.ps1 | iex"; then
                update_path "$HOME/.deno/bin" "Deno"
                print_success "Deno installed"
                return 0
            fi
        fi
    else
        local deno_install_script="/tmp/deno_install.sh"
        if download_file "https://deno.land/install.sh" "$deno_install_script"; then
            chmod +x "$deno_install_script"
            if run_cmd_arr "Running Deno installer" sh "$deno_install_script"; then
                rm -f "$deno_install_script"
                update_path "$HOME/.deno/bin" "Deno"
                if command -v deno >/dev/null 2>&1 || [[ -x "$HOME/.deno/bin/deno" ]]; then
                    print_success "Deno installed"
                    return 0
                fi
            fi
            rm -f "$deno_install_script"
        fi
    fi

    print_soft_error "Deno installation failed"
    return 1
}

check_install_ffmpeg() {
    should_install ffmpeg || return 0

    print_step "Checking FFmpeg..."

    if command -v ffmpeg >/dev/null 2>&1; then
        local ver
        ver=$(ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}')
        print_success "FFmpeg installed ($ver)"
        return 0
    fi

    print_warning "FFmpeg not found, installing..."

    if [[ "$OS_TYPE" == "windows" ]]; then
        local url="https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
        local zip="/tmp/ffmpeg.zip"
        local install_dir="$HOME/ffmpeg"

        if download_file "$url" "$zip"; then
            command -v unzip >/dev/null 2>&1 || install_package "unzip" "unzip"
            mkdir -p "$install_dir"
            if run_cmd_arr "Extracting FFmpeg" unzip -q "$zip" -d /tmp/; then
                local extracted
                extracted=$(find /tmp -maxdepth 1 -type d -name "ffmpeg-*-essentials_build" | head -n1)
                if [[ -n "$extracted" ]]; then
                    cp -r "$extracted/bin/"* "$install_dir/"
                    update_path "$install_dir" "FFmpeg"
                    rm -rf "$extracted" "$zip"
                    command -v ffmpeg >/dev/null 2>&1 && { print_success "FFmpeg installed"; return 0; }
                fi
            fi
            rm -f "$zip"
        fi
    fi

    install_package "ffmpeg" "FFmpeg" && return 0

    if [[ "$OS_TYPE" == "linux" ]] && command -v yum >/dev/null 2>&1; then
        run_as_root "Installing EPEL" yum install -y epel-release
        install_package "ffmpeg" "FFmpeg" && return 0
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

    print_warning "yt-dlp not found, installing..."

    [[ -z "$PYTHON_CMD" ]] && check_install_python

    if [[ -n "$PYTHON_CMD" ]]; then
        if run_cmd_arr "Installing yt-dlp via pip" "$PYTHON_CMD" -m pip install -U yt-dlp; then
            update_path "$HOME/.local/bin" "local bin"
            command -v yt-dlp >/dev/null 2>&1 && { print_success "yt-dlp installed via pip"; return 0; }
        fi

        # Python 3.11+ externally-managed environments
        if run_cmd_arr "Installing yt-dlp (--break-system-packages)" \
            "$PYTHON_CMD" -m pip install -U yt-dlp --break-system-packages; then
            update_path "$HOME/.local/bin" "local bin"
            command -v yt-dlp >/dev/null 2>&1 && { print_success "yt-dlp installed via pip"; return 0; }
        fi
    fi

    local url=""
    case "$OS_TYPE" in
        linux)
            if [[ "$ARCH_TYPE" == "amd64" ]]; then
                url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"
            else
                url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64"
            fi
            ;;
        macos)   url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos" ;;
        windows) url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe" ;;
    esac

    if [[ -n "$url" ]] && download_file "$url" "/tmp/yt-dlp"; then
        chmod +x /tmp/yt-dlp
        if run_as_root "Installing yt-dlp to /usr/local/bin" mv /tmp/yt-dlp /usr/local/bin/yt-dlp; then
            print_success "yt-dlp binary installed to /usr/local/bin"
            return 0
        else
            mkdir -p "$HOME/.local/bin"
            mv /tmp/yt-dlp "$HOME/.local/bin/yt-dlp"
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

    print_step "Installing ntgcalls..."

    command -v unzip >/dev/null 2>&1 || install_package "unzip" "unzip" || \
        print_error "unzip is required to install ntgcalls" "install unzip and re-run the installer"

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
                print_error "ntgcalls unavailable for macOS x86_64" "build from source"
            fi
            ;;
        windows)
            url="https://github.com/pytgcalls/ntgcalls/releases/download/$NTGCALLS_VERSION/ntgcalls.windows-x86_64-static_libs.zip"
            ;;
    esac

    [[ -z "$url" ]] && print_error "Could not determine ntgcalls URL" "check system compatibility"

    if download_file "$url" "ntgcalls.zip"; then
        mkdir -p tmp_ntg
        if run_cmd_arr "Extracting ntgcalls" unzip -q ntgcalls.zip -d tmp_ntg; then
            mkdir -p ntgcalls
            cp tmp_ntg/include/ntgcalls.h ntgcalls/
            local lib_file
            lib_file=$(find tmp_ntg/lib -type f | head -n1)
            if [[ -n "$lib_file" ]]; then
                mv "$lib_file" "./$(basename "$lib_file")"
                print_success "ntgcalls installed"
                rm -rf ntgcalls.zip tmp_ntg
                return 0
            fi
        fi
        rm -rf ntgcalls.zip tmp_ntg
    fi

    print_soft_error "ntgcalls installation failed"
    return 1
}

cleanup_temp_files() {
    print_step "Cleaning up temporary files..."
    rm -f /tmp/go*.tar.gz /tmp/go*.zip /tmp/ffmpeg.zip /tmp/yt-dlp /tmp/get-pip.py 2>/dev/null
    print_success "Cleanup complete"
}

print_summary() {
    [[ "$QUIET_MODE"    == true ]] && return
    [[ "$SKIP_SUMMARY"  == true ]] && return

    echo -e "\n${BLUE}${BOLD}==================================================${RESET}"
    echo -e "${BLUE}${BOLD}             INSTALLATION SUMMARY               ${RESET}"
    echo -e "${BLUE}${BOLD}==================================================${RESET}"
    printf "${CYAN}${BOLD}%-12s | %-12s | %-15s${RESET}\n" "Component" "Status" "Version"
    echo -e "--------------------------------------------------"

    local comp status ver color

    _check_go()     { command -v go     >/dev/null 2>&1; }
    _check_deno()   { command -v deno   >/dev/null 2>&1; }
    _check_ffmpeg() { command -v ffmpeg >/dev/null 2>&1; }
    _check_ytdlp()  { command -v yt-dlp >/dev/null 2>&1; }

    _ver_go()     { go version     2>/dev/null | head -n1 | awk '{print $3}' | sed 's/go//'; }
    _ver_deno()   { deno --version 2>/dev/null | head -n1 | awk '{print $2}'; }
    _ver_ffmpeg() { ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}'; }
    _ver_ytdlp()  { yt-dlp --version 2>/dev/null | head -n1; }

    _print_row() {
        local label="$1" key="$2" check_fn="$3" ver_fn="$4"
        if should_install "$key"; then
            if "$check_fn" >/dev/null 2>&1; then
                status="Installed"; ver=$("$ver_fn"); color=$GREEN
            else
                status="Failed"; ver="-"; color=$RED
            fi
        else
            status="Skipped"; ver="-"; color=$YELLOW
        fi
        printf "%-12s | ${color}%-12s${RESET} | %-15s\n" "$label" "$status" "$ver"
    }

    _print_row "Go"     "go"     _check_go     _ver_go
    _print_row "Deno"   "deno"   _check_deno   _ver_deno
    _print_row "FFmpeg" "ffmpeg" _check_ffmpeg _ver_ffmpeg
    _print_row "yt-dlp" "ytdlp"  _check_ytdlp  _ver_ytdlp

    comp="Python"
    if should_install python; then
        if [[ -n "$PYTHON_CMD" ]] && command -v "$PYTHON_CMD" >/dev/null 2>&1; then
            status="Installed"
            ver=$("$PYTHON_CMD" --version 2>&1 | awk '{print $2}')
            color=$GREEN
        else
            status="Failed"; ver="-"; color=$RED
        fi
    else
        status="Skipped"; ver="-"; color=$YELLOW
    fi
    printf "%-12s | ${color}%-12s${RESET} | %-15s\n" "$comp" "$status" "$ver"

    comp="ntgcalls"
    if should_install ntgcalls; then
        if [[ -f "ntgcalls/ntgcalls.h" ]]; then
            status="Installed"; ver="$NTGCALLS_VERSION"; color=$GREEN
        else
            status="Failed"; ver="-"; color=$RED
        fi
    else
        status="Skipped"; ver="-"; color=$YELLOW
    fi
    printf "%-12s | ${color}%-12s${RESET} | %-15s\n" "$comp" "$status" "$ver"

    echo -e "--------------------------------------------------"
    echo -e "\n${CYAN}Detailed log: $INSTALL_LOG${RESET}"
}

main() {
    parse_arguments "$@"

    log_info "Installation started with arguments: $*"

    detect_system

    print_step "Starting installation..."
    [[ "$QUIET_MODE" == false ]] && echo -e "${CYAN}Installing selected components...${RESET}\n"

    ensure_download_tool
    check_install_deno
    check_install_python
    check_install_pip
    check_install_go
    check_install_ffmpeg
    check_install_ytdlp
    install_ntgcalls

    cleanup_temp_files
    reload_shell_if_needed
    print_summary

    log_info "Installation completed with $WARNINGS warning(s)"

    if [[ $WARNINGS -eq 0 ]]; then
        [[ "$QUIET_MODE" == false ]] && echo -e "\n${GREEN}${BOLD}✓ Installation completed successfully! ✓${RESET}"
        exit 0
    else
        echo -e "\n${YELLOW}${BOLD}⚠ Installation completed with $WARNINGS warning(s) ⚠${RESET}"
        echo -e "${YELLOW}⚠ Some components failed. Check messages above.${RESET}"
        [[ "$QUIET_MODE" == false ]] && echo -e "${YELLOW}⚠ Review log: $INSTALL_LOG${RESET}\n"
        exit 1
    fi
}

main "$@"
