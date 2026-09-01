# Tank Blaster Server Docker

Build the server image from the repository root:

```sh
make server-docker
```

Run it with Docker Compose:

```sh
cd docker
cp .env.example .env
docker compose up -d
```

Useful settings in `.env`:

- `TANKBLASTER_SERVER_PORT`: host port, defaults to `8765`
- `TANKBLASTER_SERVER_PUBLIC_URL`: public HTTP URL used for invite links
- `TANKBLASTER_SERVER_IMAGE`: image name/tag
- `TANKBLASTER_SERVER_MEM_LIMIT`: memory limit
- `TANKBLASTER_SERVER_CPUS`: CPU limit

The container listens on `0.0.0.0:8765` internally and stores SQLite data in the `tankblaster-data` Docker volume at `/data/tankblaster.db`.
