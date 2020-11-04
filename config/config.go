package config

import (
	"fmt"
	"strings"

	"github.com/gojek-engineering/goconfig"
	"strconv"
)

const (
	DefaultCircuitBreakerMaxConnections     = 1024
	DefaultCircuitBreakerMaxRequests        = 1024
	DefaultCircuitBreakerMaxPendingRequests = 1024
	DefaultCircuitBreakerMaxRetries         = 3
)

type Config struct {
	goconfig.BaseConfig
}

type HTTPHeaderRateLimitConfig struct {
	IsEnabled     bool
	HeaderName    string
	DescriptorKey string
}

type CircuitBreakerConfig struct {
	MaxConnections     uint32
	MaxPendingRequests uint32
	MaxRequests        uint32
	MaxRetries         uint32
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

func (cfg *Config) GetHTTPHeaderRateLimitConfig() *HTTPHeaderRateLimitConfig {
	isEnableString := cfg.GetOptionalValue("HTTP_HEADER_RATE_LIMIT_ENABLED", "false")
	name := cfg.GetOptionalValue("HTTP_HEADER_RATE_LIMIT_NAME", "")
	descriptor := cfg.GetOptionalValue("HTTP_HEADER_RATE_LIMIT_DESCRIPTOR", "")
	isEnable, err := strconv.ParseBool(isEnableString)
	if err != nil {
		isEnable = false
	}

	return &HTTPHeaderRateLimitConfig{
		IsEnabled:     isEnable,
		HeaderName:    name,
		DescriptorKey: descriptor,
	}
}

func (cfg *Config) WhitelistedRoutes(svc string) []string {
	canonicalName := canonicalizeSvcName(svc)
	whitelist := cfg.GetOptionalValue(canonicalName+"_WHITELISTED_ROUTES", "/")
	return strings.Split(whitelist, ",")
}

func (cfg *Config) EnableHealthCheckCatalogService() bool {
	return cfg.GetFeature("ENABLE_HEALTH_CHECK_CATALOG_SVC")
}

func (cfg *Config) CircuitBreakerConfig(svc string) CircuitBreakerConfig {
	canonicalName := canonicalizeSvcName(svc)
	maxConnections := uint32(cfg.GetOptionalIntValue(canonicalName+"_CIRCUIT_BREAKER_MAX_CONNECTIONS", DefaultCircuitBreakerMaxConnections))
	maxPendingRequests := uint32(cfg.GetOptionalIntValue(canonicalName+"_CIRCUIT_BREAKER_MAX_PENDING_REQUESTS", DefaultCircuitBreakerMaxPendingRequests))
	maxRequests := uint32(cfg.GetOptionalIntValue(canonicalName+"_CIRCUIT_BREAKER_MAX_REQUESTS", DefaultCircuitBreakerMaxRequests))
	maxRetries := uint32(cfg.GetOptionalIntValue(canonicalName+"_CIRCUIT_BREAKER_MAX_RETRIES", DefaultCircuitBreakerMaxRetries))

	return CircuitBreakerConfig{
		MaxConnections:     maxConnections,
		MaxPendingRequests: maxPendingRequests,
		MaxRequests:        maxRequests,
		MaxRetries:         maxRetries,
	}
}

func canonicalizeSvcName(svc string) string {
	return strings.ToUpper(strings.Replace(svc, "-", "_", -1))
}
