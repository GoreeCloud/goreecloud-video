# GoreeCloud Video

GoreeCloud Video is a privacy-first, self-hosted video streaming application for movies, television, home videos, family media, and other approved video content.

> **Development status:** Active native development. GoreeCloud-owned code under `native/` is the active product-development authority. The inherited Jellyfin 10.11.11 tree remains in this repository for provenance, compatibility research, migration, recovery, and rollback; it is not the approved long-term GoreeCloud application architecture and this repository is not yet production-ready.

## Product scope

GoreeCloud Video is intentionally video-first. First-class library categories are:

- Movies
- Television
- Home Videos
- Family Media
- Other Videos

General music streaming, audiobook libraries, ebooks, podcasts, and photo-backup workflows are outside the GoreeCloud Video product boundary and belong to other GoreeCloud applications.

## Current native foundation

The first GoreeCloud-owned native server foundation is implemented under `native/server/` and currently includes:

- first-party video-library item validation with video MIME enforcement;
- deterministic video-only scan-candidate policy for common video containers while rejecting audio, images, and extensionless candidates;
- a first-party playback decision model for Direct Play, remux, transcode, and fail-closed denial based on client container, codec, resolution, bitrate, and transcoding capabilities;
- focused Go tests and a dedicated Native Server GitHub Actions validation workflow.

The playback decision model is a policy foundation only. FFmpeg worker execution, session orchestration, subtitles, alternate-audio behavior, HDR/tone-mapping policy, adaptive delivery, production identity, clients, and production deployment remain under development.

## Development direction

GoreeCloud Video is developed natively from the ground up under GoreeCloud-controlled application boundaries. Critical mature media-engine knowledge and narrowly bounded dependencies may be used where justified, but inherited product architecture must not become the permanent application authority.

The historical Jellyfin tree is retained so GoreeCloud can compare behavior, preserve license/provenance obligations, support compatibility and migration work, and maintain a recovery path while native capabilities replace inherited application behavior.

The primary long-term user experience is a first-party Glaze UI client integrated with GoreeCloud platform security, privacy, and resilience requirements.

## Upstream provenance

The repository preserves its original Jellyfin Server import history:

- Upstream project: Jellyfin Server
- Upstream repository: `jellyfin/jellyfin`
- Initial baseline tag: `v10.11.11`
- Initial baseline commit: `1fbd8739292cce610231be93daf43368733edf63`

See [`docs/goreecloud/UPSTREAM.md`](docs/goreecloud/UPSTREAM.md) for provenance and synchronization rules.

## Architecture and governance

See:

- [`docs/goreecloud/ARCHITECTURE.md`](docs/goreecloud/ARCHITECTURE.md)
- [`docs/goreecloud/PRODUCT_BOUNDARY.md`](docs/goreecloud/PRODUCT_BOUNDARY.md)
- [`docs/goreecloud/DEVELOPMENT.md`](docs/goreecloud/DEVELOPMENT.md)
- [`docs/goreecloud/CI.md`](docs/goreecloud/CI.md)

GoreeCloud-specific changes should be developed on topic branches, validated on the exact candidate revision, and reviewed before integration into `main`.

## Security and privacy

Do not commit credentials, private keys, production `.env` files, private viewing activity, media-library contents, private storage paths, or other sensitive information. GoreeCloud Video is intended to follow least-privilege, local-first, Wardveil Security, Privacy Shield, and Everkeep requirements as those integrations are implemented and validated.

See [`SECURITY.md`](SECURITY.md).

## Status

GoreeCloud Video remains active development software. Native scanning and playback-policy foundations do not constitute production playback acceptance, and no production deployment is approved by this README.

## Upstream attribution and license

This repository preserves Jellyfin-derived source and Git history. Jellyfin is licensed under the GNU General Public License version 2, and inherited source plus derivative modifications remain subject to the applicable GPL-2.0 obligations.

See [`LICENSE`](LICENSE) and [`docs/goreecloud/UPSTREAM.md`](docs/goreecloud/UPSTREAM.md).
