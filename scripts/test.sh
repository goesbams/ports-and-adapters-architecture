#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests...${NC}"

# Run tests with coverage
go test -v -cover ./... -coverprofile=coverage.out

# Check if tests passed
if [ $? -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    
    # Generate coverage report
    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}Coverage report generated: coverage.html${NC}"
else
    echo -e "${RED}Tests failed!${NC}"
    exit 1
fi