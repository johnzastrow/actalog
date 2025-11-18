#!/bin/bash

################################################################################
# ActaLog Build Script for Fresh Ubuntu Installation
################################################################################
# This script sets up a fresh Ubuntu system with all dependencies needed to
# build and run ActaLog (CrossFit workout tracker).
#
# What this script does:
# 1. Updates system packages
# 2. Installs Go (backend language)
# 3. Installs Node.js and npm (frontend dependencies)
# 4. Installs SQLite3 (default database)
# 5. Installs development tools (make, git, etc.)
# 6. Installs Go development tools (air, golangci-lint, goimports)
# 7. Builds the backend application
# 8. Builds the frontend application
# 9. Creates environment configuration files
# 10. Sets up the database
#
# Usage:
#   chmod +x build.sh
#   ./build.sh
#
# Requirements:
#   - Fresh Ubuntu 20.04+ installation
#   - sudo privileges
#   - Internet connection
################################################################################

set -e  # Exit immediately if a command exits with a non-zero status
set -u  # Treat unset variables as an error

################################################################################
# Configuration Variables
################################################################################

GO_VERSION="1.25.0"  # Go version to install (adjust as needed)
NODE_VERSION="24"     # Node.js major version (LTS)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

################################################################################
# Helper Functions
################################################################################

# Print colored status messages
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

################################################################################
# Step 1: Update System Packages
################################################################################

update_system() {
    print_status "Updating system packages..."
    sudo apt-get update -qq
    sudo apt-get upgrade -y -qq
    print_success "System packages updated"
}

################################################################################
# Step 2: Install Essential Build Tools
################################################################################

install_build_tools() {
    print_status "Installing essential build tools (make, git, curl, wget)..."
    sudo apt-get install -y -qq \
        build-essential \
        make \
        git \
        curl \
        wget \
        ca-certificates \
        gnupg \
        lsb-release
    print_success "Build tools installed"
}

################################################################################
# Step 3: Install Go
################################################################################

install_go() {
    if command_exists go; then
        INSTALLED_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_warning "Go is already installed (version: $INSTALLED_GO_VERSION)"

        # Ask user if they want to reinstall
        read -p "Do you want to reinstall Go $GO_VERSION? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_status "Skipping Go installation"
            return
        fi
    fi

    print_status "Installing Go $GO_VERSION..."

    # Download Go binary
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"

    # Remove any existing Go installation
    sudo rm -rf /usr/local/go

    # Extract Go to /usr/local
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"

    # Add Go to PATH if not already present
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
    fi

    # Set PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    export PATH=$PATH:$HOME/go/bin

    # Clean up
    rm "go${GO_VERSION}.linux-amd64.tar.gz"

    print_success "Go $GO_VERSION installed successfully"
    go version
}

################################################################################
# Step 4: Install Node.js and npm
################################################################################

install_nodejs() {
    if command_exists node; then
        INSTALLED_NODE_VERSION=$(node --version)
        print_warning "Node.js is already installed (version: $INSTALLED_NODE_VERSION)"

        # Check if it's the right major version
        NODE_MAJOR=$(node --version | cut -d'.' -f1 | sed 's/v//')
        if [ "$NODE_MAJOR" -eq "$NODE_VERSION" ]; then
            print_status "Correct Node.js version already installed, skipping..."
            return
        fi
    fi

    print_status "Installing Node.js $NODE_VERSION (LTS)..."

    # Add NodeSource repository
    curl -fsSL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | sudo -E bash -

    # Install Node.js and npm
    sudo apt-get install -y -qq nodejs

    print_success "Node.js and npm installed successfully"
    node --version
    npm --version
}

################################################################################
# Step 5: Install SQLite3
################################################################################

install_sqlite() {
    if command_exists sqlite3; then
        print_warning "SQLite3 is already installed"
        sqlite3 --version
        return
    fi

    print_status "Installing SQLite3..."
    sudo apt-get install -y -qq sqlite3 libsqlite3-dev
    print_success "SQLite3 installed successfully"
    sqlite3 --version
}

################################################################################
# Step 6: Install Go Development Tools
################################################################################

install_go_tools() {
    print_status "Installing Go development tools..."

    # Ensure GOPATH is set
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin

    # Install air (hot reload for Go)
    if ! command_exists air; then
        print_status "Installing air (hot reload)..."
        go install github.com/air-verse/air@latest
    else
        print_warning "air is already installed"
    fi

    # Install golangci-lint (linter)
    if ! command_exists golangci-lint; then
        print_status "Installing golangci-lint (linter)..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
    else
        print_warning "golangci-lint is already installed"
    fi

    # Install goimports (import formatter)
    if ! command_exists goimports; then
        print_status "Installing goimports (import formatter)..."
        go install golang.org/x/tools/cmd/goimports@latest
    else
        print_warning "goimports is already installed"
    fi

    print_success "Go development tools installed"
}

################################################################################
# Step 7: Install Backend Dependencies
################################################################################

install_backend_deps() {
    print_status "Installing backend Go dependencies..."
    cd "$SCRIPT_DIR"

    # Download and tidy Go modules
    go mod download
    go mod tidy

    print_success "Backend dependencies installed"
}

################################################################################
# Step 8: Install Frontend Dependencies
################################################################################

install_frontend_deps() {
    print_status "Installing frontend npm dependencies..."
    cd "$SCRIPT_DIR/web"

    # Clean install (removes node_modules and package-lock.json if they exist)
    if [ -d "node_modules" ]; then
        print_warning "Removing existing node_modules..."
        rm -rf node_modules
    fi

    if [ -f "package-lock.json" ]; then
        print_warning "Removing existing package-lock.json..."
        rm -f package-lock.json
    fi

    # Install dependencies
    npm install

    print_success "Frontend dependencies installed"
    cd "$SCRIPT_DIR"
}

################################################################################
# Step 9: Create Environment Configuration Files
################################################################################

setup_env_files() {
    print_status "Setting up environment configuration files..."

    # Backend .env file
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        if [ -f "$SCRIPT_DIR/.env.example" ]; then
            print_status "Creating .env from .env.example..."
            cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

            # Generate a random JWT secret
            JWT_SECRET=$(openssl rand -base64 32)

            # Update JWT_SECRET in .env
            if command_exists sed; then
                # macOS and Linux compatible sed
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    sed -i '' "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$SCRIPT_DIR/.env"
                else
                    sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" "$SCRIPT_DIR/.env"
                fi
                print_success "Generated random JWT_SECRET"
            else
                print_warning "Could not auto-generate JWT_SECRET, please update .env manually"
            fi

            print_success "Backend .env file created"
        else
            print_warning ".env.example not found, skipping .env creation"
        fi
    else
        print_warning "Backend .env file already exists, skipping..."
    fi

    # Frontend .env file (optional for production)
    if [ ! -f "$SCRIPT_DIR/web/.env" ]; then
        if [ -f "$SCRIPT_DIR/web/.env.example" ]; then
            print_status "Creating web/.env from .env.example..."
            cp "$SCRIPT_DIR/web/.env.example" "$SCRIPT_DIR/web/.env"
            print_success "Frontend .env file created"
        fi
    else
        print_warning "Frontend .env file already exists, skipping..."
    fi
}

################################################################################
# Step 10: Build Backend Application
################################################################################

build_backend() {
    print_status "Building backend application..."
    cd "$SCRIPT_DIR"

    # Run the build using Makefile
    make build

    print_success "Backend built successfully (binary: ./bin/actalog)"
}

################################################################################
# Step 11: Build Frontend Application
################################################################################

build_frontend() {
    print_status "Building frontend application..."
    cd "$SCRIPT_DIR/web"

    # Build for production
    npm run build

    print_success "Frontend built successfully (output: ./web/dist)"
    cd "$SCRIPT_DIR"
}

################################################################################
# Step 12: Initialize Database
################################################################################

init_database() {
    print_status "Initializing database..."
    cd "$SCRIPT_DIR"

    # Check if database already exists
    if [ -f "actalog.db" ]; then
        print_warning "Database file (actalog.db) already exists"
        read -p "Do you want to remove it and create a fresh database? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm actalog.db
            print_status "Removed existing database"
        else
            print_status "Keeping existing database"
            return
        fi
    fi

    # Database will be created automatically when the app runs
    # Migrations will run on first startup
    print_success "Database will be initialized on first run"
}

################################################################################
# Step 13: Create Run Script
################################################################################

create_run_script() {
    print_status "Creating convenience run script..."

    cat > "$SCRIPT_DIR/run.sh" << 'EOF'
#!/bin/bash
# Convenience script to run ActaLog in production mode

# Navigate to project directory
cd "$(dirname "$0")"

# Export Go paths
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

# Run the application
echo "Starting ActaLog..."
echo "Backend will run on http://localhost:8080"
echo "Press Ctrl+C to stop"
echo ""

./bin/actalog
EOF

    chmod +x "$SCRIPT_DIR/run.sh"
    print_success "Created run.sh script"
}

################################################################################
# Step 14: Print Summary
################################################################################

print_summary() {
    echo ""
    echo "================================================================================"
    print_success "ActaLog build completed successfully!"
    echo "================================================================================"
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Review configuration:"
    echo "   ${BLUE}nano .env${NC} (edit JWT_SECRET, DB settings, etc.)"
    echo ""
    echo "2. Start the backend:"
    echo "   ${BLUE}./run.sh${NC}          (production mode)"
    echo "   ${BLUE}make run${NC}          (via Makefile)"
    echo "   ${BLUE}make dev${NC}          (development mode with hot reload)"
    echo ""
    echo "3. For development with frontend:"
    echo "   Terminal 1: ${BLUE}make dev${NC}              (backend with hot reload)"
    echo "   Terminal 2: ${BLUE}cd web && npm run dev${NC} (frontend dev server)"
    echo ""
    echo "4. Access the application:"
    echo "   Backend API:  ${GREEN}http://localhost:8080${NC}"
    echo "   Frontend Dev: ${GREEN}http://localhost:3000${NC} (if running npm run dev)"
    echo "   Health Check: ${GREEN}http://localhost:8080/health${NC}"
    echo ""
    echo "5. First user registration:"
    echo "   The first user to register will automatically become an admin"
    echo ""
    echo "6. Run tests:"
    echo "   ${BLUE}make test${NC}         (all tests with coverage)"
    echo "   ${BLUE}make test-unit${NC}    (unit tests only)"
    echo ""
    echo "7. Other useful commands:"
    echo "   ${BLUE}make lint${NC}         (run linter)"
    echo "   ${BLUE}make fmt${NC}          (format code)"
    echo "   ${BLUE}make clean${NC}        (clean build artifacts)"
    echo ""
    echo "Documentation:"
    echo "   - Architecture:  docs/ARCHITECTURE.md"
    echo "   - Database:      docs/DATABASE_SCHEMA.md"
    echo "   - Development:   CLAUDE.md"
    echo ""
    echo "================================================================================"
}

################################################################################
# Main Execution
################################################################################

main() {
    echo "================================================================================"
    echo "                    ActaLog Build Script for Ubuntu"
    echo "================================================================================"
    echo ""
    echo "This script will:"
    echo "  1. Update system packages"
    echo "  2. Install Go $GO_VERSION"
    echo "  3. Install Node.js $NODE_VERSION (LTS)"
    echo "  4. Install SQLite3"
    echo "  5. Install development tools"
    echo "  6. Install Go tools (air, golangci-lint, goimports)"
    echo "  7. Install backend dependencies"
    echo "  8. Install frontend dependencies"
    echo "  9. Create environment files"
    echo " 10. Build backend application"
    echo " 11. Build frontend application"
    echo " 12. Initialize database"
    echo ""
    echo "================================================================================"
    echo ""

    # Confirm before proceeding
    read -p "Continue with installation? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Installation cancelled by user"
        exit 0
    fi

    echo ""

    # Execute installation steps
    update_system
    install_build_tools
    install_go
    install_nodejs
    install_sqlite
    install_go_tools
    install_backend_deps
    install_frontend_deps
    setup_env_files
    build_backend
    build_frontend
    init_database
    create_run_script

    # Print summary
    print_summary
}

# Run main function
main
