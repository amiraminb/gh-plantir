# Plantir

A CLI tool to manage GitHub pull requests where you're requested as a reviewer.

> Named after the Palantír - the seeing stones from Lord of the Rings.

## Requirements

- Go 1.21+
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- AWS CLI with SSO configured (for AI review feature)

### Setting up GitHub CLI

```bash
# Install (macOS)
brew install gh

# Authenticate
gh auth login
```

Follow the prompts to authenticate via browser or token.

## Installation

```bash
go install github.com/amiraminb/plantir@latest
```

Or build from source:

```bash
git clone https://github.com/amiraminb/plantir.git
cd plantir
go build -o plantir .
```

## Usage
Few usecases. For more look at `plantir -h`

```bash
# List PRs awaiting your review (max 20, newest first)
plantir list

# Show only PRs waiting for your response (pending review)
plantir list -p

# Filter by repository
plantir list --repo=auth

# Filter by type (feature, dependabot, bot)
plantir list --type=feature
plantir list --type=dependabot

# Show all PRs for a team (pending + reviewed)
plantir list --team=org/team-name

# Show only PRs waiting for a team's review
plantir list --team=org/team-name -p

# Show only stale PRs (older than N days)
plantir list --stale=7

# Show more results
plantir list --limit=50
plantir list --limit=0  # unlimited

# JSON output (for scripting)
plantir list --json

# Combine filters
plantir list --repo=auth --type=feature --stale=3

# Open a PR in your browser
plantir open 1234

# Show summary statistics
plantir stats

# AI-powered PR review (requires AWS SSO login)
plantir review 1234

# Review PR from a specific repo
plantir review 1234 --repo=owner/repo

# Use a different AWS profile
plantir review 1234 --profile=my-profile
```

### AI Review Setup

The `review` command uses Claude via AWS Bedrock. You'll need:
- AWS CLI configured with a profile that has Bedrock access
- The profile should have access to Claude models

SSO login is triggered automatically if your session is expired.
