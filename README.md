# consul-envoy-xds [![CircleCI](https://circleci.com/gh/gojektech/consul-envoy-xds.svg?style=svg)](https://circleci.com/gh/gojektech/consul-envoy-xds)

consul-envoy-xds is an implementation of an Envoy Control Plane/xDiscovery Service via the [Envoy data plane API](https://github.com/envoyproxy/data-plane-api). It makes services registered with Consul available as upstreams through CDS, EDS and RDS.

xDS is the set of APIs that control the Envoy dynamic configuration. A longer explanation is available on the [XDS Protocol](https://github.com/envoyproxy/data-plane-api/blob/master/XDS_PROTOCOL.md) page. Currently consul-envoy-xds implements CDS, EDS and RDS. In this implementation, the streaming version is available but the sync (Unary call) one is WIP.

If you are using Consul for service discovery and would like to use Envoy without manual configuration, consul-envoy-xds can be used. It uses [Consul Watches](https://www.consul.io/docs/agent/watches.html) and any changes to endpoints are streamed to Envoy via the Control Plane.

## Building it

1. We use Dep for dependency management, instructions to install it can be found [here](https://github.com/golang/dep).
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

Locally, the services can be configured by setting the environment variables in an `application.yml` file. When this file is unavailable, configuration is loaded from the environment variables. This is the recommended way to load configuration on production.

```
PORT: 8053
LOG_LEVEL: DEBUG
CONSUL_CLIENT_PORT: 8500
CONSUL_CLIENT_HOST: localhost
CONSUL_DC: dc1
CONSUL_TOKEN: ""
WATCHED_SERVICE: foo-service,bar_svc
FOO_SERVICE_WHITELISTED_ROUTES: /foo,/fuu
BAR_SVC_WHITELISTED_ROUTES: /bar
```

For above sample configuration, `consul-envoy-xds` will setup 2 clusters viz. foo-service and bar-svc. The foo-service cluster will have two routes in a virtual host i.e. `/foo` and `/fuu`. Similarly, bar_svc will have a route `/bar` into the same virtual host.

Currently xDS server implementation configures single virtual host with routes for all upstream clusters based on <X>_WHITELISTED_ROUTES config. This implies no two services can have any whitelisted route with same prefix.

Example entry point on production environments.

`env PORT=8053 LOG_LEVEL=INFO CONSUL_AGENT_PORT=8500 CONSUL_CLIENT_HOST=localhost CONSUL_DC=dc1 CONSUL_TOKEN="" WATCHED_SERVICE=foo-service,bar_svc BAR_SVC_WHITELISTED_ROUTES='/bar' FOO_SERVICE_WHITELISTED_ROUTES='/foo,/fuu' ./consul-envoy-xds`

##### Configuring Regex Paths:

If you have url params in the routes that needs to be whitelisted, you can use use regex in the path. To specify regex path in the whitelisted routes, use this notion: `%regex:some_route`

Example:

Let's say you have REST endpoint for serving customer details with customer id in the path: `/foo/{customer-id}/details`. To whitelist this path, you can:
```
WATCHED_SERVICE: foo-service
FOO_SERVICE_WHITELISTED_ROUTES: %regex:/foo/customer/[0-9]*/details,/fuu
```

If you have more than one path which has regex:
```
WATCHED_SERVICE: foo-service
FOO_SERVICE_WHITELISTED_ROUTES: %regex:/foo/customer/[0-9]*/details,%regex:/fuu/[0-9A-Z]*/id
```

refer to [Regex documentation](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route.proto#envoy-api-field-route-routematch-regex) for the regex pattern

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
                route: { cluster: foo-service }
          http_filters:
          - name: envoy.router
  clusters:
  - name: foo-service
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

## Using it with Rate Limitter

Example config with HTTP Header Rate Limit:
```
HTTP_HEADER_RATE_LIMIT_ENABLED: true
HTTP_HEADER_RATE_LIMIT_DESCRIPTOR: "user_id"
HTTP_HEADER_RATE_LIMIT_NAME: "User-Id"
```

For above sample configuration, `consul-envoy-xds` will add ratelimit in returned routes configuration based on Http Header. Later you need to implement envoy global gRPC rate limiting service. Refer to [Envoy Rate Limit](https://github.com/lyft/ratelimit).

#### Sample Config with Rate Limit:

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
                route: { cluster: foo-service }
          http_filters:
          - name: envoy.router
  clusters:
  - name: foo-service
    connect_timeout: 0.25s
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    type: EDS
    eds_cluster_config:
      eds_config:
        api_config_source:
          api_type: GRPC
          cluster_names: [xds_cluster]

  - name: rate_limit_cluster
    connect_timeout: 0.250s
    http2_protocol_options: {}
    hosts:
    - socket_address:
      address: $RATE_LIMIT_IP
      port_value: $RATE_LIMIT_PORT
    dns_lookup_family: V4_ONLY
    health_checks:
    - timeout:
        seconds: 1
      interval:
        seconds: 1
      unhealthy_threshold: 3
      healthy_threshold: 3
      grpc_health_check:
        service_name: $RATE_LIMIT_SERVICE_NAME

  - name: xds_cluster
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    http2_protocol_options: {}
    hosts: [{ socket_address: { address: $XDS_IP, port_value: $XDS_PORT }}]

rate_limit_service:
  grpc_service:
    envoy_grpc:
      cluster_name: rate_limit_cluster
    timeout: 0.25s

admin:
  access_log_path: /dev/null
  address:
    socket_address: { address: 127.0.0.1, port_value: 9901 }
```

#### Enable Health Check Filter

Example config to enable health check filter
```
ENABLE_HEALTH_CHECK_CATALOG_SVC: true
```

Only discover catalog service endpoints with health check status `passed`