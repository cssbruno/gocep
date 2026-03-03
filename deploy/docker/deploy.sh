#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.yaml"

compose() {
	if docker compose version >/dev/null 2>&1; then
		docker compose -f "${COMPOSE_FILE}" "$@"
		return
	fi
	docker-compose -f "${COMPOSE_FILE}" "$@"
}

echo -e "\033[0;33m######################################### pull ########################################\033[0m"
docker pull cssbruno/gocep

compose stop gocep || true
compose rm -f gocep || true
compose up -d gocep
compose ps
echo -e "\033[0;32mGenerated Run docker compose\033[0m \033[0;33m[ok]\033[0m \n"
