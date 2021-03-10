# neko-rooms

Simple room management system for [n.eko](https://github.com/m1k1o/neko).

## How to start

You need to have installed `Docker` and `docker-compose`.

### Step 1

Copy `.env.example` to `.env` and customize.

```bash
cp .env.example .env
```

### Step 2

Run start docker compose stack.

```bash
docker-compose up -d
```

### Stop 3

Use `OpenApi.yml` and send commands to your new controller.

## WARNING

This project is WIP, n.eko is not wokring on custom `/paths/`, because all static files are loaded from `/`. This needs to be fixed in upstream repository to have this instane working.

### Roadmap:
 - [ ] add authentication provider for API
 - [x] add GUI
 - [ ] add docker ssh / tcp support
 - [ ] add docker swarm support
