# Continuous Integration and Repository Automation

## Purpose

This record documents which automation inherited from the Jellyfin 10.11.11 baseline is retained, adapted, replaced, or removed for GoreeCloud Video.

Inherited automation is not trusted merely because it was functional upstream. Every workflow must fit GoreeCloud's branch names, permissions, secrets model, product boundary, and release process.

## Retained and adapted workflows

### Server Tests

`.github/workflows/ci-tests.yml` retains the inherited cross-platform .NET test suite because it directly validates the current server foundation. It now targets GoreeCloud's `main` branch and uses read-only repository permission.

### CodeQL

`.github/workflows/ci-codeql-analysis.yml` retains C# CodeQL analysis, changes branch targeting from upstream `master` to GoreeCloud `main`, and grants only the repository and security-event permissions required for analysis.

### ABI Compatibility

`.github/workflows/ci-compat.yml` retains API/assembly compatibility comparison while the server still exposes inherited Jellyfin-compatible assemblies. Jellyfin's `JF_BOT_TOKEN` dependency is removed. Pull-request comments use the repository-scoped GitHub token with explicit permissions.

This workflow may be reduced or retired later as GoreeCloud-owned interfaces replace inherited assembly boundaries.

### OpenAPI Compatibility

`.github/workflows/ci-openapi.yml` retains specification generation and pull-request comparison. Jellyfin-specific repository publication jobs and `REPO_HOST`, `REPO_USER`, and `REPO_KEY` secret dependencies are removed. The workflow targets `main` and uses the repository-scoped GitHub token for pull-request comments.

## Removed inherited workflows

The following Jellyfin community/release workflows are removed from the active GoreeCloud branch because they depend on Jellyfin-specific bots, projects, labels, triage repositories, release branches, or community processes:

- `commands.yml`
- `issue-stale.yml`
- `issue-template-check.yml`
- `project-automation.yml`
- `pull-request-conflict.yml`
- `pull-request-stale.yaml`
- `release-bump-version.yaml`

Removing these workflows does not erase their provenance: they remain available in the imported Jellyfin history and baseline tag.

## Issue templates

Inherited Jellyfin issue templates are replaced with GoreeCloud Video templates. The new templates explicitly preserve the video-only product boundary and warn reporters to remove secrets, private viewing activity, private media information, and other sensitive data before posting.

## Dependency automation

Renovate configuration no longer imports a Jellyfin organization preset. The local repository uses Renovate's recommended baseline so dependency policy can evolve under GoreeCloud ownership without silently inheriting future configuration from another organization.

## Branch protection

`main` should be protected after the retained CI workflows have produced stable check names. The intended minimum policy is:

- require pull requests before merge
- require the validated server-test and security checks selected for the current development stage
- block force pushes to `main`
- block branch deletion
- keep administrator bypass exceptional rather than routine

Branch protection is a repository-host setting and is tracked separately from these source-controlled workflow definitions.

## Security rule

No GoreeCloud workflow may depend on Jellyfin bot credentials or third-party production secrets simply because those references existed in the imported upstream repository. New secrets require a documented GoreeCloud operational purpose and least-privilege review.
