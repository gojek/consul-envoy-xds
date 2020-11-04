package config_test

import (
	"os"
	"testing"

	"github.com/gojek/consul-envoy-xds/config"
	"github.com/stretchr/testify/assert"
)

func TestWhiteListedRoutesConfig(t *testing.T) {
	os.Setenv("FOO_WHITELISTED_ROUTES", "foo")
	defer os.Unsetenv("FOO_WHITELISTED_ROUTES")

	cfg := config.Load()

	assert.Equal(t, []string{"foo"}, cfg.WhitelistedRoutes("foo"))
}

func TestWhiteListedRoutesConfigReplacesHyphens(t *testing.T) {
	os.Setenv("FOO_BAR_WHITELISTED_ROUTES", "foo")
	defer os.Unsetenv("FOO_BAR_WHITELISTED_ROUTES")

	cfg := config.Load()

	assert.Equal(t, []string{"foo"}, cfg.WhitelistedRoutes("foo-bar"))
}

func TestConfig_CircuitBreakerConfig(t *testing.T) {

	t.Run("MaxConnections", func(t *testing.T) {
		t.Run("Should return default value if config is missing", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.MissingEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_CONNECTIONS"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(config.DefaultCircuitBreakerMaxConnections), cfg.MaxConnections)
		})

		t.Run("Should return value from env vars", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.NewEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_CONNECTIONS", "100"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(100), cfg.MaxConnections)
		})
	})

	t.Run("MaxPendingRequests", func(t *testing.T) {
		t.Run("Should return default value if config is missing", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.MissingEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_PENDING_REQUESTS"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(config.DefaultCircuitBreakerMaxPendingRequests), cfg.MaxPendingRequests)
		})

		t.Run("Should return value from env vars", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.NewEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_PENDING_REQUESTS", "100"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(100), cfg.MaxConnections)
		})
	})

	t.Run("MaxRequests", func(t *testing.T) {

		t.Run("Should return default value if config is missing", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.MissingEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_REQUESTS"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(config.DefaultCircuitBreakerMaxRequests), cfg.MaxRequests)
		})

		t.Run("Should return value from env vars", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.NewEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_REQUESTS", "100"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(100), cfg.MaxRequests)
		})
	})

	t.Run("MaxRetries", func(t *testing.T) {
		t.Run("Should return default value if config is missing", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.MissingEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_RETRIES"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(config.DefaultCircuitBreakerMaxRetries), cfg.MaxRetries)
		})

		t.Run("Should return value from env vars", func(t *testing.T) {
			mutation := config.ApplyEnvVars(config.NewEnvVar("FOO_BAR_CIRCUIT_BREAKER_MAX_RETRIES", "5"))
			defer mutation.Rollback()

			cfg := config.Load().CircuitBreakerConfig("foo-bar")

			assert.Equal(t, uint32(5), cfg.MaxRetries)
		})
	})
}
