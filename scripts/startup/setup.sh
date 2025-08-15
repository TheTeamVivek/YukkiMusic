#!/bin/bash

RED='\033[1;31m'; GREEN='\033[1;32m'; CYAN='\033[1;36m'; RESET='\033[0m'
REPO="https://github.com/TheTeamVivek/YukkiMusic"
FOLDER="yukki"

if ! command -v sudo &> /dev/null || ! sudo -n true 2>/dev/null; then
  echo -e "${RED}[✖] script requires sudo privileges.${RESET}"
  exit 1
fi

echo -e "${CYAN}[~] Updating package list...${RESET}"
sudo apt update -y >/dev/null 2>&1 || {
  echo -e "${RED}[✖] Failed to update. Please run manually.${RESET}"
  exit 1
}

if apt list --upgradable 2>/dev/null | grep -qv "^Listing..."; then
  echo -ne "${CYAN}[?] Upgrades available. Proceed with upgrade? (y/n): ${RESET}"
  read yn
  if [[ $yn =~ ^[Yy]$|^yes$ ]]; then
    sudo apt upgrade -y >/dev/null 2>&1 && echo -e "${GREEN}[✓] Upgrade complete.${RESET}" || echo -e "${RED}[✖] Upgrade failed.${RESET}"
  else
    echo -e "${CYAN}[!] Upgrade skipped.${RESET}"
  fi
else
  echo -e "${GREEN}[✓] System is already up to date.${RESET}"
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo -e "${RED}[✖] Python is not installed. Installing...${RESET}"
  sudo apt install -y python3 >/dev/null 2>&1 || {
    echo -e "${RED}[✖] Failed to install Python 3.${RESET}"
    exit 1
  }
fi
echo -e "${GREEN}[✓] Python is installed: $(python3 --version 2>/dev/null)${RESET}"

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo -e "${RED}[!] FFmpeg not found. Installing...${RESET}"
  sudo apt install -y ffmpeg >/dev/null 2>&1 && echo -e "${GREEN}[✓] FFmpeg installed.${RESET}" || {
    echo -e "${RED}[✖] Failed to install FFmpeg.${RESET}"
    exit 1
  }
fi

if ! command -v git >/dev/null 2>&1; then
  echo -e "${RED}[!] Git not found. Installing...${RESET}"
  sudo apt install -y git >/dev/null 2>&1 && echo -e "${GREEN}[✓] Git installed.${RESET}" || {
    echo -e "${RED}[✖] Failed to install Git.${RESET}"
    exit 1
  }
fi

if ! python3 -m pip --version >/dev/null 2>&1; then
  echo -e "${RED}[✖] pip not found. Installing...${RESET}"
  sudo apt install -y python3-pip >/dev/null 2>&1 || {
    echo -e "${RED}[✖] Failed to install pip.${RESET}"
    exit 1
  }
fi

echo -e "${CYAN}[~] Upgrading pip...${RESET}"
python3 -m pip install --upgrade pip >/dev/null 2>&1 && echo -e "${GREEN}[✓] pip upgraded.${RESET}" || {
  echo -e "${RED}[✖] pip upgrade failed.${RESET}"
  exit 1
}

if [ ! -d .git ]; then
  if [ -d "$FOLDER" ] && [ ! -d "$FOLDER/.git" ]; then
    echo -e "${CYAN}[~] Removing existing '$FOLDER'...${RESET}"
    rm -rf "$FOLDER"
  fi
  if [ ! -d "$FOLDER/.git" ]; then
    echo -e "${CYAN}[~] Cloning YukkiMusic...${RESET}"
    git clone "$REPO" "$FOLDER" >/dev/null 2>&1 || {
      echo -e "${RED}[✖] Failed to clone repo.${RESET}"
      exit 1
    }
  fi
  cd "$FOLDER" || exit 1
fi

command -v uv >/dev/null 2>&1 || {
  echo -e "${CYAN}[~] Installing uv...${RESET}"
  python3 -m pip install uv >/dev/null 2>&1 || {
    echo -e "${RED}[✖] Failed to install uv.${RESET}"
    exit 1
  }
}

[ -d .venv ] || {
  echo -e "${CYAN}[~] Creating virtual environment...${RESET}"
  uv venv .venv >/dev/null 2>&1 || {
    echo -e "${RED}[✖] Failed to create virtual environment.${RESET}"
    exit 1
  }
}

echo -e "${GREEN}[✓] Activating virtual environment...${RESET}"
source .venv/bin/activate

echo -e "${CYAN}[~] Installing dependencies...${RESET}"
uv pip install -e . >/dev/null 2>&1 || {
  echo -e "${RED}[✖] uv pip install -e . failed.${RESET}"
  exit 1
}

echo -en "${CYAN}[*] Run config setup (scripts/envgen.py)? (y/n): ${RESET}"
read -r yn
if [[ $yn =~ ^[Yy]$ ]]; then
  python3 scripts/envgen.py
fi
