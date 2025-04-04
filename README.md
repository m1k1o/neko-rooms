# neko-rooms

<p align="center">
  <img src="https://img.shields.io/github/v/release/m1k1o/neko-rooms" alt="release">
  <img src="https://img.shields.io/github/license/m1k1o/neko-rooms" alt="license">
  <img src="https://img.shields.io/docker/pulls/m1k1o/neko-rooms" alt="pulls">
  <img src="https://img.shields.io/github/issues/m1k1o/neko-rooms" alt="issues">
  <a href="https://discord.gg/3U6hWpC" ><img src="https://discordapp.com/api/guilds/665851821906067466/widget.png" alt="Chat on discord"></a>
</p>

Simple room management system for [n.eko](https://github.com/m1k1o/neko). Self hosted rabb.it alternative.

<div align="center">
  <img src="https://github.com/m1k1o/neko-rooms/raw/master/docs/rooms.png" alt="rooms">
  <img src="https://github.com/m1k1o/neko-rooms/raw/master/docs/new_room.png" alt="new room">
  <img src="https://github.com/m1k1o/neko-rooms/raw/master/docs/neko.gif" alt="n.eko">
</div>

## Zero-knowledge installation (with HTTPS)

No experience with Docker and reverse proxy? No problem! Follow these steps to set up your Neko Rooms quickly and securely:

- Rent a VPS with public IP and OS Ubuntu.
- Get a domain name pointing to your IP (you can even get some for free).
- Run install script and follow instructions.
- Secure using HTTPs thanks to Let's Encrypt and Traefik or NGINX.

```bash
wget -O neko-rooms-traefik.sh https://raw.githubusercontent.com/m1k1o/neko-rooms/master/traefik/install
sudo bash neko-rooms-traefik.sh
```

### Community Installation Scripts

We have community-contributed installation scripts available. Check out our [community installation guides](./community/README.md) for instructions on installing neko-rooms on various Linux distributions. These scripts are maintained by the community and support different Linux distributions like Arch Linux, Fedora, and more.

## How to start

If you want to use Traefik as reverse proxy, visit [installation guide for traefik as reverse proxy](./traefik/).

Otherwise modify variables in `docker-compose.yml` and just run `docker-compose up -d`.

### Download images / update

You need to pull all your images, that you want to use with neko-room. Otherwise, you might get this error: `Error response from daemon: No such image:` (see issue #1).

```sh
docker pull ghcr.io/m1k1o/neko/firefox
docker pull ghcr.io/m1k1o/neko/chromium
# etc...
```

If you want to update neko image, you need to pull new image and recreate all rooms, that use old image. To update neko rooms, simply run:

```sh
docker-compose pull
docker-compose up -d
```

### Enable storage

You might have encountered this error:

> Mounts cannot be specified because storage is disabled or unavailable.

If you didn't specify storage yet, you can do it using [this tutorial](./docs/storage.md).

### Use nvidia GPU

If you want to use nvidia GPU, you need to install [nvidia-docker](https://github.com/NVIDIA/nvidia-docker).

Change neko images to nvidia images in `docker-compose.yml` using envorinment variable `NEKO_ROOMS_NEKO_IMAGES`:

```bash
NEKO_ROOMS_NEKO_IMAGES="
  ghcr.io/m1k1o/neko/nvidia-chromium:latest
  ghcr.io/m1k1o/neko/nvidia-google-chrome:latest
  ghcr.io/m1k1o/neko/nvidia-microsoft-edge:latest
  ghcr.io/m1k1o/neko/nvidia-brave:latest
"
```

When creating new room, you need to specify to use GPU in expext settings.

### Docs

For more information visit [docs](./docs).

### Roadmap:
 - [x] add GUI
 - [x] add HTTPS support
 - [x] add authentication provider for traefik
 - [x] allow specifying custom ENV variables
 - [x] allow mounting directories for persistent data
 - [x] optionally remove Traefik as dependency
 - [ ] add upgrade button
 - [ ] auto pull images, that do not exist
 - [ ] add bearer token to for API
 - [ ] add docker SSH / TCP support
 - [ ] add docker swarm support
 - [ ] add k8s support
