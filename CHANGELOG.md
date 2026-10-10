# Changelog

All notable changes to Tank Blaster will be documented in this file.

The format is inspired by Keep a Changelog, and the project uses the version in `VERSION` for release builds.

## [Unreleased]

### Changed

- `make all` now keeps the Android release as `dist/tankblaster-android.apk` instead of creating a versioned Android ZIP.

### Notes

- Public issue and pull request workflow will start after the repository is published on GitHub.

## [1.0.7] - 2026-10-09

### Added

- Update check in the player selection screen that compares the running version with the semantic version in the GitHub `main` branch and links to `https://www.tankblaster.de`.
- Online play now supports computer players and multiple human players in the same session.
- Online synchronization snapshots for economy, inventories, shop order, palm state, and cloud aggression.
- Regression coverage for online shop synchronization, deterministic projectile effects, airstrike behavior, and update checks.

### Changed

- Airstrike behavior now follows the original Axel Lauer implementation more closely, with fixed bomb timing, original-style offsets, wind compensation, and synchronized online impacts.
- Online projectile and death effects now use deterministic seeds for zero-power explosions, color blops, splitter bombs, MFS triple shots, and airstrikes.
- Online fire commands now reject duplicate command sequences to prevent repeated shots under reconnect or packet replay conditions.
- Release configuration now starts with debug mode disabled.

### Fixed

- Prevented double purchases in the online shop when local slot-player commands echo back from the server.
- Kept online shop money, scores, weapon stock, palm aggression, and cloud aggression synchronized between host and joined clients.
- Prevented divergent terrain after tank explosions, color blops, splitter bombs, MFS triple shots, and airstrike impacts.

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
