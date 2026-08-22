# Upstream Provenance

## Initial source foundation

GoreeCloud Video was initialized by importing the Jellyfin Server repository history and creating GoreeCloud's `main` branch from the Jellyfin `v10.11.11` tag.

- Upstream project: Jellyfin Server
- Upstream repository: https://github.com/jellyfin/jellyfin
- Upstream license: GNU GPL version 2
- Initial baseline tag: `v10.11.11`
- Initial baseline commit: `1fbd8739292cce610231be93daf43368733edf63`
- GoreeCloud repository: `GoreeCloud/goreecloud-video`

The imported Git history is intentionally preserved. GoreeCloud must not rewrite history to make this derived project appear original.

## Recommended remotes

A development checkout should use:

```text
origin   git@github.com:GoreeCloud/goreecloud-video.git
upstream https://github.com/jellyfin/jellyfin.git
```

`origin` is the canonical GoreeCloud repository. `upstream` is used for reviewing later Jellyfin changes.

## Synchronization policy

Upstream updates are reviewed, not automatically absorbed. Before accepting a later upstream baseline or patch set:

1. fetch the upstream refs
2. identify the exact commit range
3. review security fixes separately from product changes
4. identify conflicts with the GoreeCloud Video product boundary
5. preserve upstream authorship and commit history where changes are imported
6. run affected tests and build validation
7. document the accepted baseline or selected commits

Large upstream merges should not silently reintroduce product surfaces GoreeCloud has intentionally removed or replaced.

## Provenance categories

When practical, significant new work should be understood as one of:

- unchanged inherited upstream code
- modified upstream code
- independently implemented GoreeCloud code
- third-party library or component
- generated code or external asset

This distinction supports maintenance, security review, and licensing clarity throughout the fork-to-native transition.
