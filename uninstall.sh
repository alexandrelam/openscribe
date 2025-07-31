#!/bin/bash

# Vibe Speech-to-Text Global Uninstaller Script
# This script removes the globally installed vibe command

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMMAND_NAME="vibe"
CLI_COMMAND_NAME="vibe-cli"
INSTALL_DIR="$HOME/.local/bin"

# Helper functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_header() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                  Vibe Speech-to-Text Uninstaller             ║"
    echo "║               Remove globally installed 'vibe' command      ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Get absolute path to project directory
get_project_dir() {
    # Get the directory where this script is located
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    echo "$script_dir"
}

# Check installation status
check_installation() {
    print_info "Checking installation status..."
    
    local installed_files=()
    local launcher_path="$INSTALL_DIR/$COMMAND_NAME"
    local cli_launcher_path="$INSTALL_DIR/$CLI_COMMAND_NAME"
    
    # Check main launcher
    if [[ -f "$launcher_path" ]]; then
        installed_files+=("$launcher_path")
        print_info "Found: $launcher_path"
    fi
    
    # Check CLI launcher
    if [[ -f "$cli_launcher_path" ]] || [[ -L "$cli_launcher_path" ]]; then
        installed_files+=("$cli_launcher_path")
        print_info "Found: $cli_launcher_path"
    fi
    
    # Check if command is in PATH
    if command -v "$COMMAND_NAME" &> /dev/null; then
        local command_path
        command_path="$(command -v "$COMMAND_NAME")"
        print_info "Command in PATH: $command_path"
    fi
    
    if [[ ${#installed_files[@]} -eq 0 ]]; then
        print_warning "No Vibe installation found"
        return 1
    fi
    
    print_success "Found ${#installed_files[@]} installed file(s)"
    return 0
}

# Remove installation files
remove_installation() {
    print_info "Removing installation files..."
    
    local removed_count=0
    local launcher_path="$INSTALL_DIR/$COMMAND_NAME"
    local cli_launcher_path="$INSTALL_DIR/$CLI_COMMAND_NAME"
    
    # Remove main launcher
    if [[ -f "$launcher_path" ]]; then
        rm -f "$launcher_path"
        print_success "Removed: $launcher_path"
        ((removed_count++))
    fi
    
    # Remove CLI launcher (symlink or file)
    if [[ -f "$cli_launcher_path" ]] || [[ -L "$cli_launcher_path" ]]; then
        rm -f "$cli_launcher_path"
        print_success "Removed: $cli_launcher_path"
        ((removed_count++))
    fi
    
    if [[ $removed_count -eq 0 ]]; then
        print_warning "No installation files found to remove"
        return 1
    fi
    
    print_success "Removed $removed_count installation file(s)"
    return 0
}

# Ask about config cleanup
ask_config_cleanup() {
    local project_dir="$1"
    local config_file="$project_dir/config.json"
    
    if [[ ! -f "$config_file" ]]; then
        return 0  # No config file to remove
    fi
    
    echo ""
    print_info "Configuration file found: $config_file"
    print_warning "This contains your Vibe settings (language, VAD parameters, etc.)"
    
    while true; do
        read -p "Do you want to remove the configuration file? (y/N): " -n 1 -r
        echo
        case $REPLY in
            [Yy])
                if rm -f "$config_file"; then
                    print_success "Configuration file removed"
                else
                    print_error "Failed to remove configuration file"
                fi
                break
                ;;
            [Nn]|"")
                print_info "Configuration file preserved"
                break
                ;;
            *)
                print_warning "Please answer y or n"
                ;;
        esac
    done
}

# Verify uninstallation
verify_uninstallation() {
    print_info "Verifying uninstallation..."
    
    local issues_found=0
    
    # Check if command still exists in PATH
    if command -v "$COMMAND_NAME" &> /dev/null; then
        local remaining_path
        remaining_path="$(command -v "$COMMAND_NAME")"
        print_warning "$COMMAND_NAME still found in PATH: $remaining_path"
        print_info "This might be a different installation or cached command"
        ((issues_found++))
    fi
    
    # Check for remaining files
    local launcher_path="$INSTALL_DIR/$COMMAND_NAME"
    local cli_launcher_path="$INSTALL_DIR/$CLI_COMMAND_NAME"
    
    if [[ -f "$launcher_path" ]]; then
        print_warning "Launcher file still exists: $launcher_path"
        ((issues_found++))
    fi
    
    if [[ -f "$cli_launcher_path" ]] || [[ -L "$cli_launcher_path" ]]; then
        print_warning "CLI launcher still exists: $cli_launcher_path"
        ((issues_found++))
    fi
    
    if [[ $issues_found -eq 0 ]]; then
        print_success "Uninstallation verified successfully"
        return 0
    else
        print_warning "Uninstallation completed with $issues_found issue(s)"
        print_info "You may need to restart your terminal or clear command cache"
        return 1
    fi
}

# Show completion message
show_completion_message() {
    echo ""
    echo -e "${GREEN}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                   Uninstallation Complete!                  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    echo "🗑️  Vibe has been uninstalled from your system"
    echo ""
    echo "What was removed:"
    echo "  • Global 'vibe' command"
    echo "  • CLI 'vibe-cli' shortcut"
    echo "  • Launcher scripts from ~/.local/bin"
    echo ""
    echo "What was preserved:"
    echo "  • Project source code"
    echo "  • Virtual environment"
    echo "  • Configuration file (unless you chose to remove it)"
    echo ""
    echo "To reinstall: run ${BLUE}./install.sh${NC} from the project directory"
    echo "To completely remove: delete the entire project directory"
}

# Show help
show_help() {
    echo "Vibe Speech-to-Text Uninstaller"
    echo ""
    echo "Usage:"
    echo "  ./uninstall.sh              Interactive uninstallation"
    echo "  ./uninstall.sh --force      Force removal without confirmation"
    echo "  ./uninstall.sh --help       Show this help message"
    echo ""
    echo "This script removes:"
    echo "  • ~/.local/bin/vibe launcher"
    echo "  • ~/.local/bin/vibe-cli symlink"
    echo "  • Optionally: config.json (with confirmation)"
    echo ""
    echo "The project source code and virtual environment are preserved."
}

# Main uninstallation function
main() {
    local force_mode=false
    
    # Parse command line arguments
    case "$1" in
        --help|-h)
            show_help
            exit 0
            ;;
        --force|-f)
            force_mode=true
            ;;
        *)
            if [[ -n "$1" ]]; then
                print_error "Unknown option: $1"
                echo "Use './uninstall.sh --help' for usage information."
                exit 1
            fi
            ;;
    esac
    
    print_header
    
    # Check if installation exists
    if ! check_installation; then
        print_info "Nothing to uninstall"
        exit 0
    fi
    
    # Get confirmation unless force mode
    if [[ "$force_mode" != true ]]; then
        echo ""
        print_warning "This will remove the globally installed 'vibe' command"
        print_info "The project source code will be preserved"
        
        while true; do
            read -p "Do you want to continue? (y/N): " -n 1 -r
            echo
            case $REPLY in
                [Yy])
                    break
                    ;;
                [Nn]|"")
                    print_info "Uninstallation cancelled"
                    exit 0
                    ;;
                *)
                    print_warning "Please answer y or n"
                    ;;
            esac
        done
    fi
    
    # Remove installation
    if ! remove_installation; then
        print_error "Uninstallation failed"
        exit 1
    fi
    
    # Ask about config cleanup (unless force mode)
    if [[ "$force_mode" != true ]]; then
        local project_dir
        project_dir="$(get_project_dir)"
        ask_config_cleanup "$project_dir"
    fi
    
    # Verify uninstallation
    verify_uninstallation
    
    # Show completion message
    show_completion_message
}

# Run main function
main "$@"