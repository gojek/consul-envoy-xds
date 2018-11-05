package config

import (
	"fmt"

	"strings"

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

func (cfg *Config) WatchedServices() []string {
	return strings.Split(cfg.GetValue("WATCHED_SERVICE"), ",")
}

func (cfg *Config) AppName() string {
	return cfg.GetOptionalValue("APP_NAME", "consul-envoy-xds")
}

func (cfg *Config) NewRelicLicenseKey() string {
	return cfg.GetValue("NEW_RELIC_LICENSE_KEY")
}

func (cfg *Config) NewRelicEnabled() bool {
	return cfg.GetFeature("NEW_RELIC_ENABLED")
}

func (cfg *Config) WhitelistedRoutes(svc string) []string {
	canonicalName := strings.Replace(svc, "-", "_", -1)
	whitelist := cfg.GetOptionalValue(strings.ToUpper(canonicalName)+"_WHITELISTED_ROUTES", "/")
	return strings.Split(whitelist, ",")
}
