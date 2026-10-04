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
make live-test-auto
make live-test-cp
make linux
make windows
make android
make server
```

Release builds temporarily force `resources/config.yaml` to use `debug.enabled: false` and restore the previous file contents after the build.

## Pull Requests

- Keep changes focused and small enough to review.
- Run relevant tests before opening a pull request.
- Include a short description of user-visible behavior changes.
- Update `README.md` or `CHANGELOG.md` when changing build, deployment, configuration, or release behavior.

## Assets

Do not add third-party graphics, sounds, fonts, or music unless their license is clear and compatible with the repository license. Include attribution and license notes in the pull request.

The original Tank Blaster graphics and sounds are included with permission from Axel Lauer. The bundled module music comes from public-domain library material from the 1990s/2000s.

## Style

- Follow the existing Go package layout and naming style.
- Prefer small, explicit functions over broad rewrites.
- Keep generated build artifacts out of commits.
