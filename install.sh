#!/bin/sh
set -e

# Helper function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

echo "========================================================"
echo " Starting Password Strength Analyzer Installer"
echo "========================================================"

# 1. Evaluate Host Kernel Target
OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS_RAW" in
  linux*)   OS="linux" ;;
  darwin*)  OS="darwin" ;;
  msys*|mingw*|cygwin*)  OS="windows" ;;
  *)        OS="$OS_RAW" ;;
esac

# 2. Evaluate Hardware Core Architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Fatal Error: Unsupported Architecture ($ARCH)"; exit 1 ;;
esac

echo "Detected System: ${OS} (${ARCH})"

# 3. Check and Install Download Dependencies (curl/wget)
DOWNLOADER=""
if command_exists curl; then
  DOWNLOADER="curl"
elif command_exists wget; then
  DOWNLOADER="wget"
else
  echo "Warning: Neither 'curl' nor 'wget' was found. Attempting to install downloader..."
  if [ "$OS" = "linux" ]; then
    if command_exists apt-get; then
      echo "Detected Debian/Ubuntu. Installing curl via apt-get..."
      sudo apt-get update && sudo apt-get install -y curl
      DOWNLOADER="curl"
    elif command_exists yum; then
      echo "Detected RHEL/CentOS. Installing curl via yum..."
      sudo yum install -y curl
      DOWNLOADER="curl"
    elif command_exists dnf; then
      echo "Detected Fedora/RHEL. Installing curl via dnf..."
      sudo dnf install -y curl
      DOWNLOADER="curl"
    elif command_exists pacman; then
      echo "Detected Arch Linux. Installing curl via pacman..."
      sudo pacman -S --noconfirm curl
      DOWNLOADER="curl"
    fi
  elif [ "$OS" = "darwin" ]; then
    if command_exists brew; then
      echo "Detected macOS. Installing curl via Homebrew..."
      brew install curl
      DOWNLOADER="curl"
    fi
  fi

  if [ -z "$DOWNLOADER" ]; then
    echo "Fatal Error: Could not install downloader. Please install curl or wget manually."
    exit 1
  fi
fi

# 4. Try to Download Binary
SUFX=""
if [ "$OS" = "windows" ]; then
  SUFX=".exe"
fi
BINARY_NAME="pass_analy_${OS}_${ARCH}${SUFX}"
RELEASE_URL="https://github.com/production_repo/releases/download/v1.0.0/${BINARY_NAME}"
DOWNLOAD_SUCCESS=0

echo "Attempting to download pre-compiled binary: ${BINARY_NAME}..."
if [ "$DOWNLOADER" = "curl" ]; then
  if curl -sSfL -o "pass_analy${SUFX}" "${RELEASE_URL}"; then
    DOWNLOAD_SUCCESS=1
  fi
elif [ "$DOWNLOADER" = "wget" ]; then
  if wget -q -O "pass_analy${SUFX}" "${RELEASE_URL}"; then
    DOWNLOAD_SUCCESS=1
  fi
fi

# 5. Fallback: If download fails, compile from source
if [ "$DOWNLOAD_SUCCESS" -ne 1 ]; then
  echo "Warning: Download of pre-compiled binary failed (or release is unavailable)."
  echo "Attempting compilation from source..."

  # Check for Go installation
  if ! command_exists go; then
    echo "Warning: Go compiler ('go') not found. Attempting to install Go compiler..."
    if [ "$OS" = "linux" ]; then
      if command_exists apt-get; then
        echo "Installing Go via apt-get..."
        sudo apt-get update && sudo apt-get install -y golang-go
      elif command_exists yum; then
        echo "Installing Go via yum..."
        sudo yum install -y golang
      elif command_exists dnf; then
        echo "Installing Go via dnf..."
        sudo dnf install -y golang
      elif command_exists pacman; then
        echo "Installing Go via pacman..."
        sudo pacman -S --noconfirm go
      else
        echo "Fatal Error: Unsupported package manager. Please install Go compiler manually."
        exit 1
      fi
    elif [ "$OS" = "darwin" ]; then
      if command_exists brew; then
        echo "Installing Go via Homebrew..."
        brew install go
      else
        echo "Fatal Error: Homebrew not found. Please install Go or Homebrew manually."
        exit 1
      fi
    else
      echo "Fatal Error: Go compiler is required to build from source on Windows. Please install Go."
      exit 1
    fi
  fi

  # Check again if go exists now
  if command_exists go; then
    echo "Go compiler detected. Preparing dependencies..."
    go mod tidy
    echo "Compiling password analyzer..."
    go build -ldflags="-s -w" -o "pass_analy${SUFX}" .
    echo "Compilation successful!"
  else
    echo "Fatal Error: Go installation failed or could not be found."
    exit 1
  fi
fi

# 6. Make Executable and Run Self-Test
chmod +x "pass_analy${SUFX}"
echo "Running post-installation checks..."
if echo "check_test_123" | "./pass_analy${SUFX}" --mode-static > /dev/null 2>&1; then
  echo "Self-test passed successfully!"
else
  echo "Fatal Error: Post-installation self-test failed."
  exit 1
fi

# 7. Global Deployment
GLOBAL_DIR="/usr/local/bin"
if [ ! -d "$GLOBAL_DIR" ]; then
  # Fallback for minimal environments (like Git Bash on Windows)
  if [ -d "/usr/bin" ]; then
    GLOBAL_DIR="/usr/bin"
  else
    GLOBAL_DIR="."
  fi
fi

if [ "$GLOBAL_DIR" != "." ]; then
  echo "Installing globally to: $GLOBAL_DIR..."
  if [ -w "$GLOBAL_DIR" ]; then
    cp "pass_analy${SUFX}" "$GLOBAL_DIR/pass_analy${SUFX}"
    echo "Successfully installed: $GLOBAL_DIR/pass_analy${SUFX}"
  else
    echo "Administrative permissions required. Copying with sudo..."
    if command_exists sudo; then
      sudo cp "pass_analy${SUFX}" "$GLOBAL_DIR/pass_analy${SUFX}"
      echo "Successfully installed: $GLOBAL_DIR/pass_analy${SUFX}"
    else
      echo "Warning: 'sudo' not found. Could not copy to $GLOBAL_DIR automatically."
      echo "Please manually copy pass_analy${SUFX} to your system PATH."
    fi
  fi
else
  echo "Could not find a standard global binary folder (/usr/local/bin or /usr/bin)."
  echo "The binary is available locally as: ./pass_analy${SUFX}"
fi

echo "========================================================"
echo " Deployment Complete!"
echo "========================================================"
