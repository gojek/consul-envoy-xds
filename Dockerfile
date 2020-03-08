FROM golang:1.12-alpine
RUN apk update
RUN apk add git
WORKDIR /usr/src
ADD go.mod .
RUN go mod download
ADD $PWD /usr/src/consul-envoy-xds
WORKDIR /usr/src/consul-envoy-xds
RUN go build -o out/consul-envoy-xds

FROM alpine:3.9.2
WORKDIR /opt/consul-envoy-xds
COPY --from=0 usr/src/consul-envoy-xds/out/consul-envoy-xds .
RUN apk update

ENV PORT=8053 \
	LOG_LEVEL=INFO \
	CONSUL_CLIENT_PORT=8500 \
	CONSUL_CLIENT_HOST= \
	CONSUL_DC=dc1 \
	CONSUL_TOKEN= \
	WATCHED_SERVICE=
EXPOSE ${PORT}
CMD ./consul-envoy-xds
