# Tank Blaster

![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)
![AI-assisted development: about 90%](https://img.shields.io/badge/AI--assisted%20development-about%2090%25-purple)

Tank Blaster is a small artillery game with an unreasonable amount of confidence: tanks, wind, awkward trajectories, questionable shopping decisions, and just enough chaos to make a perfect shot feel like science.

In short: players pick tanks, buy weapons, aim carefully, and then blame physics.

This repository contains the Go/Ebiten remake of Tank Blaster with desktop, Android, and online multiplayer builds.

> **AI-assisted development**
>
> About 90% of this remake was developed with AI assistance. Architecture, decisions, testing, review, and publication remain human-owned and human-accountable.

## Repository

- License: [GPL-3.0](LICENSE)
- Changelog: [CHANGELOG.md](CHANGELOG.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)

## Assets

The original Tank Blaster graphics and sounds are included with permission from the original author, Axel Lauer. The bundled music tracks come from public-domain library material from the 1990s/2000s.

Please do not add new third-party graphics, sounds, fonts, or music unless their origin and license are clear and compatible with this repository.

## Development

Requirement: the Go version declared in [go.mod](go.mod). Windows builds use `rsrc` automatically through `go run`. For Android setup, see [docs/android.md](docs/android.md).

Run the desktop game locally:

```sh
make run
```

Useful development commands:

```sh
make test
make live-test
make live-test-auto
make live-test-cp
```

The live-test targets start a local multiplayer server and two clients. They use 5 rounds by default:

```sh
make live-test-auto LIVE_TEST_ROUNDS=3
```

Useful debug overrides:

```sh
TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=intro make run
TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online make run
```

The normal game flow starts with the intro scene. Its assets live in `resources/intro/`.

## Build

The release version is stored in [VERSION](VERSION). Desktop binaries, server binaries, and the Android app embed that version together with the Git commit and build time.

For Android releases, `tankblasterVersionCode` in [android/gradle.properties](android/gradle.properties) must be increased monotonically for every published APK.

Release builds make sure that `resources/config.yaml` is embedded with `debug.enabled: false`. If Make has to change the file temporarily, it prints a notice and restores the previous contents after the build.

Build everything for a release:

```sh
make all
```

`make all` creates:

- `dist/tankblaster-VERSION-linux.zip`
- `dist/tankblaster-VERSION-windows.zip`
- `dist/tankblaster-VERSION-android.zip`
- `dist/tankblaster-server-VERSION-linux.tar`

Android is built as a release APK for `make all` and `make android`. Debug APKs are only built explicitly with `make android-debug`.

Individual targets:

```sh
make
make build
make linux
make windows
make android
make server
```

Check embedded version information:

```sh
bin/tankblaster-server --version
go version -m bin/tankblaster
```

The desktop game window title is simply `Tank Blaster`.

If the Android toolchain is misbehaving:

```sh
make android-env
```

Recommended before a public release:

```sh
make test
make -n all
make all
```

## Local Configuration

Tank Blaster stores the last selected local round count, the last selected online round count, the options dialog settings, and the online identity together in `tank.cfg`.

OS-specific locations:

- Linux/FreeBSD: `$XDG_CONFIG_HOME/tankblaster/tank.cfg`, otherwise `~/.config/tankblaster/tank.cfg`
- Windows: `%AppData%\tankblaster\tank.cfg`
- macOS: `~/Library/Application Support/tankblaster/tank.cfg`
- Android: app-internal storage, typically `/data/user/0/com.runzhammer.tankblaster.android/files/tank.cfg`

## Server

Run the multiplayer server locally:

```sh
make run-server
```

Server configuration:

- Example config: [config/server.example.yaml](config/server.example.yaml)
- Docker config: [docker/server.yaml](docker/server.yaml)

Important runtime overrides:

```sh
TANKBLASTER_SERVER_ADDRESS=0.0.0.0:8765
TANKBLASTER_SERVER_PUBLIC_URL=https://tankblaster.example
TANKBLASTER_SERVER_DB=/data/tankblaster.db
TANKBLASTER_SERVER_ALLOWED_ORIGINS=https://tankblaster.example
TANKBLASTER_SERVER_MAX_CONNECTIONS=256
TANKBLASTER_SERVER_MAX_SESSIONS=128
TANKBLASTER_SERVER_MAX_QUEUE_LENGTH=256
```

`TANKBLASTER_SERVER_ALLOWED_ORIGINS` is a comma-separated list of browser WebSocket origins. It can stay empty for native clients. The server also uses HTTP timeouts, WebSocket read limits, connection/session/queue limits, and escaped invite pages.

## Docker / Podman

Build the server image:

```sh
make server-docker
```

Export the server image as a deployment tarball:

```sh
make server-docker-tar
```

The Docker targets automatically use Docker, Podman, `distrobox-host-exec podman`, or host Podman through `/app/bin/host-spawn` when running inside this Distrobox setup.

Docker Compose:

```sh
cd docker
cp .env.example .env
docker compose up -d
```

Host-network deployment, suitable for a root server with TLS/reverse proxy or firewall in front:

```sh
cd docker
cp .env.example .env
docker compose -f docker-compose.host.yml up -d
```

The Compose files run the server with reduced privileges, a read-only root filesystem, `cap_drop: ALL`, `no-new-privileges`, tmpfs for `/tmp`, and persistent SQLite storage. The host-network variant uses `/root/docker/tankblaster/data` as the default data directory.
