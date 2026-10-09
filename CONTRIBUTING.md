# Contributing to Tank Blaster

Thanks for taking the time to improve Tank Blaster.

## Development Setup

- Install the Go version declared in `go.mod`.
- For Android builds, follow `docs/android.md`.
- Run the desktop game with:

```sh
make run
```

- Run tests with:

```sh
make test
```

## Useful Targets

```sh
make live-test
make linux
make windows
make android
make server
```

Release builds temporarily force `resources/config.yaml` to use `debug.enabled: false` and restore the previous file contents after the build.

Before opening a pull request, prefer running:

```sh
make test
make -n all
```

Use `make all` for release verification when the Android toolchain and Docker/Podman are available.

## Pull Requests

- Keep changes focused and small enough to review.
- Run relevant tests before opening a pull request.
- Include a short description of user-visible behavior changes.
- Update `README.md` or `CHANGELOG.md` when changing build, deployment, configuration, or release behavior.
- Do not commit generated binaries, release zips, local databases, `.env` files, or credentials.

## AI-Assisted Contributions

This project openly documents that a large part of the remake was developed with AI assistance. AI-assisted patches are welcome, but contributors remain responsible for the code they submit.

Please review generated code carefully, keep prompts or notes available when they clarify a change, and make sure tests and licensing obligations still make sense.

## Assets

Do not add third-party graphics, sounds, fonts, or music unless their license is clear and compatible with the repository license. Include attribution and license notes in the pull request.

The original Tank Blaster graphics and sounds are included with permission from Axel Lauer. The bundled module music comes from public-domain library material from the 1990s/2000s.

If an asset license is unclear, leave the asset out until provenance is resolved.

## Style

- Follow the existing Go package layout and naming style.
- Prefer small, explicit functions over broad rewrites.
- Keep generated build artifacts out of commits.
- Use `gofmt` for Go files.
- Prefer repository-local helpers and patterns over new abstractions.
