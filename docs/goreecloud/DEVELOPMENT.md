# Development Plan

## Repository foundation

The GoreeCloud repository preserves the complete imported Jellyfin history required to maintain derived changes. `main` began at Jellyfin `v10.11.11` / `1fbd8739292cce610231be93daf43368733edf63`.

Meaningful GoreeCloud work should be developed through reviewable branches and pull requests.

## Phase 0 — Governance and provenance

Establish:

- truthful upstream provenance
- GoreeCloud repository identity
- ownership and contribution guidance
- security reporting
- architecture and product-boundary documentation
- CI review and cleanup
- branch-protection expectations

## Phase 1 — Video-only server baseline

Identify, disable, isolate, or remove inherited product surfaces that are outside GoreeCloud Video's scope while preserving required media-server behavior for movies, television, home videos, family media, and other videos.

A feature should not be removed merely because its name suggests another media type if the underlying implementation is also required for video playback. Changes must be evidence-based and tested.

## Phase 2 — GoreeCloud-owned boundaries

Introduce interfaces and service boundaries around user/profile behavior, libraries, metadata, playback policy, sessions, recommendations, transcoding, and client-facing APIs. Prefer incremental seams that can coexist with inherited code.

## Phase 3 — First-party Glaze UI client

Build the GoreeCloud-controlled client experience without making Jellyfin Web a permanent product dependency.

## Phase 4 — Production playback validation

Validate Direct Play, remuxing, software transcoding, hardware acceleration, HDR behavior, subtitles, audio-track selection, concurrent streams, resource limits, and failure recovery on representative target hardware.

## Phase 5 — Living-room and offline clients

Prioritize Android mobile, Android TV / Google TV, and Linux packaging before expanding to additional platforms.

## Phase 6 — Discovery and shared viewing

Add privacy-preserving recommendations, configurable discovery rows, intro/recap/credits markers, and first-party synchronized Watch Together behavior.

## Validation rule

A web page loading or a single media file playing is not production acceptance. Production readiness requires multi-user isolation, representative library testing, supported client validation, playback/transcoding coverage, security controls, storage permissions, backup/restore, monitoring, rollback, and documented recovery procedures.
