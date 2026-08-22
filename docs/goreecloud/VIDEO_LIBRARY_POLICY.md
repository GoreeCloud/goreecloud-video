# Video Library Policy

## Purpose

GoreeCloud Video deliberately narrows the inherited Jellyfin library-management surface to video-oriented libraries while retaining the mature shared media engine needed for video playback.

This is a product-boundary policy. It is not permission to remove audio-track handling, subtitle processing, image handling, media probing, FFmpeg integration, transcoding, or other shared capabilities required by movies and television.

## Inherited collection-type mapping

The Jellyfin-compatible `CollectionTypeOptions` contract currently exposes these writable values:

- `movies`
- `tvshows`
- `music`
- `musicvideos`
- `homevideos`
- `boxsets`
- `books`
- `mixed`

GoreeCloud Video currently permits these inherited types for first-class library creation:

| GoreeCloud use | Inherited collection type | Status |
| --- | --- | --- |
| Movies | `movies` | Allowed |
| TV Shows | `tvshows` | Allowed |
| Home Videos | `homevideos` | Allowed |
| Family Media | `homevideos` initially | Allowed through named-library/product metadata rather than a new inherited enum value |
| Other Videos | `homevideos` or `mixed`, depending on organization | Allowed through compatibility mapping rather than a new inherited enum value |
| Mixed Movies and TV | `mixed` | Allowed compatibility type |

The following inherited first-class library types are blocked from new library creation:

- `music`
- `musicvideos`
- `books`
- `boxsets`
- an unspecified/null collection type

`boxsets` is blocked as a user-created first-class library type. Movie and television collection/curation features may still use inherited collection infrastructure internally where required.

## Enforcement boundary

The initial Milestone 1 enforcement point is the administrative virtual-library creation endpoint. `LibraryStructureController.AddVirtualFolder` validates the requested inherited collection type with `GoreeCloudVideoLibraryPolicy` before delegating to the inherited library manager.

Keeping the policy outside the inherited collection enum preserves protocol and upstream compatibility while giving GoreeCloud Video an explicit, independently testable product rule.

## First-party capability contract

First-party GoreeCloud clients must not infer supported library types from the complete inherited Jellyfin enum. The server exposes GoreeCloud Video's narrower contract at:

`GET /GoreeCloud/Video/Libraries/SupportedTypes`

The endpoint is protected by the same elevated/first-time-setup authorization policy used for library administration and returns the types represented by `GoreeCloudVideoLibraryPolicy.SupportedCollectionTypes`.

The initial response contract contains only:

- `movies`
- `tvshows`
- `homevideos`
- `mixed`

This gives the future first-party Glaze UI administration experience a GoreeCloud-owned capability source instead of coupling it to every collection type inherited from Jellyfin. New first-class library types must be added to the policy and its tests before a client exposes them.

## Known transition gaps

This policy is the first server-side boundary, not the completion of Milestone 1.

In particular, Jellyfin's inherited `homevideos` behavior historically covers both videos and photos. GoreeCloud Video must narrow scanning and presentation behavior so a Home Videos or Family Media library does not become a general photo-library surface. That work must be validated separately before Milestone 1 is considered complete.

Inherited music, book, photo, and related implementation code may remain in the repository while it is still shared, required for compatibility, or awaiting safe isolation. Presence in source does not make those surfaces supported GoreeCloud Video features.

## Validation requirements

Before this milestone can be marked complete, validation must cover at minimum:

- allowed and blocked library-creation types
- the first-party supported-library-types capability endpoint
- movie and TV library scanning
- Home Videos scanning without exposing a general photo-library product
- Direct Play
- remuxing
- transcoding
- subtitle selection and burn-in where required
- alternate audio tracks
- representative metadata and artwork flows

Playback-engine validation is intentionally separate from this policy because the policy must not accidentally remove capabilities needed by video playback.
