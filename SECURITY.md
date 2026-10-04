# Security Policy

## Supported Versions

Security fixes are expected to target the current public release and the current `main` branch. Older experimental builds are not supported.

## Reporting a Vulnerability

Please do not open a public issue for a suspected vulnerability.

Use GitHub's private vulnerability reporting feature if it is enabled for the repository. If that is not available, contact the maintainer privately through the contact information listed on the GitHub project page.

Helpful reports include:

- affected version or commit
- deployment context, especially whether the server is reachable from the internet
- whether the issue affects the desktop game, Android app, multiplayer server, Docker image, or release packaging
- reproduction steps
- expected and actual behavior
- logs or packet captures, if they do not contain secrets

Please do not include private player tokens, database dumps, server credentials, or `.env` files unless explicitly requested through a private channel.

## Server Deployment Notes

The multiplayer server is designed to run with limited privileges. The Docker Compose examples use a read-only root filesystem, `cap_drop: ALL`, `no-new-privileges`, tmpfs for `/tmp`, and configurable connection/session limits.

For public deployments, run the server behind TLS and set a correct `TANKBLASTER_SERVER_PUBLIC_URL`. Keep `TANKBLASTER_SERVER_ALLOWED_ORIGINS` empty for native clients only, or set it to explicit browser origins if a browser client is introduced.

The server stores persistent player identity and leaderboard data in SQLite. Treat the database as private operational data and keep it out of public commits and release archives.
