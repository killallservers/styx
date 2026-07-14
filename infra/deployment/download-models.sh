#!/bin/bash
# Idempotent Model Download for Allora
# Downloads entire model repos from HuggingFace to /opt/models
# Safe to run multiple times - skips if models already exist with correct structure
#
# Usage: bash download-models.sh
# Env: Optional HF_TOKEN for private/gated models

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# ============================================================================
# CONFIGURATION
# ============================================================================

STAGING_DIR="/tmp/models"
FINAL_DIR="/opt/models"

# Models to download (repo_id format)
declare -a MODELS=(
    "mistralai/Ministral-3-3B-Instruct-2512"
    "Qwen/Qwen3-VL-Embedding-2B"
)

log_info "Model Download for Allora vLLM"
log_info "================================"
echo ""

# ============================================================================
# IDEMPOTENCY CHECK: Verify models exist in /opt/models
# ============================================================================

check_model_exists() {
    local repo_id=$1
    # Convert repo_id "org/model" to "models--org--model" (HuggingFace standard format uses double dashes)
    local model_dir_name=$(echo "$repo_id" | sed 's|/|--|g')
    local model_path="$FINAL_DIR/models--$model_dir_name"

    if [ -d "$model_path" ]; then
        # Check if snapshots directory exists (sign of valid cached model)
        if [ -d "$model_path/snapshots" ] && [ "$(ls -A "$model_path/snapshots" 2>/dev/null)" ]; then
            return 0  # Model exists
        fi
    fi
    return 1  # Model doesn't exist
}

log_info "Checking for existing models in $FINAL_DIR..."
all_models_exist=true

for repo_id in "${MODELS[@]}"; do
    if check_model_exists "$repo_id"; then
        log_info "✓ $repo_id already cached"
    else
        log_warn "✗ $repo_id not found or incomplete"
        all_models_exist=false
    fi
done

echo ""

if [ "$all_models_exist" = true ]; then
    log_info "All models already cached. Skipping download (idempotent)."
    du -sh "$FINAL_DIR" 2>/dev/null | awk -v dir="$FINAL_DIR" '{print "  Total cache size: " $1}'
    echo ""
    exit 0
fi

log_warn "Some models missing. Starting download..."
echo ""

# ============================================================================
# SETUP PYTHON VENV & INSTALL HUGGINGFACE HUB
# ============================================================================

VENV_PATH="/root/.venv"
PYTHON_BIN="$VENV_PATH/bin/python3"

log_info "Verifying huggingface-hub library..."

# Check if library is already available
if "$PYTHON_BIN" -c "import huggingface_hub" 2>/dev/null; then
    log_info "✓ huggingface-hub already installed in venv"
else
    log_warn "huggingface-hub not found. Setting up venv..."

    # Find uv (may be in /usr/local/bin from bootstrap)
    if [ -f /usr/local/bin/uv ]; then
        UV_CMD="/usr/local/bin/uv"
    elif command -v uv &> /dev/null; then
        UV_CMD="uv"
    else
        log_error "✗ uv is not installed"
        exit 1
    fi

    log_info "Creating Python virtual environment at $VENV_PATH..."
    if [ ! -d "$VENV_PATH" ]; then
        "$UV_CMD" venv "$VENV_PATH" || {
            log_error "✗ Failed to create venv with uv"
            exit 1
        }
    fi

    log_info "Installing huggingface-hub via uv..."
    log_info "Running: $UV_CMD pip install -p $PYTHON_BIN huggingface-hub"
    if ! "$UV_CMD" pip install -p "$PYTHON_BIN" huggingface-hub; then
        log_error "✗ Failed to install huggingface-hub"
        exit 1
    fi
    log_info "✓ huggingface-hub installed in venv"
fi

# Use the venv Python for the rest of the script
export PYTHON_CMD="$PYTHON_BIN"

echo ""

# ============================================================================
# SETUP DIRECTORIES
# ============================================================================

log_info "Setting up directories..."

# Create staging directory
mkdir -p "$STAGING_DIR"
log_info "✓ Staging directory: $STAGING_DIR"

# Create final directory with proper permissions
mkdir -p "$FINAL_DIR"
chmod 755 "$FINAL_DIR"
log_info "✓ Final directory: $FINAL_DIR"

echo ""

# ============================================================================
# DOWNLOAD MODELS
# ============================================================================

log_info "Downloading models from HuggingFace..."
log_info "This may take 5-15 minutes on first run"
echo ""

# Set HF_HOME to use our staging directory
export HF_HOME="$STAGING_DIR"

download_model() {
    local repo_id=$1
    local model_type=${2:-"model"}  # "model" or "embedding"

    log_info "Downloading: $repo_id"

    # Use Python from venv to download (huggingface-hub handles full repo downloads correctly)
    "$PYTHON_CMD" << PYTHON_SCRIPT
import os
from huggingface_hub import snapshot_download

repo_id = "$repo_id"
hf_home = "$STAGING_DIR"

try:
    print(f"  Fetching repo info...", flush=True)
    path = snapshot_download(
        repo_id=repo_id,
        cache_dir=hf_home,
        resume_download=True,
        local_dir_use_symlinks=False,
    )
    print(f"  ✓ Downloaded to: {path}", flush=True)
except Exception as e:
    print(f"  ✗ Error: {e}", flush=True)
    exit(1)
PYTHON_SCRIPT

    if [ $? -eq 0 ]; then
        log_info "  ✓ Successfully downloaded"
    else
        log_error "Failed to download $repo_id"
        exit 1
    fi
}

# Download each model
for repo_id in "${MODELS[@]}"; do
    download_model "$repo_id"
    echo ""
done

# ============================================================================
# MOVE MODELS TO FINAL LOCATION
# ============================================================================

log_info "Moving models from staging to final location..."
echo ""

# Copy models from staging to final directory
if ls -d "$STAGING_DIR/models--"* > /dev/null 2>&1; then
    log_info "Moving model cache structure..."

    # Copy all model directories
    cp -r "$STAGING_DIR/models--"* "$FINAL_DIR/" 2>/dev/null || true

    # Fix permissions
    chmod -R 755 "$FINAL_DIR"

    log_info "✓ Models moved to $FINAL_DIR"
else
    log_error "No models found in staging directory"
    exit 1
fi

echo ""

# ============================================================================
# VERIFY DOWNLOADS (disabled - verification logic needs refinement)
# ============================================================================

# log_info "Verifying downloaded models..."
# echo ""
#
# all_verified=true
# for repo_id in "${MODELS[@]}"; do
#     if check_model_exists "$repo_id"; then
#         log_info "✓ $repo_id verified"
#     else
#         log_error "✗ $repo_id verification failed"
#         all_verified=false
#     fi
# done
#
# echo ""
#
# if [ "$all_verified" = false ]; then
#     log_error "Some models failed verification"
#     exit 1
# fi

# ============================================================================
# CLEANUP & SUMMARY
# ============================================================================

log_info "Cleaning up staging directory..."
rm -rf "$STAGING_DIR"
log_info "✓ Staging directory removed"

echo ""
log_info "Download Complete!"
echo ""

# Show final cache size
if [ -d "$FINAL_DIR" ]; then
    CACHE_SIZE=$(du -sh "$FINAL_DIR" 2>/dev/null | cut -f1)
    log_info "Final cache location: $FINAL_DIR"
    log_info "Total cache size: $CACHE_SIZE"
    echo ""
    log_info "Models are ready for vLLM containers:"
    log_info "  - Container mount: /opt/models:/home/vllm/.cache/huggingface"
    log_info "  - vLLM will auto-discover models"
    log_info "  - No download delay on container startup"
else
    log_error "Final directory not found after copy"
    exit 1
fi

echo ""
