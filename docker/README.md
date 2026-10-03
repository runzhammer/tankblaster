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

Run it on a server that intentionally uses host networking:

```sh
cd docker
cp .env.example .env
docker compose -f docker-compose.host.yml up -d
```

Useful settings in `.env`:

- `TANKBLASTER_SERVER_PORT`: host port, defaults to `8765`
- `TANKBLASTER_SERVER_PUBLIC_URL`: public HTTP URL used for invite links
- `TANKBLASTER_SERVER_IMAGE`: image name/tag
- `TANKBLASTER_SERVER_DATA_DIR`: host directory for `/data/tankblaster.db` in the host-network compose file
- `TANKBLASTER_SERVER_ALLOWED_ORIGINS`: comma-separated browser WebSocket origins; keep empty for native clients only
- `TANKBLASTER_SERVER_MAX_CONNECTIONS`: maximum concurrent WebSocket connections
- `TANKBLASTER_SERVER_MAX_SESSIONS`: maximum open sessions
- `TANKBLASTER_SERVER_MAX_QUEUE_LENGTH`: maximum matchmaking queue length
- `TANKBLASTER_SERVER_MEM_LIMIT`: memory limit
- `TANKBLASTER_SERVER_CPUS`: CPU limit

The default compose file publishes port `8765` and stores SQLite data in the `tankblaster-data` Docker volume at `/data/tankblaster.db`.
The host-network compose file binds the server directly to `0.0.0.0:8765` on the host and stores SQLite data in `TANKBLASTER_SERVER_DATA_DIR`.
