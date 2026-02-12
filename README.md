# Plantir

A `gh` CLI extension to manage GitHub pull requests where you're requested as a reviewer, and summarize merged PRs for your team using AI.

> Named after the Palantír - the seeing stones from Lord of the Rings.

## Requirements

- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- AWS credentials with Bedrock access (only needed for the `summary` command)

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
gh extension install amiraminb/gh-plantir
```

Or build from source:

```bash
git clone https://github.com/amiraminb/gh-plantir.git
cd gh-plantir
make install
```

## Upgrade

```bash
gh extension upgrade plantir
```

## Usage

Few usecases. For more look at `gh plantir -h`

```bash
# List PRs awaiting your review (max 20, newest first)
gh plantir list

# Show only PRs waiting for your response (pending review)
gh plantir list -p

# Filter by repository
gh plantir list --repo=auth

# Show all PRs for a team (pending + reviewed)
gh plantir list --team=org/team-name

# Show only PRs waiting for a team's review
gh plantir list --team=org/team-name -p

# Show more results
gh plantir list --limit=50
gh plantir list --limit=0  # unlimited

# JSON output (for scripting)
gh plantir list --json

# Open a PR in your browser
gh plantir open 1234
```

### Summary

Fetch merged PRs reviewed by a team and generate an AI-powered summary using AWS Bedrock (Claude). Dependabot PRs are excluded.

```bash
# Configure AWS settings (required before first use)
gh plantir config

# Summarize merged PRs for a team on a specific date
gh plantir summary -t org/team-name -d 2026-02-11

# Summarize a date range
gh plantir summary -t org/team-name -d 2026-02-01..2026-02-11

# List fetched PRs without generating a summary
gh plantir summary -t org/team-name -d 2026-02-11 -l
```
