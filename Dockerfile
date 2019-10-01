FROM golang:1.13.1-alpine AS build-env

RUN apk add --no-cache make git gcc musl-dev
ADD . /go/src/github.com/gocardless/pgbouncer_exporter
WORKDIR /go/src/github.com/gocardless/pgbouncer_exporter
RUN PREFIX=/go/bin/ make

FROM alpine:3.8

RUN apk add --no-cache curl
WORKDIR /app
COPY --from=build-env /go/bin/pgbouncer_exporter /

USER postgres
EXPOSE 9127
ENTRYPOINT [ "/pgbouncer_exporter" ]
