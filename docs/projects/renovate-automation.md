# Renovate Bot Dependency Automation

## Overview

Replace manual Dependabot alert handling with Renovate Bot for automated dependency management. Renovate will auto-open and auto-merge patch/minor dependency PRs after CI passes, eliminating the need to manually respond to Dependabot security alerts. Dependabot security updates will be disabled to avoid the Go-specific noise problem where Dependabot raises false alerts for unimported vulnerable subpackages.

## Relevant System Architecture Before Project Started

- No `.github/` directory exists — no CI, no Dependabot config, no workflows
- Dependencies managed via `go.mod` / `go.sum` (Go modules)
- Direct dependencies: `go-git/go-git`, `google/uuid`, `spf13/cobra`, `stretchr/testify`
- Security alerts handled manually via GitHub Dependabot alerts UI
- Recent history shows manual version bumps in response to Dependabot/security alerts (e.g. `golang.org/x/crypto`, `go-git`)

## Target Architecture

```text
GitHub Repo
├── Branch protection (main)  ← Requires CI status check to pass before merge
└── .github/
    ├── workflows/
    │   └── ci.yml           ← Runs `go test` on push and PRs
    └── renovate.json        ← Renovate config: automerge patch/minor after CI, group Go deps
Renovate Bot (GitHub App)
    ├── Opens PRs for outdated/vulnerable deps
    ├── Opens grouped PRs → GitHub notifies → you merge manually after CI passes
    └── Holds major version bumps for manual review
```

## Issues

### [ ] I1: Add GitHub Actions CI workflow

Add a `.github/workflows/ci.yml` that runs `go test` on every push and pull request. This is a prerequisite for Renovate's automerge — without it, Renovate merges immediately without validating that tests pass.

#### Details About the System Prior to the Start of this Issue

No `.github/` directory exists. Tests live in `main_test.go` (12 test functions covering sync scenarios, fast-forward behavior, merge detection, and go-git behavior). The module is `github.com/gotascii/gitsync`, targeting Go 1.24.

#### Steps

- [x] Create `.github/workflows/ci.yml`:

  ```yaml
  name: CI
  on:
    push:
      branches: [main]
    pull_request:
  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with:
            go-version-file: go.mod
        - run: go test ./...
  ```

#### Tests

- [ ] Push the workflow file — Actions tab shows a `CI` workflow run on `main`
- [ ] `go test` step passes (green)
- [ ] Open a test PR — CI runs and reports status on the PR

### [ ] I2: Configure branch protection on main

Require the CI status check to pass before any PR can be merged into `main`. This ensures Renovate's `platformAutomerge` waits for `go test` to go green rather than merging immediately.

#### Details About the System Prior to the Start of this Issue

No branch protection rules exist on `main`. Without them, `platformAutomerge` bypasses CI entirely.

#### Steps

- [ ] Go to repo Settings → Branches → Add branch protection rule for `main`
- [ ] Enable "Require status checks to pass before merging"
- [ ] Search for and add the `test` status check (from the CI workflow job name)
- [ ] Enable "Require branches to be up to date before merging"
- [ ] Enable "Do not allow bypassing the above settings"

#### Tests

- [ ] Open a test PR with a failing `go test` — merge button is blocked
- [ ] Fix the test — merge button becomes available after CI passes

### [ ] I3: Install and configure Renovate Bot

Enable the Renovate GitHub App on this repository and add a `renovate.json` config that automatically opens grouped PRs for Go dependency updates. PRs require manual merge — GitHub notifies on PR open via the usual notification settings.

#### Details About the System Prior to the Start of this Issue

No Renovate config exists. CI will be in place (from I1) but without branch protection rules requiring it, `platformAutomerge` merges immediately without waiting for CI. Branch protection (I2) must be configured first to enforce the CI gate.

#### Steps

- [ ] Install Renovate GitHub App on the repo via <https://github.com/apps/renovate> — enable for `gotascii/gitsync` only
- [ ] Create `.github/renovate.json`:

  ```json
  {
    "$schema": "https://docs.renovatebot.com/renovate-schema.json",
    "extends": ["config:recommended"],
    "packageRules": [
      {
        "matchManagers": ["gomod"],
        "groupName": "Go dependencies",
        "matchUpdateTypes": ["patch", "minor", "digest"]
      }
    ]
  }
  ```

- [ ] Disable Dependabot security updates in repo Settings → Code security → Dependabot security updates → Disable (to avoid Go noise-machine false positives)

#### Tests

- [ ] Renovate skips onboarding (config already exists) and goes directly to scanning dependencies
- [ ] Renovate opens a grouped "Go dependencies" PR (or confirms all deps are up to date)
- [ ] Renovate opens grouped PR and GitHub sends a notification — merge manually after CI passes
- [ ] No Dependabot security update PRs appear after disabling
