# Contributing to GoreeCloud Video

GoreeCloud Video is a maintained Jellyfin-derived project undergoing a controlled fork-to-native transition. Contributions must preserve truthful upstream provenance while moving the product toward GoreeCloud-controlled architecture.

## Branch workflow

1. Start from the current `main` branch.
2. Create a focused topic branch such as `agent/<purpose>`, `feature/<purpose>`, `fix/<purpose>`, or `docs/<purpose>`.
3. Keep commits focused and describe why a change is required.
4. Validate affected builds, tests, configuration, and documentation.
5. Open a pull request into `main`.
6. Do not merge a change that falsely claims production readiness or native independence.

## Development priorities

Prefer changes that establish or strengthen GoreeCloud-owned boundaries around:

- product configuration
- user/profile behavior
- libraries and media policy
- metadata providers
- playback and transcoding policy
- session orchestration
- recommendations and discovery
- client-facing APIs
- Glaze UI clients
- GoreeCloud Identity, Notify, Monitor, and Everkeep integrations

Do not rewrite mature foundational technology merely to increase the amount of GoreeCloud-owned code. Retain mature dependencies when replacement would add risk without meaningful product independence.

## Product boundary

The product is video-first. New general-purpose music, audiobook, ebook, podcast, or photo-backup product surfaces should not be introduced into GoreeCloud Video.

## Upstream changes

The initial upstream baseline is Jellyfin `v10.11.11` at commit `1fbd8739292cce610231be93daf43368733edf63`.

When reviewing later upstream changes:

- fetch from the Jellyfin upstream repository
- review the exact commit range
- preserve upstream authorship and history
- evaluate security and compatibility impact
- avoid blindly merging features that conflict with GoreeCloud Video's product boundary
- document significant accepted or rejected upstream changes

## Documentation

Architecture, migration, security, deployment, and non-obvious implementation decisions are part of the implementation. Update documentation in the same pull request when behavior or boundaries change.

## Licensing

Inherited Jellyfin source remains subject to GPL-2.0 and applicable copyright notices. Do not remove or obscure required license, copyright, attribution, modification, or source-availability obligations.
