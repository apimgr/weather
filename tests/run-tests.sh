#!/usr/bin/env bash
# Run all tests for the weather service

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧪 Weather Service Test Suite${NC}"
echo -e "${BLUE}==============================${NC}"
echo ""

# Get to project root
cd "$(dirname "$0")/.."

# Function to run tests
run_tests() {
    local name=$1
    local path=$2
    local flags=$3

    echo -e "${BLUE}Running ${name}...${NC}"

    if go test ${flags} ${path}; then
        echo -e "${GREEN}✅ ${name} passed${NC}"
        echo ""
        return 0
    else
        echo -e "${RED}❌ ${name} failed${NC}"
        echo ""
        return 1
    fi
}

# Parse arguments
COVERAGE=false
VERBOSE=false
BENCH=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -b|--bench)
            BENCH=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [-c|--coverage] [-v|--verbose] [-b|--bench]"
            exit 1
            ;;
    esac
done

# Build test flags
TEST_FLAGS=""
if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="${TEST_FLAGS} -v"
fi
if [ "$COVERAGE" = true ]; then
    TEST_FLAGS="${TEST_FLAGS} -cover -coverprofile=coverage.out"
fi

# Track failures
FAILED=0

# Run unit tests
echo -e "${YELLOW}📦 Unit Tests${NC}"
echo "-----------------------------------"

run_tests "Service Tests" "./tests/unit/services/..." "${TEST_FLAGS}" || FAILED=$((FAILED + 1))
run_tests "Handler Tests" "./tests/unit/handlers/..." "${TEST_FLAGS}" || FAILED=$((FAILED + 1))

# Run integration tests
echo -e "${YELLOW}🔗 Integration Tests${NC}"
echo "-----------------------------------"

run_tests "API Integration Tests" "./tests/integration/..." "${TEST_FLAGS}" || FAILED=$((FAILED + 1))

# Run e2e tests
echo -e "${YELLOW}🌐 End-to-End Tests${NC}"
echo "-----------------------------------"

run_tests "Setup Flow Tests" "./tests/e2e/..." "${TEST_FLAGS}" || FAILED=$((FAILED + 1))

# Run benchmarks if requested
if [ "$BENCH" = true ]; then
    echo -e "${YELLOW}⚡ Benchmarks${NC}"
    echo "-----------------------------------"

    go test -bench=. -benchmem ./tests/... || FAILED=$((FAILED + 1))
    echo ""
fi

# Generate coverage report if requested
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}📊 Coverage Report${NC}"
    echo "-----------------------------------"

    go tool cover -func=coverage.out
    echo ""

    # Generate HTML report
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}Coverage report saved to: coverage.html${NC}"
    echo ""
fi

# Summary
echo "==============================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ ${FAILED} test suite(s) failed${NC}"
    exit 1
fi
