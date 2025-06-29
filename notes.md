try for docker:
at least if you edit the `front-end` you have: `docker compose up --build -d frontend`

```
# View all container logs in real-time
docker compose -f deploy/docker-compose.yml logs -f

# View logs for a specific service
docker compose -f deploy/docker-compose.yml logs -f api
docker compose -f deploy/docker-compose.yml logs -f frontend
docker compose -f deploy/docker-compose.yml logs -f postgres

# Check container status
docker compose -f deploy/docker-compose.yml ps
```

## make

```
make logs          # All logs
make logs-api       # API logs only
make status         # Service status
```
