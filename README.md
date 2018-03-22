# consul-envoy-xds

consul-envoy-xds is an implementation of an Envoy Control Plane/xDiscovery Service via the [Envoy data plane API](https://github.com/envoyproxy/data-plane-api). It makes services registered with Consul available as endpoints through [EDS](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/eds.proto.html).

xDS is the set of APIs that control the Envoy dynamic configuration. A longer explanation is available on the [XDS Protocol](https://github.com/envoyproxy/data-plane-api/blob/master/XDS_PROTOCOL.md) page. Currently consul-envoy-xds implements only EDS, however other xDS are planned. In this implementation, the streaming version is available but the sync (Unary call) one is WIP.

If you are using Consul for service discovery and would like to use Envoy without manual configuration, consul-envoy-xds can be used. It uses [Consul Watches](https://www.consul.io/docs/agent/watches.html) and any changes to endpoints are streamed to Envoy via the Control Plane.

## Building it

1. We use Glide for dependency management, instructions to install it can be found [here](http://glide.sh/).
2. Fetch dev dependencies and create dev config (application.yml) from sample.

   ```
   make setup
   ```
3. Run `make` to fetch dependencies, run the tests and build.

    ```
    make
    ```
4. Run it.

    ```
    ./out/consul-envoy-xds
    ```

## Using it

Locally, the service can be configured by setting the environment variables in an `application.yml` file. When this file is unavailable, configuration is loaded from the environment variables. This is the recommended way to load configuration on production.

```
PORT: 8053
LOG_LEVEL: DEBUG
CONSUL_AGENT_PORT: 8500
CONSUL_CLIENT_HOST: localhost
CONSUL_DC: dc1
CONSUL_TOKEN: ""
WATCHED_SERVICE: foo-service
```

Example entry point on production environments.

`env PORT=8053 LOG_LEVEL=INFO CONSUL_AGENT_PORT=8500 CONSUL_CLIENT_HOST=localhost CONSUL_DC=dc1 CONSUL_TOKEN="" WATCHED_SERVICE=foo-service ./consul-envoy-xds`


#### Sample Config:

Replace `$XDS\_IP` and `$XDS\_PORT` with wherever you're running consul-envoy-xds. Refer to [Envoy documentation](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api) for other options.

```yaml
static_resources:
  listeners:
  - name: listener_0
    address:
      socket_address: { address: 0.0.0.0, port_value: 443 }
    filter_chains:
    - filters:
      - name: envoy.http_connection_manager
        config:
          stat_prefix: ingress_http
          codec_type: AUTO
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                route: { cluster: foo-svc }
          http_filters:
          - name: envoy.router
  clusters:
  - name: foo-svc
    connect_timeout: 0.25s
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    type: EDS
    eds_cluster_config:
      eds_config:
        api_config_source:
          api_type: GRPC
          cluster_names: [xds_cluster]
  - name: xds_cluster
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    hosts: [{ socket_address: { address: $XDS_IP, port_value: $XDS_PORT }}]
admin:
  access_log_path: /dev/null
  address:
    socket_address: { address: 127.0.0.1, port_value: 9901 }
```
