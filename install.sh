#!/bin/bash

# ========================
# COLORS & UI
# ========================
RED="\033[1;31m"
GREEN="\033[1;32m"
YELLOW="\033[1;33m"
BLUE="\033[1;34m"
CYAN="\033[1;36m"
BOLD="\033[1m"
RESET="\033[0m"

WARNINGS=0
SHELL_RELOAD_NEEDED=0
INSTALL_LOG="/tmp/install_$(date +%s).log"

# Installation flags
INSTALL_ALL=true
INSTALL_DENO=false
INSTALL_GO=false
INSTALL_PYTHON=false
INSTALL_PIP=false
INSTALL_FFMPEG=false
INSTALL_YTDLP=false
INSTALL_NTGCALLS=false
INSTALL_LIBS=false
SKIP_SUMMARY=false
QUIET_MODE=false

# ========================
# HELPER FUNCTIONS
# ========================

show_help() {
    cat << EOF
${CYAN}${BOLD}Usage:${RESET} $0 [OPTIONS]

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
  -l, --libs              Install additional libraries only
  -q, --quiet             Quiet mode (minimal output)
  --skip-summary          Skip final summary

${CYAN}${BOLD}Examples:${RESET}
  $0                      # Install everything
  $0 --deno               # Install only Deno
  $0 --ntgcalls           # Install only ntgcalls (useful for Docker)
  $0 --go --ffmpeg        # Install only Go and FFmpeg
  $0 --python --pip       # Install only Python and pip
  $0 --all --quiet        # Install everything in quiet mode

${CYAN}${BOLD}Docker Usage:${RESET}
  RUN chmod +x install.sh && ./install.sh --ntgcalls
  RUN ./install.sh --go --ffmpeg --yt-dlp

EOF
    exit 0
}

print_banner() {
    [[ "$QUIET_MODE" == true ]] && return

    printf "%s\n" "${CYAN}╔════════════════════════════════════════════════════════════╗"
    printf "%s\n" "║           🚀 Application Setup & Installation 🚀           ║"
    printf "%s\n\n" "╚════════════════════════════════════════════════════════════╝${RESET}"
}

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

version_ge() { printf '%s\n%s' "$2" "$1" | sort -V -C; }
mark_shell_reload() { SHELL_RELOAD_NEEDED=1; }

log_info() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1" >> "$INSTALL_LOG"
}

reload_shell_if_needed() {
    if [[ $SHELL_RELOAD_NEEDED -eq 1 ]]; then
        print_info "Reloading shell environment..."
        for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
            if [[ -f "$profile" ]]; then
                # shellcheck disable=SC1090
                source "$profile" 2>/dev/null || true
                print_success "Environment reloaded from $profile"
                break
            fi
        done
        export PATH="/usr/local/go/bin:$HOME/.deno/bin:$HOME/.local/bin:$PATH"
    fi
}

# ========================
# PARSE ARGUMENTS
# ========================
parse_arguments() {
    if [[ $# -eq 0 ]]; then
        return
    fi
    
    INSTALL_ALL=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                ;;
            -a|--all)
                INSTALL_ALL=true
                shift
                ;;
            -g|--go)
                INSTALL_GO=true
                shift
                ;;
            -d|--deno)
                INSTALL_DENO=true
                shift
                ;;
            -p|--python)
                INSTALL_PYTHON=true
                shift
                ;;
            --pip)
                INSTALL_PIP=true
                shift
                ;;
            -f|--ffmpeg)
                INSTALL_FFMPEG=true
                shift
                ;;
            -y|--yt-dlp|--ytdlp)
                INSTALL_YTDLP=true
                shift
                ;;
            -n|--ntgcalls)
                INSTALL_NTGCALLS=true
                shift
                ;;
            -l|--libs)
                INSTALL_LIBS=true
                shift
                ;;
            -q|--quiet)
                QUIET_MODE=true
                shift
                ;;
            --skip-summary)
                SKIP_SUMMARY=true
                shift
                ;;
            *)
                echo -e "${RED}Unknown option: $1${RESET}"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
}

should_install() {
    local component=$1
    
    if [[ "$INSTALL_ALL" == true ]]; then
        return 0
    fi
    
    case $component in
        go) [[ "$INSTALL_GO" == true ]] && return 0 ;;
        deno) [[ "$INSTALL_DENO" == true ]] && return 0 ;;
        python) [[ "$INSTALL_PYTHON" == true ]] && return 0 ;;
        pip) [[ "$INSTALL_PIP" == true ]] && return 0 ;;
        ffmpeg) [[ "$INSTALL_FFMPEG" == true ]] && return 0 ;;
        ytdlp) [[ "$INSTALL_YTDLP" == true ]] && return 0 ;;
        ntgcalls) [[ "$INSTALL_NTGCALLS" == true ]] && return 0 ;;
        libs) [[ "$INSTALL_LIBS" == true ]] && return 0 ;;
    esac
    
    return 1
}

# ========================
# DETECT SYSTEM
# ========================
detect_system() {
    print_step "Detecting system..."
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    
    case "$OS" in
        Linux) OS_TYPE="linux" ;;
        Darwin) OS_TYPE="macos" ;;
        MINGW*|MSYS*|CYGWIN*) OS_TYPE="windows" ;;
        *) print_error "Unsupported OS: $OS" "use Linux/macOS/Windows" ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64) ARCH_TYPE="amd64" ;;
        aarch64|arm64) ARCH_TYPE="arm64" ;;
        *) print_error "Unsupported arch: $ARCH" "use x86_64/arm64" ;;
    esac
    
    print_success "System: $OS_TYPE ($ARCH_TYPE)"
    log_info "Detected system: $OS_TYPE ($ARCH_TYPE)"
}

# ========================
# DOWNLOAD TOOLS
# ========================
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
    
    case "$OS_TYPE" in
        linux)
            for pm in apt yum dnf pacman; do
                if command -v $pm >/dev/null 2>&1; then
                    case $pm in
                        apt) sudo apt update >/dev/null 2>&1 && sudo apt install -y curl >/dev/null 2>&1 ;;
                        yum) sudo yum install -y curl >/dev/null 2>&1 ;;
                        dnf) sudo dnf install -y curl >/dev/null 2>&1 ;;
                        pacman) sudo pacman -Sy --noconfirm curl >/dev/null 2>&1 ;;
                    esac
                    if command -v curl >/dev/null 2>&1; then
                        print_success "curl installed via $pm"
                        DOWNLOAD_TOOL="curl"
                        return 0
                    fi
                fi
            done
            ;;
        macos)
            if command -v brew >/dev/null 2>&1 && brew install curl >/dev/null 2>&1; then
                print_success "curl installed via Homebrew"
                DOWNLOAD_TOOL="curl"
                return 0
            fi
            ;;
    esac
    
    print_error "Could not install curl/wget" "install curl or wget"
}

download_file() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        case "$DOWNLOAD_TOOL" in
            curl) 
                if curl -sSL -o "$output" "$url" 2>/dev/null; then
                    return 0
                fi
                ;;
            wget) 
                if wget -q -O "$output" "$url" 2>/dev/null; then
                    return 0
                fi
                ;;
            *) return 1 ;;
        esac
        
        retry_count=$((retry_count + 1))
        if [ $retry_count -lt $max_retries ]; then
            print_info "Download failed, retrying... ($retry_count/$max_retries)"
            sleep 2
        fi
    done
    
    return 1
}

# ========================
# PACKAGE INSTALLER
# ========================
install_package() {
    local package=$1
    local display_name=${2:-$package}
    
    print_info "Installing $display_name..."
    
    case "$OS_TYPE" in
        linux)
            for pm in apt yum dnf pacman; do
                if command -v $pm >/dev/null 2>&1; then
                    case $pm in
                        apt) 
                            sudo apt update >/dev/null 2>&1 || true
                            if sudo apt install -y "$package" >/dev/null 2>&1; then
                                print_success "$display_name installed via apt"
                                log_info "$display_name installed via apt"
                                return 0
                            fi
                            ;;
                        yum) 
                            if sudo yum install -y "$package" >/dev/null 2>&1; then
                                print_success "$display_name installed via yum"
                                log_info "$display_name installed via yum"
                                return 0
                            fi
                            ;;
                        dnf) 
                            if sudo dnf install -y "$package" >/dev/null 2>&1; then
                                print_success "$display_name installed via dnf"
                                log_info "$display_name installed via dnf"
                                return 0
                            fi
                            ;;
                        pacman) 
                            if sudo pacman -Sy --noconfirm "$package" >/dev/null 2>&1; then
                                print_success "$display_name installed via pacman"
                                log_info "$display_name installed via pacman"
                                return 0
                            fi
                            ;;
                    esac
                fi
            done
            ;;
        macos)
            if command -v brew >/dev/null 2>&1 && brew install "$package" >/dev/null 2>&1; then
                print_success "$display_name installed via Homebrew"
                log_info "$display_name installed via Homebrew"
                return 0
            fi
            ;;
    esac
    
    print_soft_error "Failed to install $display_name"
    return 1
}

# ========================
# COMPONENT INSTALLERS
# ========================
check_install_python() {
    should_install python || return 0
    
    print_step "Checking Python..."
    
    for cmd in python3 python; do
        if command -v $cmd >/dev/null 2>&1; then
            PYTHON_CMD="$cmd"
            PYTHON_VERSION=$($PYTHON_CMD --version 2>&1 | awk '{print $2}')
            
            # Check minimum version (3.8+)
            PYTHON_MAJOR=$(echo "$PYTHON_VERSION" | cut -d. -f1)
            PYTHON_MINOR=$(echo "$PYTHON_VERSION" | cut -d. -f2)
            
            if [[ "$PYTHON_MAJOR" -ge 3 ]] && [[ "$PYTHON_MINOR" -ge 8 ]]; then
                print_success "Python found: $PYTHON_VERSION ($PYTHON_CMD)"
                log_info "Python $PYTHON_VERSION found"
                export PYTHON_CMD
                return 0
            else
                print_warning "Python $PYTHON_VERSION is too old (need 3.8+)"
            fi
        fi
    done
    
    print_warning "Python not found or outdated, installing..."
    
    case "$OS_TYPE" in
        linux) install_package "python3" "Python3" && PYTHON_CMD="python3" ;;
        macos) install_package "python@3.12" "Python" && PYTHON_CMD="python3" ;;
    esac
    
    export PYTHON_CMD
    [[ -n "$PYTHON_CMD" ]] && command -v "$PYTHON_CMD" >/dev/null 2>&1 && return 0
    return 1
}

check_install_pip() {
    should_install pip || return 0
    
    print_step "Checking pip..."
    
    [[ -z "$PYTHON_CMD" ]] && { print_warning "Python unavailable, skipping pip"; return 1; }
    
    if $PYTHON_CMD -m pip --version >/dev/null 2>&1; then
        PIP_VERSION=$($PYTHON_CMD -m pip --version 2>&1 | awk '{print $2}')
        print_success "pip installed ($PIP_VERSION)"
        return 0
    fi
    
    print_warning "pip not found, installing..."
    
    case "$OS_TYPE" in
        linux) 
            if install_package "python3-pip" "pip"; then
                log_info "pip installed via package manager"
                return 0
            fi
            ;;
        macos) 
            if $PYTHON_CMD -m ensurepip >/dev/null 2>&1; then
                print_success "pip installed via ensurepip"
                log_info "pip installed via ensurepip"
                return 0
            fi
            ;;
    esac
    
    print_info "Trying get-pip.py..."
    if download_file "https://bootstrap.pypa.io/get-pip.py" "/tmp/get-pip.py"; then
        if $PYTHON_CMD /tmp/get-pip.py >/dev/null 2>&1; then
            rm -f /tmp/get-pip.py
            print_success "pip installed via get-pip.py"
            log_info "pip installed via get-pip.py"
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
    
    # Single Go version - always 1.25
    GO_VERSION="1.25.5"
    
    if command -v go >/dev/null 2>&1; then
        CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Found Go $CURRENT_GO"
        
        # Check if current version is 1.25+
        if version_ge "$CURRENT_GO" "1.25"; then
            print_success "Go $CURRENT_GO (sufficient)"
            log_info "Go $CURRENT_GO already installed"
            return 0
        fi
        print_warning "Go $CURRENT_GO < required 1.25, upgrading..."
    else
        print_warning "Go not installed"
    fi
    
    print_step "Installing Go $GO_VERSION..."
    log_info "Installing Go $GO_VERSION"
    
    case "$OS_TYPE" in
        linux) GO_ARCHIVE="go${GO_VERSION}.${OS_TYPE}-${ARCH_TYPE}.tar.gz" ;;
        macos) GO_ARCHIVE="go${GO_VERSION}.darwin-${ARCH_TYPE}.tar.gz" ;;
        windows) GO_ARCHIVE="go${GO_VERSION}.windows-${ARCH_TYPE}.zip" ;;
    esac
    
    GO_URL="https://go.dev/dl/${GO_ARCHIVE}"
    print_info "Downloading: $GO_URL"
    
    if ! download_file "$GO_URL" "/tmp/${GO_ARCHIVE}"; then
        print_error "Failed to download Go after retries" "download from https://go.dev/dl/"
    fi
    
    print_info "Installing..."
    
    if [[ "$OS_TYPE" == "windows" ]]; then
        command -v unzip >/dev/null 2>&1 || install_package "unzip" "unzip"
        sudo rm -rf /usr/local/go 2>/dev/null || true
        if ! sudo unzip -q "/tmp/${GO_ARCHIVE}" -d /usr/local/ 2>/dev/null; then
            print_error "Extract failed" "extract Go manually"
        fi
    else
        sudo rm -rf /usr/local/go 2>/dev/null || true
        if ! sudo tar -C /usr/local -xzf "/tmp/${GO_ARCHIVE}" 2>/dev/null; then
            print_error "Extract failed" "extract Go manually"
        fi
    fi
    
    if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
        export PATH=$PATH:/usr/local/go/bin
        
        for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
            if [[ -f "$profile" ]] && ! grep -q "/usr/local/go/bin" "$profile" 2>/dev/null; then
                # shellcheck disable=SC2016
                echo 'export PATH=$PATH:/usr/local/go/bin' >> "$profile"
                print_info "Added Go to PATH in $profile"
                mark_shell_reload
                break
            fi
        done
    fi
    
    # Verify installation
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go verification failed" "install Go manually"
    fi
    
    INSTALLED_GO=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $INSTALLED_GO installed"
    log_info "Go $INSTALLED_GO successfully installed"
    
    rm -f "/tmp/${GO_ARCHIVE}"
}

check_install_deno() {
    should_install deno || return 0
    
    print_step "Checking Deno..."
    
    if command -v deno >/dev/null 2>&1; then
        DENO_VERSION=$(deno --version 2>/dev/null | head -n1 | awk '{print $2}')
        print_success "Deno already installed ($DENO_VERSION)"
        log_info "Deno $DENO_VERSION already installed"
        return 0
    fi
    
    print_warning "Deno not found, installing..."
    log_info "Installing Deno"
    
    print_info "Downloading Deno installer..."
    
    if [[ "$OS_TYPE" == "windows" ]]; then
        print_info "Installing Deno for Windows..."
        if command -v powershell >/dev/null 2>&1; then
            if powershell -Command "irm https://deno.land/install.ps1 | iex" 2>/dev/null; then
                DENO_BIN="$HOME/.deno/bin"
                
                sleep 2
                
                if [[ -d "$DENO_BIN" ]]; then
                    export PATH="$DENO_BIN:$PATH"
                    
                    for profile in "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.profile"; do
                        if [[ -f "$profile" ]] && ! grep -q ".deno/bin" "$profile" 2>/dev/null; then
                            echo "export PATH=\"\$HOME/.deno/bin:\$PATH\"" >> "$profile"
                            print_info "Added Deno to PATH in $profile"
                            mark_shell_reload
                            break
                        fi
                    done
                    
                    print_success "Deno installed for Windows"
                    log_info "Deno installed for Windows"
                    return 0
                else
                    print_soft_error "Deno installation directory not found"
                    return 1
                fi
            fi
        else
            print_soft_error "PowerShell not available for Deno installation"
            return 1
        fi
    else
        # Linux/macOS installation
        DENO_INSTALL_SCRIPT="/tmp/deno_install.sh"
        
        if download_file "https://deno.land/install.sh" "$DENO_INSTALL_SCRIPT"; then
            chmod +x "$DENO_INSTALL_SCRIPT"
            
            if sh "$DENO_INSTALL_SCRIPT" >/dev/null 2>&1; then
                rm -f "$DENO_INSTALL_SCRIPT"
                
                DENO_BIN="$HOME/.deno/bin"
                
                export PATH="$DENO_BIN:$PATH"
                
                if ! grep -q ".deno/bin" "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile" 2>/dev/null; then
                    for profile in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.profile"; do
                        if [[ -f "$profile" ]]; then
                            echo "export PATH=\"\$HOME/.deno/bin:\$PATH\"" >> "$profile"
                            print_info "Added Deno to PATH in $profile"
                            mark_shell_reload
                            break
                        fi
                    done
                fi
                
                # Verify installation
                if command -v deno >/dev/null 2>&1 || [[ -x "$DENO_BIN/deno" ]]; then
                    DENO_VERSION=$(deno --version 2>/dev/null | head -n1 | awk '{print $2}' || echo "installed")
                    print_success "Deno $DENO_VERSION installed"
                    log_info "Deno $DENO_VERSION successfully installed"
                    return 0
                fi
            fi
            rm -f "$DENO_INSTALL_SCRIPT"
        fi
    fi
    
    print_soft_error "Deno installation failed"
    return 1
}

check_install_ffmpeg() {
    should_install ffmpeg || return 0
    
    print_step "Checking FFmpeg..."
    
    if command -v ffmpeg >/dev/null 2>&1; then
        FFMPEG_VERSION=$(ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}')
        print_success "FFmpeg installed ($FFMPEG_VERSION)"
        log_info "FFmpeg $FFMPEG_VERSION already installed"
        return 0
    fi
    
    print_warning "FFmpeg not found, installing..."
    log_info "Installing FFmpeg"
    
    if [[ "$OS_TYPE" == "windows" ]]; then
        # Windows-specific FFmpeg installation
        print_info "Downloading FFmpeg for Windows..."
        
        FFMPEG_URL="https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
        FFMPEG_ZIP="/tmp/ffmpeg.zip"
        INSTALL_DIR="/usr/local/ffmpeg"
        
        if download_file "$FFMPEG_URL" "$FFMPEG_ZIP"; then
            print_info "Extracting FFmpeg..."
            command -v unzip >/dev/null 2>&1 || install_package "unzip" "unzip"
            
            sudo mkdir -p "$INSTALL_DIR" 2>/dev/null || mkdir -p "$HOME/ffmpeg"
            if sudo unzip -q "$FFMPEG_ZIP" -d /tmp/ 2>/dev/null; then
                EXTRACTED_DIR=$(find /tmp -maxdepth 1 -type d -name "ffmpeg-*-essentials_build" | head -n1)
                if [[ -n "$EXTRACTED_DIR" ]]; then
                    if sudo cp -r "$EXTRACTED_DIR/bin/"* "$INSTALL_DIR/" 2>/dev/null || cp -r "$EXTRACTED_DIR/bin/"* "$HOME/ffmpeg/" 2>/dev/null; then
                        # Add to PATH
                        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
                            export PATH="$INSTALL_DIR:$PATH"
                            for profile in "$HOME/.bashrc" "$HOME/.bash_profile"; do
                                if [[ -f "$profile" ]] && ! grep -q "$INSTALL_DIR" "$profile" 2>/dev/null; then
                                    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$profile"
                                    mark_shell_reload
                                    break
                                fi
                            done
                        fi
                        
                        rm -rf "$EXTRACTED_DIR" "$FFMPEG_ZIP"
                        if command -v ffmpeg >/dev/null 2>&1; then
                            print_success "FFmpeg installed for Windows"
                            log_info "FFmpeg installed for Windows"
                            return 0
                        fi
                    fi
                fi
            fi
            rm -f "$FFMPEG_ZIP"
        fi
        
        print_soft_error "FFmpeg Windows install failed, trying package manager..."
    fi
    
    if install_package "ffmpeg" "FFmpeg"; then
        if command -v ffmpeg >/dev/null 2>&1; then
            log_info "FFmpeg installed via package manager"
            return 0
        fi
    fi
    
    if [[ "$OS_TYPE" == "linux" ]] && command -v yum >/dev/null 2>&1; then
        print_info "Trying EPEL repository..."
        sudo yum install -y epel-release >/dev/null 2>&1 || true
        if install_package "ffmpeg" "FFmpeg"; then
            if command -v ffmpeg >/dev/null 2>&1; then
                log_info "FFmpeg installed via EPEL"
                return 0
            fi
        fi
    fi
    
    command -v ffmpeg >/dev/null 2>&1 || print_soft_error "FFmpeg install failed"
}

check_install_ytdlp() {
    should_install ytdlp || return 0
    
    print_step "Checking yt-dlp..."
    
    if command -v yt-dlp >/dev/null 2>&1; then
        YTDLP_VERSION=$(yt-dlp --version 2>/dev/null || echo 'unknown')
        print_success "yt-dlp installed ($YTDLP_VERSION)"
        log_info "yt-dlp $YTDLP_VERSION already installed"
        return 0
    fi
    
    print_warning "yt-dlp not found"
    log_info "Installing yt-dlp"
    
    [[ -z "$PYTHON_CMD" ]] && check_install_python
    
    if [[ -n "$PYTHON_CMD" ]] && $PYTHON_CMD -m pip --version >/dev/null 2>&1; then
        print_step "Installing via pip..."
        
        if $PYTHON_CMD -m pip install -U yt-dlp >/dev/null 2>&1; then
            print_success "yt-dlp installed via pip"
            log_info "yt-dlp installed via pip"
        else
            PYTHON_VER=$($PYTHON_CMD --version 2>&1 | awk '{print $2}' | cut -d. -f1,2)
            if version_ge "$PYTHON_VER" "3.11" && $PYTHON_CMD -m pip install -U yt-dlp --break-system-packages >/dev/null 2>&1; then
                print_success "yt-dlp installed (--break-system-packages)"
                log_info "yt-dlp installed with --break-system-packages"
            elif command -v pipx >/dev/null 2>&1 || install_package "pipx" "pipx"; then
                if pipx install yt-dlp >/dev/null 2>&1; then
                    print_success "yt-dlp installed via pipx"
                    log_info "yt-dlp installed via pipx"
                fi
            fi
        fi
        
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            export PATH="$HOME/.local/bin:$PATH"
            for profile in "$HOME/.bashrc" "$HOME/.zshrc"; do
                if [[ -f "$profile" ]] && ! grep -q ".local/bin" "$profile" 2>/dev/null; then
                    # shellcheck disable=SC2016
                    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$profile"
                    mark_shell_reload
                    break
                fi
            done
        fi
        
        if command -v yt-dlp >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    print_step "Trying binary install..."
    
    case "$OS_TYPE" in
        linux)
            [[ "$ARCH_TYPE" == "amd64" ]] && YTDLP_URL="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"
            [[ "$ARCH_TYPE" == "arm64" ]] && YTDLP_URL="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64"
            ;;
        macos) YTDLP_URL="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos" ;;
        windows) YTDLP_URL="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe" ;;
    esac
    
    if [[ -n "$YTDLP_URL" ]] && download_file "$YTDLP_URL" "/tmp/yt-dlp"; then
        if sudo mv /tmp/yt-dlp "/usr/local/bin/yt-dlp" 2>/dev/null && sudo chmod +x "/usr/local/bin/yt-dlp" 2>/dev/null; then
            print_success "Binary installed to /usr/local/bin"
            log_info "yt-dlp binary installed to /usr/local/bin"
        elif mkdir -p "$HOME/.local/bin" && mv /tmp/yt-dlp "$HOME/.local/bin/yt-dlp" 2>/dev/null && chmod +x "$HOME/.local/bin/yt-dlp"; then
            export PATH="$HOME/.local/bin:$PATH"
            print_success "Binary installed to ~/.local/bin"
            log_info "yt-dlp binary installed to ~/.local/bin"
        fi
    fi
    
    reload_shell_if_needed
    command -v yt-dlp >/dev/null 2>&1 || print_soft_error "yt-dlp install failed"
}

install_ntgcalls() {
    should_install ntgcalls || return 0
    
    print_step "Installing ntgcalls..."
    
    VERSION="v2.0.6"
    
    command -v unzip >/dev/null 2>&1 || install_package "unzip" "unzip" || print_error "unzip required" "install unzip"
    
    case "$OS_TYPE" in
        linux)
            [[ "$ARCH_TYPE" == "amd64" ]] && URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.linux-x86_64-static_libs.zip"
            [[ "$ARCH_TYPE" == "arm64" ]] && URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.linux-arm64-static_libs.zip"
            ;;
        macos)
            [[ "$ARCH_TYPE" == "arm64" ]] && URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.macos-arm64-static_libs.zip"
            [[ "$ARCH_TYPE" == "amd64" ]] && print_error "ntgcalls unavailable for macOS x86_64" "use arm64 or build from source"
            ;;
        windows) URL="https://github.com/pytgcalls/ntgcalls/releases/download/$VERSION/ntgcalls.windows-x86_64-static_libs.zip" ;;
    esac
    
    [[ -z "$URL" ]] && print_error "Could not determine ntgcalls URL" "check system compatibility"
    
    print_info "Downloading: $URL"
    if ! download_file "$URL" "ntgcalls.zip"; then
        print_error "Download failed after retries" "download manually from $URL"
    fi
    
    mkdir -p tmp
    if ! unzip -q ntgcalls.zip -d tmp 2>/dev/null; then
        rm -rf ntgcalls.zip tmp
        print_error "Extract failed" "extract manually"
    fi
    
    mkdir -p ntgcalls
    
    if [[ -f "tmp/include/ntgcalls.h" ]]; then
        if ! cp "tmp/include/ntgcalls.h" "ntgcalls/"; then
            rm -rf ntgcalls.zip tmp
            print_error "Copy header failed" "copy manually"
        fi
    else
        rm -rf ntgcalls.zip tmp
        print_error "Header not found" "verify download"
    fi
    
    LIB_FILE=$(find "tmp/lib" -type f 2>/dev/null | head -n1)
    if [[ -n "$LIB_FILE" ]]; then
        LIB_NAME=$(basename "$LIB_FILE")
        if ! mv "$LIB_FILE" "./$LIB_NAME"; then
            rm -rf ntgcalls.zip tmp
            print_error "Move library failed" "move manually"
        fi
    else
        rm -rf ntgcalls.zip tmp
        print_error "Library not found" "verify download"
    fi
    
    rm -rf ntgcalls.zip tmp
    print_success "ntgcalls installed"
    log_info "ntgcalls successfully installed"
}

install_additional_libs() {
    should_install libs || return 0
    
    print_step "Checking additional libraries..."
    
    case "$OS_TYPE" in
        linux) LIBS=("gcc" "zlib1g-dev" "git") ;;
        macos) LIBS=("gcc" "git") ;;
        *) return 0 ;;
    esac
    
    for lib in "${LIBS[@]}"; do
        command -v "${lib%%-dev}" >/dev/null 2>&1 && continue
        dpkg -l | grep -q "^ii  $lib" 2>/dev/null && continue
        rpm -qa | grep -q "$lib" 2>/dev/null && continue
        install_package "$lib" "$lib" || print_soft_error "Could not install $lib"
    done
}

# ========================
# CLEANUP
# ========================
cleanup_temp_files() {
    print_step "Cleaning up temporary files..."
    rm -f /tmp/go*.tar.gz /tmp/go*.zip /tmp/ffmpeg.zip /tmp/yt-dlp /tmp/get-pip.py 2>/dev/null
    print_success "Cleanup complete"
}

# ========================
# FINAL SUMMARY
# ========================
print_summary() {
    [[ "$SKIP_SUMMARY" == true ]] && return
    
    echo -e "\n${CYAN}${BOLD}╔════════════════════════════════════════════════════════════╗${RESET}"
    echo -e "${CYAN}${BOLD}║                    Installation Summary                   ║${RESET}"
    echo -e "${CYAN}${BOLD}╚════════════════════════════════════════════════════════════╝${RESET}\n"
    
    # Check each component
    if should_install go; then
        if command -v go >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Go: $(go version | awk '{print $3}')${RESET}"
        else
            echo -e "${RED}✗ Go: Not installed${RESET}"
        fi
    fi
    
    if should_install deno; then
        if command -v deno >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Deno: $(deno --version 2>/dev/null | head -n1 | awk '{print $2}')${RESET}"
        else
            echo -e "${RED}✗ Deno: Not installed${RESET}"
        fi
    fi

    if should_install ffmpeg; then
        if command -v ffmpeg >/dev/null 2>&1; then
            echo -e "${GREEN}✓ FFmpeg: $(ffmpeg -version 2>/dev/null | head -n1 | awk '{print $3}')${RESET}"
        else
            echo -e "${RED}✗ FFmpeg: Not installed${RESET}"
        fi
    fi
    
    if should_install ytdlp; then
        if command -v yt-dlp >/dev/null 2>&1; then
            echo -e "${GREEN}✓ yt-dlp: $(yt-dlp --version 2>/dev/null)${RESET}"
        else
            echo -e "${RED}✗ yt-dlp: Not installed${RESET}"
        fi
    fi
    
    if should_install python; then
        if [[ -n "$PYTHON_CMD" ]]; then
            echo -e "${GREEN}✓ Python: $($PYTHON_CMD --version 2>&1 | awk '{print $2}')${RESET}"
        else
            echo -e "${YELLOW}⚠ Python: Not found${RESET}"
        fi
    fi
    
    if should_install ntgcalls; then
        if [[ -f "ntgcalls/ntgcalls.h" ]]; then
            echo -e "${GREEN}✓ ntgcalls: Installed${RESET}"
        else
            echo -e "${RED}✗ ntgcalls: Not installed${RESET}"
        fi
    fi
    
    [[ "$QUIET_MODE" == false ]] && echo -e "\n${CYAN}Installation log saved to: $INSTALL_LOG${RESET}"
}

# ========================
# MAIN
# ========================
main() {
    parse_arguments "$@"
    
    print_banner
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
    install_additional_libs
    install_ntgcalls
    
    cleanup_temp_files
    reload_shell_if_needed
    print_summary
    
    log_info "Installation completed with $WARNINGS warnings"
    
    echo -e "\n${GREEN}${BOLD}╔════════════════════════════════════════════════════════════╗${RESET}"
    if [[ $WARNINGS -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}║          ✓ Installation completed successfully! ✓          ║${RESET}"
    else
        echo -e "${YELLOW}${BOLD}║     ⚠ Installation completed with $WARNINGS warning(s) ⚠      ║${RESET}"
    fi
    echo -e "${GREEN}${BOLD}╚════════════════════════════════════════════════════════════╝${RESET}"
    
    if [[ $WARNINGS -gt 0 ]]; then
        echo -e "${YELLOW}⚠ Some components failed. Check messages above.${RESET}"
        [[ "$QUIET_MODE" == false ]] && echo -e "${YELLOW}⚠ Review log file: $INSTALL_LOG${RESET}\n"
        exit 1
    fi
    
    exit 0
}

main "$@"
