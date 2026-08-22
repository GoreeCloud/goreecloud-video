# GoreeCloud Video Architecture

## Current state

The current server source is based on Jellyfin 10.11.11. This provides a mature reference implementation for media probing, library scanning, streaming, transcoding, subtitles, sessions, metadata, and device interoperability.

The current state is a **GoreeCloud-maintained fork**, not a native GoreeCloud successor.

## Architectural direction

GoreeCloud will progressively place inherited application-specific behavior behind GoreeCloud-owned interfaces. The goal is controlled replacement rather than a destructive full rewrite.

Priority ownership boundaries are:

1. product configuration and capability flags
2. user accounts, profiles, authorization, and policy
3. library definitions and media classification
4. metadata provider abstraction
5. playback policy and device capability decisions
6. transcoding-worker orchestration
7. sessions and synchronized viewing
8. recommendations and discovery
9. client-facing APIs
10. first-party Glaze UI clients
11. GoreeCloud platform integrations

## Retained foundations

Mature foundational technology may remain when replacement would add risk without meaningful independence. Examples may include:

- FFmpeg-compatible media processing
- .NET runtime and libraries
- standards-based codecs and containers
- operating-system media acceleration interfaces
- mature database or serialization libraries

Retention of a dependency does not make the product non-native if GoreeCloud independently controls the meaningful product architecture around it.

## Deployment target

The planned private deployment path is:

```text
TrueNAS media datasets
  → least-privilege mounts
  → GoreeCloud Family Services VM
  → GoreeCloud Video server / media workers
  → Caddy HTTPS
  → NetBird private access
  → approved GoreeCloud Video clients
```

Production deployment is not part of the current repository-foundation milestone.

## Data separation

Keep these classes separate:

- authoritative media
- application database and user state
- metadata indexes and artwork cache
- temporary transcode segments and scratch data
- configuration and safe deployment definitions
- secrets and credentials

Disposable caches and transcodes must not be treated as irreplaceable data. Secrets must not be committed to Git.

## Native-transition rule

A component is not considered native merely because it is renamed, rebranded, visually restyled, or heavily modified. Native status requires substantive architectural and developmental independence.
