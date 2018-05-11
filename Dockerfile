FROM golang:1.9.6-stretch
RUN curl https://glide.sh/get | sh
ADD . /go/src/github.com/gojektech/consul-envoy-xds
WORKDIR /go/src/github.com/gojektech/consul-envoy-xds
RUN make

FROM alpine:3.7
RUN mkdir /lib64; \
	ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=0 /go/src/github.com/gojektech/consul-envoy-xds/out/consul-envoy-xds /etc/bin/
ENV PORT=8053 \
	LOG_LEVEL=INFO \
	CONSUL_CLIENT_PORT=8500 \
	CONSUL_CLIENT_HOST= \
	CONSUL_DC=dc1 \
	CONSUL_TOKEN= \
	WATCHED_SERVICE=
EXPOSE ${PORT}
ENTRYPOINT ["/etc/bin/consul-envoy-xds"]