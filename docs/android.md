# Android Build

## Voraussetzungen

- Go passend zum Projekt
- JDK 21, lokal z. B. unter `/home/reznor/.local/share/tankblaster-android/jdk-21.0.12.1+1`
- Android SDK unter `/home/reznor/Android/Sdk`
- Android Platform SDK 36
- Android Build Tools 36.0.0
- Android NDK 27.3.13750724
- `ebitenmobile` passend zur Ebitengine-Version aus `go.mod`

## Build

```bash
./scripts/build-android.sh
```

Alternativ per Makefile:

```bash
make android-debug
make android-release
```

Das Script erzeugt zuerst `android/app/libs/tankblaster.aar`, baut danach die Android-App mit Gradle und kopiert die APK nach:

```text
dist/tankblaster-debug.apk
dist/tankblaster-release.apk
```

## Installation

```bash
/home/reznor/Android/Sdk/platform-tools/adb install -r dist/tankblaster-debug.apk
```

## Lokale Einstellungen

Unter Android liegt die gemeinsame YAML-Datei `tank.cfg` fuer lokale Einstellungen und Online-Identitaet im app-internen Dateispeicher:

```text
/data/user/0/com.runzhammer.tankblaster.android/files/tank.cfg
```

Der Pfad ist privat fuer die App, braucht keine Storage-Permission und wird bei Deinstallation der App entfernt.
