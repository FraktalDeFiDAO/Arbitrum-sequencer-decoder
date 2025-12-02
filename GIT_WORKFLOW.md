# Enterprise Git Workflow for Arbitrum-sequencer-decoder

## Branch Strategy: GitFlow-inspired approach

### Core Branches
- `main` - Production-ready code (protected)
- `develop` - Integration branch for features (protected)
- `feature/*` - Feature development branches
- `release/*` - Release preparation branches
- `hotfix/*` - Emergency fixes for production

### Branch Naming Conventions
- Feature: `feature/issue-number-description` (e.g., `feature/123-uniswap-v3-decoder`)
- Bug Fix: `fix/issue-number-description` (e.g., `fix/456-curve-calldata-parsing`)
- Hotfix: `hotfix/issue-number-description` (e.g., `hotfix/789-critical-security-fix`)
- Release: `release/vX.Y.Z` (e.g., `release/v1.2.0`)

## Commit Message Standards

### Format
```
<Type>: Short summary (<72 characters)

[Optional body explaining the change in detail]

[Optional footer with metadata like fixes #issue-number]
```

### Valid Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semi-colons, etc.)
- `refactor`: Code changes that neither fix a bug nor add a feature
- `perf`: Performance improvements
- `test`: Adding or modifying tests
- `chore`: Other changes that don't modify src or test files

### Examples
```
feat: Add Uniswap V3 decoder with price simulation

Added comprehensive Uniswap V3 decoding capabilities with
tick math implementation for accurate price estimation.

Fixes #123
```

```
fix: Handle zero amount edge case in Curve decoder

The Curve decoder was panicking when encountering transactions
with zero amounts. Now properly handles this edge case by
returning an appropriate error.

Fixes #456
```

## Pull Request Standards

### Requirements
- PRs must target the `develop` branch for features/fixes
- PRs to `main` must come from a `release` or `hotfix` branch
- Minimum 1 code review approval required
- All CI checks must pass
- PR title should follow commit message format
- Description must include:
  - Summary of changes
  - Testing performed
  - Breaking changes (if any)
  - Related issues

### PR Template
```markdown
## Summary
<!-- Brief description of the changes -->

## Type of Change
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
<!-- Describe how the changes were tested -->

## Breaking Changes
<!-- List any breaking changes -->

## Related Issues
Fixes # (issue numbers)
```

## Code Review Process

### Review Requirements
- All PRs require at least 1 approval from a team lead or senior developer
- Critical security fixes may require 2 approvals
- Automated checks must pass before review
- Reviewers should check:
  - Code quality and style compliance
  - Test coverage (minimum 80% for new code)
  - Performance implications
  - Security considerations
  - Documentation updates

### Review Turnaround Time
- Standard PRs: 24 hours
- Critical issues: 4 hours
- Weekends/holidays: 48 hours

## Release Process

### Versioning
- Follow Semantic Versioning (MAJOR.MINOR.PATCH)
- MAJOR: Breaking changes
- MINOR: New features (non-breaking)
- PATCH: Bug fixes

### Release Steps
1. Create release branch from develop: `release/vX.Y.Z`
2. Update version in relevant files
3. Update CHANGELOG.md
4. Run comprehensive tests
5. Get PR approval to merge to main
6. Tag the release on main branch
7. Create GitHub release with changelog
8. Merge main back to develop

## Git Hooks and Quality Checks

### Pre-commit Hook
Required checks before commits:
- Code formatting verification (gofmt)
- Linting (golangci-lint)
- Vulnerability scanning (gosec)
- Commit message format validation

### Pre-push Hook
Required checks before pushes:
- All tests pass
- Branch naming validation
- No secrets in code

## Security and Compliance

### Code Signing
- All commits on main and develop must be signed
- Use GPG keys for commit signing

### Sensitive Information
- No API keys, secrets, or credentials in code
- Use environment variables or secret management systems
- Regular scanning for exposed secrets

### Access Management
- Write access to main/develop requires team lead approval
- Temporary access granted for specific tasks only
- Regular access audits

## Workflows

### Feature Development
1. Create feature branch from develop
2. Implement feature with tests
3. Commit changes following standards
4. Push to remote
5. Create PR to develop with proper description
6. Address review comments
7. Merge after approval

### Hotfix Process
1. Create hotfix branch from main
2. Implement critical fix
3. Create PR to both main and develop
4. Expedited review process
5. Merge after approval
6. Create hotfix release

### Release Process
1. Create release branch from develop
2. Update versions and documentation
3. Comprehensive testing
4. Create PR to main
5. Merge after approval
6. Tag release and create GitHub release
7. Merge back to develop
</content>