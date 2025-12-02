# Branch Protection Rules for Arbitrum-sequencer-decoder

## Protected Branches

### Main Branch (`main`)
The `main` branch is the production branch and requires the highest level of protection:

- **Required Status Checks:**
  - All automated tests must pass
  - Code coverage must not decrease below 80%
  - Linting checks must pass
  - Security scanning must pass

- **Required Reviews:**
  - Minimum 2 approvals required for pull requests
  - Approved reviews must be from senior developers or team leads
  - Reviewers cannot approve their own pull requests

- **Other Protections:**
  - Cannot be force pushed
  - Cannot be deleted
  - Push restrictions (only authorized users can push)
  - Require linear history (no merge commits, only rebase and merge or squash and merge)

### Develop Branch (`develop`)
The `develop` branch is the integration branch for features and requires moderate protection:

- **Required Status Checks:**
  - All automated tests must pass
  - Linting checks must pass
  - Optional code coverage threshold (70%)

- **Required Reviews:**
  - Minimum 1 approval required for pull requests
  - Reviewers should be team members familiar with the code area

- **Other Protections:**
  - Cannot be force pushed
  - Cannot be deleted
  - Push restrictions (only team members can push)

## Branch Management Policies

### Who Can Push to Protected Branches
- Team leads and senior developers can directly push to `develop` for urgent fixes
- No one can directly push to `main` - all changes must come through pull requests from release branches
- Emergency hotfixes to `main` require approval from project owner

### Pull Request Requirements
- PRs to `main` must come from a release branch (`release/vX.Y.Z`)
- PRs to `develop` can come from feature or fix branches
- PRs must have clear, descriptive titles and descriptions
- PRs must reference related issues
- PRs must not have any failing checks

### Branch Cleanup
- Feature branches should be deleted after successful merge
- Release branches should be deleted after merging to both `main` and `develop`
- Stale branches (not updated in >30 days) will be cleaned up automatically

## Enforcement Mechanisms

### GitHub Settings
The following settings should be configured in the repository:

1. **Main branch:**
   - Require pull request reviews before merging
   - Dismiss stale pull request approvals when new commits are pushed
   - Require status checks to pass before merging
   - Require branches to be up to date before merging
   - Restrict who can push to matching branches
   - Prevent force pushing to matching branches
   - Prevent deletion of matching branches

2. **Develop branch:**
   - Require pull request reviews before merging
   - Require status checks to pass before merging
   - Restrict who can push to matching branches