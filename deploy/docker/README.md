# Docker Example (Optional)

Docker/container usage is intentionally secondary in this repository.
The primary usage is as a Go library (`pkg/cep`).

## Files
- `Dockerfile`
- `docker-compose.yaml`
- `deploy.sh`

## Quick run (published image)
```bash
docker run --name gocep --rm -p 8080:8080 cssbruno/gocep:latest
```

## Compose run (local helper)
From repository root:
```bash
make compose
```

Or directly:
```bash
sh deploy/docker/deploy.sh
```
