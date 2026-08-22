# GoreeCloud Video

GoreeCloud Video is a privacy-first, self-hosted video streaming project for movies, TV shows, home videos, family media, and other approved video content. It is being developed as a GoreeCloud-maintained Jellyfin-derived foundation with a controlled fork-to-native transition and a first-party Glaze UI experience.

> **Development status:** Early foundation work. This repository currently remains substantially derived from Jellyfin 10.11.11. GoreeCloud-specific product boundaries, native interfaces, Glaze UI clients, and video-only behavior are being introduced incrementally. This repository should not yet be treated as a production-ready GoreeCloud Video release.

## Product scope

GoreeCloud Video is intentionally video-first. Planned first-class library types are:

- Movies
- TV Shows
- Home Videos
- Family Media
- Other Videos

General music streaming, audiobook libraries, ebooks, podcasts, and photo-backup workflows are outside the GoreeCloud Video product boundary. Those capabilities belong to other GoreeCloud applications.

## Development direction

The project begins from a mature Jellyfin server baseline so GoreeCloud can retain proven media scanning, codec handling, subtitle behavior, device negotiation, streaming, transcoding, and hardware-acceleration work while progressively replacing application-specific product behavior with GoreeCloud-controlled architecture.

The intended progression is:

`Jellyfin baseline → GoreeCloud-maintained fork → GoreeCloud product layer → architectural independence → native GoreeCloud-controlled capabilities`

The primary long-term user experience will be a first-party Glaze UI client rather than a permanent Jellyfin Web reskin.

## Initial upstream baseline

- Upstream project: Jellyfin Server
- Upstream repository: https://github.com/jellyfin/jellyfin
- Initial baseline tag: `v10.11.11`
- Initial baseline commit: `1fbd8739292cce610231be93daf43368733edf63`
- GoreeCloud default branch: `main`
- Recommended local upstream remote: `https://github.com/jellyfin/jellyfin.git`

See [`docs/goreecloud/UPSTREAM.md`](docs/goreecloud/UPSTREAM.md) for provenance and synchronization rules.

## Architecture

GoreeCloud-owned boundaries will progressively isolate user/profile management, libraries, metadata, playback policy, sessions, recommendations, transcoding orchestration, client APIs, and integrations from inherited upstream implementation details.

See:

- [`docs/goreecloud/ARCHITECTURE.md`](docs/goreecloud/ARCHITECTURE.md)
- [`docs/goreecloud/PRODUCT_BOUNDARY.md`](docs/goreecloud/PRODUCT_BOUNDARY.md)
- [`docs/goreecloud/DEVELOPMENT.md`](docs/goreecloud/DEVELOPMENT.md)

## Security and privacy

GoreeCloud Video is intended to follow GoreeCloud privacy-by-default, least-privilege, Wardveil Security, and sensitive-information-separation requirements. Do not commit credentials, private keys, production `.env` files, private viewing data, media-library contents, or other secrets to this repository.

See [`SECURITY.md`](SECURITY.md).

## Contributing

GoreeCloud-specific changes should be developed on topic branches and reviewed before merging into `main`. Upstream provenance and license obligations must remain explicit.

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## Upstream attribution and license

This repository is derived from the Jellyfin Server project and preserves its Git history. Jellyfin is licensed under the GNU General Public License version 2. The inherited source and modifications remain subject to applicable GPL-2.0 requirements.

The upstream Jellyfin project is available at https://github.com/jellyfin/jellyfin.

See [`LICENSE`](LICENSE) for the full license text.
