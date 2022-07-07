# build
FROM golang:1.18-alpine as build

ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN cd src && go build -o /build/pepe -ldflags "-X main.version=beta -s -w"

# base
FROM ghcr.io/umputun/baseimage/app:v1.9.1 as base

# run
FROM scratch

COPY --from=build /build/pepe /srv/pepe
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/group /etc/group

WORKDIR /srv

ENTRYPOINT ["/srv/pepe"]