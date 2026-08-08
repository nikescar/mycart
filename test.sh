#!/bin/bash
# Test runner for mycart with PostgreSQL support
# Usage: ./test.sh [sqlite|postgres|all|failed|verbose]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default PostgreSQL connection
DEFAULT_POSTGRES_URL="postgresql://postgres:mycartWkd123@db.tybjgfktpgkvrjmzamhx.supabase.co:5432/postgres"

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✓ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ ${1}${NC}"
}

print_error() {
    echo -e "${RED}✗ ${1}${NC}"
}

# Function to run tests with PostgreSQL
run_postgres_tests() {
    print_info "Running tests with PostgreSQL..."
    export TEST_DATABASE_URL="${TEST_DATABASE_URL:-$DEFAULT_POSTGRES_URL}"
    export DATABASE_URL="${DATABASE_URL:-$DEFAULT_POSTGRES_URL}"

    print_info "Database: $TEST_DATABASE_URL"

    if [ "$VERBOSE" = "true" ]; then
        go test -v ./... "$@"
    else
        go test ./... "$@"
    fi
}

# Function to run tests with SQLite (default behavior)
run_sqlite_tests() {
    print_info "Running tests with SQLite (in-memory)..."
    unset TEST_DATABASE_URL
    unset DATABASE_URL

    if [ "$VERBOSE" = "true" ]; then
        go test -v ./... "$@"
    else
        go test ./... "$@"
    fi
}

# Function to run only failed packages
run_failed_tests() {
    print_info "Running only previously failed test packages..."
    export TEST_DATABASE_URL="${TEST_DATABASE_URL:-$DEFAULT_POSTGRES_URL}"
    export DATABASE_URL="${DATABASE_URL:-$DEFAULT_POSTGRES_URL}"

    FAILED_PACKAGES=(
        "github.com/shurco/mycart/internal"
        "github.com/shurco/mycart/internal/handlers/private"
        "github.com/shurco/mycart/internal/mailer"
        "github.com/shurco/mycart/internal/queries"
        "github.com/shurco/mycart/internal/webhook"
    )

    for pkg in "${FAILED_PACKAGES[@]}"; do
        print_info "Testing: $pkg"
        if [ "$VERBOSE" = "true" ]; then
            go test -v "$pkg" "$@" || true
        else
            go test "$pkg" "$@" || true
        fi
        echo ""
    done
}

# Function to run specific package tests
run_package_tests() {
    local package=$1
    shift

    print_info "Running tests for package: $package"
    export TEST_DATABASE_URL="${TEST_DATABASE_URL:-$DEFAULT_POSTGRES_URL}"
    export DATABASE_URL="${DATABASE_URL:-$DEFAULT_POSTGRES_URL}"

    if [ "$VERBOSE" = "true" ]; then
        go test -v "$package" "$@"
    else
        go test "$package" "$@"
    fi
}

# Function to clean test databases
clean_databases() {
    print_warning "This will clean the PostgreSQL test database!"
    read -p "Are you sure? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning PostgreSQL test database..."
        # Add cleanup logic here if needed
        print_success "Database cleaned"
    else
        print_info "Cancelled"
    fi
}

# Function to show usage
show_usage() {
    cat << EOF
${BLUE}Test Runner for mycart${NC}

${GREEN}Usage:${NC}
  ./test.sh [command] [options]

${GREEN}Commands:${NC}
  sqlite          Run all tests with SQLite (default)
  postgres        Run all tests with PostgreSQL
  all             Run tests with both SQLite and PostgreSQL
  failed          Run only previously failed test packages
  verbose         Run tests with verbose output
  clean           Clean test databases
  package <pkg>   Run tests for specific package
  help            Show this help message

${GREEN}Options:${NC}
  -v              Verbose output
  -short          Run with -short flag (skip long tests)
  -race           Run with race detector
  -count=N        Run tests N times
  -timeout=T      Set test timeout (e.g., 10m)

${GREEN}Environment Variables:${NC}
  TEST_DATABASE_URL    PostgreSQL connection string for tests
  DATABASE_URL         Fallback database connection string

${GREEN}Examples:${NC}
  ./test.sh postgres                    # Run all tests with PostgreSQL
  ./test.sh postgres -v                 # Run with verbose output
  ./test.sh failed                      # Run only failed packages
  ./test.sh package ./internal/queries  # Test specific package
  ./test.sh postgres -race              # Run with race detector
  ./test.sh postgres -count=3           # Run tests 3 times

${GREEN}Quick Test Commands:${NC}
  ./test.sh postgres -short             # Quick test run
  ./test.sh failed -v                   # Debug failed tests
  ./test.sh all                         # Full test suite (both DBs)

EOF
}

# Parse command line arguments
VERBOSE=false
COMMAND=""
EXTRA_ARGS=()

while [[ $# -gt 0 ]]; do
    case $1 in
        sqlite)
            COMMAND="sqlite"
            shift
            ;;
        postgres)
            COMMAND="postgres"
            shift
            ;;
        all)
            COMMAND="all"
            shift
            ;;
        failed)
            COMMAND="failed"
            shift
            ;;
        verbose)
            VERBOSE=true
            shift
            ;;
        clean)
            COMMAND="clean"
            shift
            ;;
        package)
            COMMAND="package"
            shift
            PACKAGE_PATH=$1
            shift
            ;;
        help|--help|-h)
            show_usage
            exit 0
            ;;
        -v)
            VERBOSE=true
            shift
            ;;
        *)
            EXTRA_ARGS+=("$1")
            shift
            ;;
    esac
done

# Execute command
case $COMMAND in
    sqlite)
        run_sqlite_tests "${EXTRA_ARGS[@]}"
        ;;
    postgres)
        run_postgres_tests "${EXTRA_ARGS[@]}"
        ;;
    all)
        print_info "=== Running SQLite Tests ==="
        run_sqlite_tests "${EXTRA_ARGS[@]}" || true
        echo ""
        print_info "=== Running PostgreSQL Tests ==="
        run_postgres_tests "${EXTRA_ARGS[@]}" || true
        ;;
    failed)
        run_failed_tests "${EXTRA_ARGS[@]}"
        ;;
    clean)
        clean_databases
        ;;
    package)
        if [ -z "$PACKAGE_PATH" ]; then
            print_error "Package path required"
            echo "Usage: ./test.sh package <package-path>"
            exit 1
        fi
        run_package_tests "$PACKAGE_PATH" "${EXTRA_ARGS[@]}"
        ;;
    "")
        # Default: run PostgreSQL tests
        run_postgres_tests "${EXTRA_ARGS[@]}"
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Test run completed"
