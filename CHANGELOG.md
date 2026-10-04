# Changelog

All notable changes to Tank Blaster will be documented in this file.

The format is inspired by Keep a Changelog, and the project uses the version in `VERSION` for release builds.

## [Unreleased]

### Notes

- Public issue and pull request workflow will start after the repository is published on GitHub.

## [1.0.0] - 2026-10-04

Initial public release candidate.

### Added

- Public repository metadata: GPL-3.0 license, contribution guide, security policy, and changelog.
- Intro scene with original intro artwork, text panels, MOD music playback, sound overlay, skip-on-click, and animated `remake` badge.
- Hall-of-Fame background music using bundled MOD playback.
- OS-specific user configuration storage in `tank.cfg`, including options, local rounds, online rounds, and online identity.
- Build metadata embedded into desktop, server, and Android builds.
- `make all` release packaging for Linux, Windows, Android, and server container tarball.
- Docker Compose examples for local port mapping and host-network server deployment.

### Changed

- Go module path renamed to `github.com/runzhammer/tankblaster`.
- Desktop window title is `Tank Blaster`.
- Release builds force `resources/config.yaml` to `debug.enabled: false` while building and restore the previous file afterwards.
- Android release builds are used for `make android` and `make all`; debug APKs require `make android-debug`.
- Multiplayer server hardened with HTTP timeouts, WebSocket read limits, connection/session/queue limits, stricter WebSocket origin handling, and escaped invite pages.

### Notes

- Original Tank Blaster graphics and sounds are included with permission from Axel Lauer.
- Bundled module music comes from public-domain library material from the 1990s/2000s.
