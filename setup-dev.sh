#!/bin/bash

# Setup script for Arbitrum-sequencer-decoder development environment
# This script helps set up the local development environment following enterprise standards

echo "Setting up Arbitrum-sequencer-decoder development environment..."

# Check if we're in the project root
if [ ! -f "go.mod" ] && [ ! -f "README.md" ]; then
    echo "Error: This script should be run from the project root directory"
    exit 1
fi

# Configure git settings specific to this project
echo "Configuring project-specific git settings..."

git config core.hooksPath .git-hooks
echo "✓ Git hooks configured to use .git-hooks directory"

# Set up commit template (optional)
cat > .git-commit-template << 'EOF'
# <type>: Short summary (50 chars or less)
#
# Optional extended description. Wrap at 72 chars.
#
# Type can be:
#  feat     A new feature
#  fix      A bug fix
#  docs     Documentation only changes
#  style    Changes that do not affect the meaning of the code
#  refactor Code changes that neither fix a bug nor add a feature
#  perf     A code change that improves performance
#  test     Adding missing tests or correcting existing tests
#  chore    Other changes that don't modify src or test files
#
# Example: feat(decoder): add Uniswap V3 exactInput decoder
#
# Reference issues: #IssueNumber

EOF

git config commit.template .git-commit-template
echo "✓ Commit template configured"

# Check for required tools
echo "Checking for required tools..."

if command -v go >/dev/null 2>&1; then
    echo "✓ Go is installed: $(go version)"
    GO_VERSION=$(go version | cut -d " " -f 3)
    echo "  Go version: $GO_VERSION"
else
    echo "✗ Go is not installed or not in PATH"
    echo "  Please install Go 1.25+ from https://golang.org/dl/"
fi

if command -v golangci-lint >/dev/null 2>&1; then
    echo "✓ golangci-lint is installed: $(golangci-lint --version)"
else
    echo "⚠ golangci-lint is not installed"
    echo "  Consider installing for code quality checks:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

if command -v gosec >/dev/null 2>&1; then
    echo "✓ gosec is installed: $(gosec --version 2>&1 | head -n1)"
else
    echo "⚠ gosec is not installed"
    echo "  Consider installing for security scanning:"
    echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi

# Initialize the Go module if it doesn't exist
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init arbitrum-sequencer-decoder
    echo "✓ Go module initialized"
fi

echo
echo "Development environment setup complete!"
echo
echo "Next steps:"
echo "1. Review docs/general/GIT_WORKFLOW.md for branching and commit standards"
echo "2. Review docs/general/BRANCH_PROTECTION.md for protected branch rules"
echo "3. Review agent_docs/ for technical implementation details"
echo "4. Create your first feature branch: git checkout -b feature/description"
echo
echo "Remember: All commits should follow the conventional commit format"
echo "Example: feat(decoder): implement Uniswap V3 calldata parser"