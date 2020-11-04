package eds

import "github.com/gojek/consul-envoy-xds/config"

type Service struct {
	Name                 string
	Whitelist            []string
	CircuirBreakerConfig config.CircuitBreakerConfig
}
