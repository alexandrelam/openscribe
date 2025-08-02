#!/bin/bash

# Whisper.cpp Model Download Script
# Downloads common Whisper models for use with the whisper.cpp provider

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Model directory
MODELS_DIR="$HOME/whisper-models"
BASE_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

# Model information - using simple arrays for bash 3.2 compatibility
MODEL_KEYS=(
    "tiny-en" "base-en" "small-en" "tiny" "base" "small" "medium" "large-v3"
)

MODEL_FILES=(
    "ggml-tiny.en.bin" "ggml-base.en.bin" "ggml-small.en.bin" 
    "ggml-tiny.bin" "ggml-base.bin" "ggml-small.bin" 
    "ggml-medium.bin" "ggml-large-v3.bin"
)

MODEL_SIZES=(
    "39MB" "148MB" "488MB" "39MB" "148MB" "488MB" "1.5GB" "3.1GB"
)

MODEL_SPEEDS=(
    "Fastest" "Fast" "Medium" "Fastest" "Fast" "Medium" "Slow" "Slowest"
)

MODEL_QUALITY=(
    "Basic" "Good" "Better" "Basic" "Good" "Better" "Great" "Best"
)

MODEL_LANGS=(
    "English only" "English only" "English only" "Multilingual" "Multilingual" "Multilingual" "Multilingual" "Multilingual"
)

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    Whisper Model Downloader                   ║${NC}"
    echo -e "${BLUE}║                   for whisper.cpp provider                    ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

get_model_index() {
    local key=$1
    for i in "${!MODEL_KEYS[@]}"; do
        if [[ "${MODEL_KEYS[$i]}" == "$key" ]]; then
            echo $i
            return 0
        fi
    done
    echo -1
}

show_models() {
    echo -e "${YELLOW}Available Models:${NC}"
    echo ""
    printf "%-12s %-20s %-8s %-8s %-8s %s\n" "Key" "Filename" "Size" "Speed" "Quality" "Languages"
    echo "────────────────────────────────────────────────────────────────────────────"
    
    for i in "${!MODEL_KEYS[@]}"; do
        printf "%-12s %-20s %-8s %-8s %-8s %s\n" \
            "${MODEL_KEYS[$i]}" "${MODEL_FILES[$i]}" "${MODEL_SIZES[$i]}" \
            "${MODEL_SPEEDS[$i]}" "${MODEL_QUALITY[$i]}" "${MODEL_LANGS[$i]}"
    done
    echo ""
}

download_model() {
    local key=$1
    local index=$(get_model_index "$key")
    
    if [[ $index -eq -1 ]]; then
        echo -e "${RED}❌ Unknown model key: $key${NC}"
        return 1
    fi
    
    local filename="${MODEL_FILES[$index]}"
    local size="${MODEL_SIZES[$index]}"
    local speed="${MODEL_SPEEDS[$index]}"
    local quality="${MODEL_QUALITY[$index]}"
    local langs="${MODEL_LANGS[$index]}"
    local url="$BASE_URL/$filename"
    local filepath="$MODELS_DIR/$filename"
    
    echo -e "${BLUE}📥 Downloading $filename ($size)...${NC}"
    echo -e "   Speed: $speed | Quality: $quality | Languages: $langs"
    echo ""
    
    # Check if file already exists
    if [[ -f "$filepath" ]]; then
        echo -e "${YELLOW}⚠️  File already exists: $filepath${NC}"
        read -p "Overwrite? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}⏩ Skipping $filename${NC}"
            return 0
        fi
    fi
    
    # Download with progress bar
    if curl -L --progress-bar -o "$filepath" "$url"; then
        echo -e "${GREEN}✅ Successfully downloaded: $filename${NC}"
        echo -e "   Location: $filepath"
        echo ""
    else
        echo -e "${RED}❌ Failed to download: $filename${NC}"
        rm -f "$filepath"  # Remove partial file
        return 1
    fi
}

interactive_mode() {
    while true; do
        echo -e "${YELLOW}Select models to download (space-separated) or 'q' to quit:${NC}"
        echo -e "${BLUE}Examples:${NC}"
        echo "  small-en base-en     (English models only)"
        echo "  small base           (Multilingual models)"
        echo "  tiny-en small medium (Mixed selection)"
        echo "  all                  (Download all models - ~6GB total)"
        echo ""
        read -p "Enter model keys: " -r selection
        
        if [[ "$selection" == "q" ]]; then
            break
        elif [[ "$selection" == "all" ]]; then
            echo -e "${YELLOW}⚠️  This will download ~6GB of models. Continue? (y/N):${NC}"
            read -p "" -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                for key in "${MODEL_KEYS[@]}"; do
                    download_model "$key"
                done
            fi
        else
            local failed=0
            for key in $selection; do
                if ! download_model "$key"; then
                    failed=$((failed + 1))
                fi
            done
            
            if [[ $failed -eq 0 ]]; then
                echo -e "${GREEN}🎉 All selected models downloaded successfully!${NC}"
            else
                echo -e "${YELLOW}⚠️  $failed model(s) failed to download${NC}"
            fi
        fi
        
        echo ""
        echo -e "${BLUE}Download more models? (y/N):${NC}"
        read -p "" -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            break
        fi
    done
}

show_status() {
    echo -e "${BLUE}📂 Model Directory: $MODELS_DIR${NC}"
    echo ""
    
    if [[ ! -d "$MODELS_DIR" ]]; then
        echo -e "${YELLOW}⚠️  Models directory doesn't exist yet${NC}"
        return
    fi
    
    echo -e "${YELLOW}Currently Downloaded Models:${NC}"
    local found_models=0
    
    for i in "${!MODEL_KEYS[@]}"; do
        local filename="${MODEL_FILES[$i]}"
        if [[ -f "$MODELS_DIR/$filename" ]]; then
            local file_size=$(ls -lh "$MODELS_DIR/$filename" | awk '{print $5}')
            echo -e "${GREEN}✅ $filename ($file_size)${NC}"
            found_models=$((found_models + 1))
        fi
    done
    
    if [[ $found_models -eq 0 ]]; then
        echo -e "${YELLOW}   No models found${NC}"
    else
        echo ""
        echo -e "${GREEN}Total models: $found_models${NC}"
    fi
    echo ""
}

main() {
    print_header
    
    # Create models directory
    mkdir -p "$MODELS_DIR"
    echo -e "${GREEN}📁 Models directory: $MODELS_DIR${NC}"
    echo ""
    
    # Check if whisper-cli is available
    if ! command -v whisper-cli >/dev/null 2>&1; then
        echo -e "${RED}❌ whisper-cli not found!${NC}"
        echo -e "${YELLOW}Install whisper.cpp first:${NC}"
        echo "   macOS: brew install whisper-cpp"
        echo "   Linux: See README.md for build instructions"
        echo ""
        exit 1
    else
        echo -e "${GREEN}✅ whisper-cli found: $(which whisper-cli)${NC}"
        echo ""
    fi
    
    show_status
    show_models
    
    # Parse command line arguments
    if [[ $# -eq 0 ]]; then
        interactive_mode
    else
        case "$1" in
            "list"|"ls")
                show_models
                ;;
            "status")
                show_status
                ;;
            "help"|"-h"|"--help")
                echo "Usage: $0 [command|model-keys...]"
                echo ""
                echo "Commands:"
                echo "  list, ls     Show available models"
                echo "  status       Show downloaded models"
                echo "  help         Show this help"
                echo ""
                echo "Model Keys:"
                show_models
                echo "Examples:"
                echo "  $0                    # Interactive mode"
                echo "  $0 small-en base-en   # Download specific models"
                echo "  $0 all               # Download all models"
                ;;
            "all")
                echo -e "${YELLOW}⚠️  This will download ~6GB of models. Continue? (y/N):${NC}"
                read -p "" -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    for key in "${MODEL_KEYS[@]}"; do
                        download_model "$key"
                    done
                fi
                ;;
            *)
                # Download specified models
                local failed=0
                for key in "$@"; do
                    if ! download_model "$key"; then
                        failed=$((failed + 1))
                    fi
                done
                
                if [[ $failed -eq 0 ]]; then
                    echo -e "${GREEN}🎉 All models downloaded successfully!${NC}"
                else
                    echo -e "${YELLOW}⚠️  $failed model(s) failed to download${NC}"
                    exit 1
                fi
                ;;
        esac
    fi
    
    echo -e "${GREEN}✨ Done! You can now use the whisper.cpp provider in the application.${NC}"
    echo -e "${BLUE}💡 Tip: Select 'Whisper.cpp' provider in Settings → Whisper Provider${NC}"
}

main "$@"