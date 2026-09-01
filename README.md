# Tank Blaster

Tank Blaster ist ein kleines Artillerie-Spiel mit großem Selbstbewusstsein: Panzer, Wind, krumme Flugbahnen, fragwürdige Kaufentscheidungen im Shop und genau genug Chaos, damit der perfekte Schuss sich wie Wissenschaft anfühlt.

Kurz gesagt: Spieler wählen Panzer, kaufen Zeug, zielen sorgfältig und machen danach so, als sei die Physik schuld gewesen.

## Entwicklung

Voraussetzung: Go gemäß [go.mod](go.mod).

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

## Build

Linux/Desktop:

```sh
make
make build
```

Headless Multiplayer-Server:

```sh
make server
```

Windows:

```sh
make windows
```

Android:

```sh
make android
make android-release
```

Falls die Android-Toolchain zickt:

```sh
make android-env
```

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
```

## Docker / Podman

Image bauen:

```sh
make server-docker
```

Das Target nutzt automatisch Docker, Podman oder in einer Distrobox `distrobox-host-exec podman`.

Docker Compose:

```sh
cd docker
cp .env.example .env
docker compose up -d
```

Die Compose-Datei läuft mit reduzierten Rechten, read-only Root-Dateisystem und persistentem SQLite-Volume.
