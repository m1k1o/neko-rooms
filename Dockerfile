# Step 1: build executable binary
FROM golang:1.16-buster as builder
WORKDIR /app

COPY . .
RUN go get -v -t -d .; \
    go build -o bin/neko_rooms cmd/neko_rooms/main.go

# Step 2: build a small image
#FROM scratch
#COPY --from=builder /app/bin/neko_rooms /app/bin/neko_rooms

ENV DOCKER_API_VERSION=1.39
ENV NEKO_ROOMS_BIND=:8080
EXPOSE 8080

ENTRYPOINT [ "/app/bin/neko_rooms" ]
CMD [ "serve" ]
