FROM golang:1.12.9-alpine AS build-env

RUN apk add --no-cache make git gcc musl-dev
ADD . /go/src/github.com/gocardless/pgbouncer_exporter
WORKDIR /go/src/github.com/gocardless/pgbouncer_exporter
RUN PREFIX=/go/bin/ make

FROM alpine:3.11.5

RUN apk add --no-cache curl
WORKDIR /app
COPY --from=build-env /go/bin/pgbouncer_exporter /

USER postgres
EXPOSE 9127
ENTRYPOINT [ "/pgbouncer_exporter" ]
