# Security Policy

## Project status

GoreeCloud Video is in active early development and is not yet classified as a production-ready GoreeCloud release. Security-sensitive behavior inherited from Jellyfin remains under review as the project moves toward GoreeCloud-controlled architecture.

## Reporting a vulnerability

Do not publish exploit details, credentials, private media information, access tokens, or other sensitive information in a public issue.

For a suspected vulnerability in GoreeCloud-specific code, use GitHub's private security-reporting facilities for this repository when available. If private reporting is unavailable, contact the repository owner through an established private GoreeCloud administrative channel before disclosing technical details publicly.

For a vulnerability that clearly exists unchanged in upstream Jellyfin code, also review the upstream Jellyfin security process so the upstream project can receive an appropriate report.

## Security boundaries

Changes affecting any of the following require explicit security review:

- authentication and authorization
- user/profile isolation
- session and token handling
- filesystem and media-library access
- upload, download, subtitle, and attachment handling
- transcoding worker privileges
- network listeners and reverse-proxy assumptions
- API input validation
- database migrations and persistent user data
- secrets and configuration loading
- container privileges and device access
- remote playback and device registration

## Sensitive information

Never commit:

- passwords or API tokens
- private SSH or TLS keys
- OAuth client secrets
- production `.env` files
- database credentials
- private media-library contents or paths that expose sensitive information
- household viewing history or private user data
- production backups or database exports

Use safe templates and documented placeholders instead.

## Dependency and upstream security

The maintained-fork phase retains substantial upstream code. GoreeCloud should continue monitoring relevant Jellyfin security fixes while independently reviewing GoreeCloud-specific changes. Upstream updates are not merged automatically; each update must be reviewed for compatibility, security impact, and product-boundary effects.
