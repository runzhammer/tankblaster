# Tank Blaster

![License: GPL-3.0](https://img.shields.io/badge/License-GPL--3.0-blue.svg)
![AI-assisted development: about 90%](https://img.shields.io/badge/AI--assisted%20development-about%2090%25-purple)

Tank Blaster ist ein kleines Artillerie-Spiel mit großem Selbstbewusstsein: Panzer, Wind, krumme Flugbahnen, fragwürdige Kaufentscheidungen im Shop und genau genug Chaos, damit der perfekte Schuss sich wie Wissenschaft anfühlt.

Kurz gesagt: Spieler wählen Panzer, kaufen Zeug, zielen sorgfältig und machen danach so, als sei die Physik schuld gewesen.

Dieses Repository ist die Go-/Ebiten-Neuimplementierung von Tank Blaster mit Desktop-, Android- und Online-Multiplayer-Builds.

> **AI-assisted development**
>
> Etwa 90% der Entwicklung dieses Remakes entstanden mit KI-Unterstützung. Architektur, Entscheidungen, Tests, Review und Veröffentlichung bleiben menschlich verantwortet.

## Repository

- Lizenz: [GPL-3.0](LICENSE)
- Änderungen: [CHANGELOG.md](CHANGELOG.md)
- Beiträge: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security Policy: [SECURITY.md](SECURITY.md)

## Assets

Die ursprünglichen Tank-Blaster-Grafiken und Sounds sind mit Erlaubnis des Originalautors Axel Lauer enthalten. Die enthaltenen Musikstücke stammen aus Public-Domain-Bibliotheksmaterial aus den 1990er-/2000er-Jahren.

Neue Assets sollen nur aufgenommen werden, wenn Herkunft und Lizenz eindeutig dokumentiert und mit diesem Repository vereinbar sind.

## Entwicklung

Voraussetzung: Go gemäß [go.mod](go.mod). Für Windows-Builds wird `rsrc` automatisch über `go run` genutzt. Für Android siehe [docs/android.md](docs/android.md).

```sh
make run
```

Nützliche Dev-Kommandos:

```sh
make test
make live-test
make live-test-auto
make live-test-cp
```

Die Live-Tests starten lokal einen Multiplayer-Server und zwei Clients. Standardmäßig laufen sie über 5 Runden:

```sh
make live-test-auto LIVE_TEST_ROUNDS=3
```

Nützliche Debug-Overrides:

```sh
TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=intro make run
TANKBLASTER_DEBUG_ENABLED=1 TANKBLASTER_DEBUG_START_SCENE=online make run
```

Der normale Spielstart läuft über die Intro-Scene. Die dafür verwendeten Assets liegen unter `resources/intro/`.

## Build

Die Release-Version steht in [VERSION](VERSION). Desktop-Binaries, Server-Binaries und die Android-App betten diese Version zusammen mit Git-Commit und Build-Zeit ein.
Fuer Android muss `tankblasterVersionCode` in [android/gradle.properties](android/gradle.properties) bei jedem veroeffentlichten APK monoton erhoeht werden.

Release-Builds stellen vor dem Bauen sicher, dass `resources/config.yaml` mit `debug.enabled: false` eingebettet wird. Falls die Datei dafuer temporaer geaendert werden muss, meldet Make das transparent und stellt den vorherigen Inhalt danach wieder her.

Alles fuer ein Release bauen:

```sh
make all
```

`make all` erzeugt:

- `dist/tankblaster-VERSION-linux.zip`
- `dist/tankblaster-VERSION-windows.zip`
- `dist/tankblaster-VERSION-android.zip`
- `dist/tankblaster-server-VERSION-linux.tar`

Android wird dabei ausschliesslich als Release-APK gebaut. Debug-APKs entstehen nur mit `make android-debug`.

Einzelziele:

```sh
make
make build
make linux
make windows
make android
make server
```

Version pruefen:

```sh
bin/tankblaster-server --version
```

Das Desktop-Spiel zeigt die Version im Fenstertitel. Die eingebetteten Go-Buildinformationen lassen sich ausserdem mit `go version -m bin/tankblaster` auslesen.

Falls die Android-Toolchain zickt:

```sh
make android-env
```

Vor einem öffentlichen Release sinnvoll:

```sh
make test
make -n all
make all
```

## Lokale Einstellungen

Tank Blaster speichert die zuletzt gewaehlte lokale Rundenzahl, die zuletzt gewaehlte Online-Rundenzahl, die Optionen aus dem Optionsdialog und die Online-Identitaet gemeinsam in `tank.cfg`.

OS-spezifische Ablageorte:

- Linux/FreeBSD: `$XDG_CONFIG_HOME/tankblaster/tank.cfg`, sonst `~/.config/tankblaster/tank.cfg`
- Windows: `%AppData%\tankblaster\tank.cfg`
- macOS: `~/Library/Application Support/tankblaster/tank.cfg`
- Android: app-interner Speicher, typischerweise `/data/user/0/com.runzhammer.tankblaster.android/files/tank.cfg`

## Server

Lokal starten:

```sh
make run-server
```

Server-Konfiguration:

- Beispiel: [config/server.example.yaml](config/server.example.yaml)
- Docker-Konfiguration: [docker/server.yaml](docker/server.yaml)

Wichtige Runtime-Overrides:

```sh
TANKBLASTER_SERVER_ADDRESS=0.0.0.0:8765
TANKBLASTER_SERVER_PUBLIC_URL=https://tankblaster.example
TANKBLASTER_SERVER_DB=/data/tankblaster.db
TANKBLASTER_SERVER_ALLOWED_ORIGINS=https://tankblaster.example
TANKBLASTER_SERVER_MAX_CONNECTIONS=256
TANKBLASTER_SERVER_MAX_SESSIONS=128
TANKBLASTER_SERVER_MAX_QUEUE_LENGTH=256
```

`TANKBLASTER_SERVER_ALLOWED_ORIGINS` ist eine kommagetrennte Liste fuer Browser-WebSocket-Origins. Fuer native Clients kann sie leer bleiben. Der Server hat ausserdem HTTP-Timeouts, WebSocket-Read-Limits, Session-/Queue-Limits und escaped die Invite-Seite.

## Docker / Podman

Image bauen:

```sh
make server-docker
```

Image als Tar fuer Deployment exportieren:

```sh
make server-docker-tar
```

Die Docker-Targets nutzen automatisch Docker, Podman, `distrobox-host-exec podman` oder in dieser Distrobox Host-Podman via `/app/bin/host-spawn`.

Docker Compose:

```sh
cd docker
cp .env.example .env
docker compose up -d
```

Host-Network-Deployment, passend fuer einen Root-Server mit vorgeschaltetem TLS/Reverse-Proxy oder Firewall:

```sh
cd docker
cp .env.example .env
docker compose -f docker-compose.host.yml up -d
```

Die Compose-Dateien laufen mit reduzierten Rechten, read-only Root-Dateisystem, `cap_drop: ALL`, `no-new-privileges`, tmpfs fuer `/tmp` und persistentem SQLite-Speicher. Die Host-Network-Variante nutzt standardmaessig `/root/docker/tankblaster/data` als Datenverzeichnis.
