package config

import (
	"fmt"
	"github.com/gojektech/consul-envoy-xds/agent"
	"github.com/gojektech/consul-envoy-xds/eds"

	"github.com/gojek-engineering/goconfig"
)

type Config struct {
	goconfig.BaseConfig
}

func Load() *Config {
	cfg := &Config{}
	cfg.LoadWithOptions(map[string]interface{}{"newrelic": false, "db": false})
	return cfg
}

func (cfg *Config) Port() int {
	return cfg.GetIntValue("PORT")
}

func (cfg *Config) LogLevel() string {
	return cfg.GetValue("LOG_LEVEL")
}

func (cfg *Config) ConsulClientHost() string {
	return cfg.GetValue("CONSUL_CLIENT_HOST")
}

func (cfg *Config) ConsulClientPost() int {
	return cfg.GetIntValue("CONSUL_CLIENT_PORT")
}

func (cfg *Config) ConsulToken() string {
	return cfg.GetValue("CONSUL_TOKEN")
}

func (cfg *Config) ConsulAddress() string {
	return fmt.Sprintf("%s:%d", cfg.ConsulClientHost(), cfg.ConsulClientPost())
}

func (cfg *Config) ConsulDC() string {
	return cfg.GetValue("CONSUL_DC")
}

func (cfg *Config) WatchedServiceName() string {
	return cfg.GetValue("WATCHED_SERVICE")
}

func (cfg *Config) WatchedService() eds.Endpoint {
	return eds.NewEndpoint(cfg.WatchedServiceName(), agent.NewAgent(cfg.ConsulAddress(), cfg.ConsulToken(), cfg.ConsulDC()))
}
