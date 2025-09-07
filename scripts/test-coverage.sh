#!/bin/bash

# test-coverage.sh - Generate test coverage report for Obtura

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_DIR="coverage"
COVERAGE_FILE="${COVERAGE_DIR}/coverage.out"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"
COVERAGE_PROFILE="${COVERAGE_DIR}/profile.out"

echo -e "${GREEN}🧪 Running tests with coverage...${NC}"

# Create coverage directory
mkdir -p ${COVERAGE_DIR}

# Run tests with coverage for all packages
echo -e "${YELLOW}Running unit tests...${NC}"
go test -v -race -coverprofile=${COVERAGE_PROFILE} -covermode=atomic ./...

# Merge coverage profiles if needed (for multiple packages)
echo -e "${YELLOW}Processing coverage data...${NC}"
go tool cover -func=${COVERAGE_PROFILE} -o=${COVERAGE_FILE}

# Generate HTML coverage report
echo -e "${YELLOW}Generating HTML report...${NC}"
go tool cover -html=${COVERAGE_PROFILE} -o=${COVERAGE_HTML}

# Display coverage summary
echo -e "\n${GREEN}📊 Coverage Summary:${NC}"
go tool cover -func=${COVERAGE_PROFILE} | grep -E '^total:|^github.com/btassone/obtura' | tail -20

# Calculate total coverage percentage
TOTAL_COVERAGE=$(go tool cover -func=${COVERAGE_PROFILE} | grep '^total:' | awk '{print $3}' | sed 's/%//')

# Display coverage badge
echo -e "\n${GREEN}📈 Total Coverage: ${TOTAL_COVERAGE}%${NC}"

# Check coverage threshold (optional)
THRESHOLD=70.0
if (( $(echo "$TOTAL_COVERAGE < $THRESHOLD" | bc -l) )); then
    echo -e "${RED}❌ Coverage is below ${THRESHOLD}%${NC}"
    echo -e "${YELLOW}Consider adding more tests to improve coverage.${NC}"
else
    echo -e "${GREEN}✅ Coverage meets threshold of ${THRESHOLD}%${NC}"
fi

echo -e "\n${GREEN}📄 Reports generated:${NC}"
echo -e "  - Text report: ${COVERAGE_FILE}"
echo -e "  - HTML report: ${COVERAGE_HTML}"
echo -e "\n${YELLOW}Open ${COVERAGE_HTML} in your browser to view detailed coverage.${NC}"

# Optional: Open coverage report in browser
if command -v open &> /dev/null; then
    read -p "Open coverage report in browser? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open ${COVERAGE_HTML}
    fi
fi