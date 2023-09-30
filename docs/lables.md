# custom labels

You can add custom labels for every room.

Example: Expose port 3000 from every room under `3000-(room-name)`:

```bash
-e "NEKO_ROOMS_INSTANCE_LABELS=
    traefik.http.services.{containerName}-3000-tcp.loadbalancer.server.port=3000
    traefik.http.routers.{containerName}-3000-tcp.entrypoints={traefikEntrypoint}
    traefik.http.routers.{containerName}-3000-tcp.rule=PathPrefix(\`/3000-{roomName}\`)
    traefik.http.middlewares.{containerName}-3000-tcp-prf.stripprefix.prefixes=/3000-{roomName}/
    traefik.http.routers.{containerName}-3000-tcp.middlewares={containerName}-3000-tcp-prf
    traefik.http.routers.{containerName}-3000-tcp.service={containerName}-3000-tcp
"
```
