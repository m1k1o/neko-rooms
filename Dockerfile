FROM golang:1.16-buster as server
WORKDIR /app

#
# install dependencies
#RUN set -eux; apt-get update; \
#    apt-get install -y --no-install-recommends;
#    #
#    # clean up
#    apt-get clean -y; \
#    rm -rf /var/lib/apt/lists/* /var/cache/apt/*

#
# build server
COPY . .

RUN go get -v -t -d .; \
    ./build

ENTRYPOINT [ "bin/neko_rooms" ]
CMD [ "serve" ]
