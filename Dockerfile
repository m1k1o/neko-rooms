#
# STAGE 1: build static web files
#
FROM node:14 as frontend
WORKDIR /src

#
# install dependencies
COPY client/package*.json ./
RUN npm install

#
# build client
COPY client/ .
RUN npm run build

#
# STAGE 2: build executable binary
#
FROM golang:1.18-buster as builder
WORKDIR /app

COPY . .
RUN go get -v -t -d .; \
    CGO_ENABLED=0 go build -o bin/neko_rooms cmd/neko_rooms/main.go

#
# STAGE 3: build a small image
#
FROM scratch
COPY --from=builder /app/bin/neko_rooms /app/bin/neko_rooms
COPY --from=frontend /src/dist/ /var/www

ENV DOCKER_API_VERSION=1.39
ENV NEKO_ROOMS_BIND=:8080
ENV NEKO_ROOMS_ADMIN_STATIC=/var/www

EXPOSE 8080

ENTRYPOINT [ "/app/bin/neko_rooms" ]
CMD [ "serve" ]
