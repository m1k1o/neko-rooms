# Installation guide for NGINX as a reverse proxy with authentication

First, make sure that you have your `.htpasswd` file created, if you don't, you can follow this guide in the NGINX docs:

https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-http-basic-authentication/

This guide assumes that your `.htpasswd` file is located at `/etc/apache2/.htpasswd`.

## Installation:

Download your `docker-compose.yml` file, and make necessary changes to the environment variables, more notably, `NEKO_ROOMS_NAT1TO1` and `NEKO_ROOMS_INSTANCE_URL`.

Run your docker-compose file, done by running `sudo docker-compose up -d` on Linux in the same directory as your compose file.

Next, move onto NGINX. First, open up your NGINX config, and make any alterations necessary for the traefik `proxy-pass`;
The container's IP will also work there.
Your NGINX config at this point should be good to go, install and restart NGINX!

## Certificates:

If you wish to have SSL for Neko, you can use certbot to get that done! A guide is linked below for installation.

https://certbot.eff.org/instructions
