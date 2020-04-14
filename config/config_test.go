package config_test

import (
	"os"
	"testing"

	"github.com/gojektech/consul-envoy-xds/config"
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

func TestCorsConfig(t *testing.T) {
	_ = os.Setenv("CORS_ALLOW_ORIGIN", "*")
	_ = os.Setenv("CORS_ALLOW_METHODS", "POST,OPTIONS")
	_ = os.Setenv("CORS_ALLOW_HEADERS", "keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web")
	_ = os.Setenv("CORS_EXPOSE_HEADERS", "grpc-status,grpc-message")
	_ = os.Setenv("ENABLE_CORS", "true")
	defer os.Unsetenv("CORS_ALLOW_ORIGIN")
	defer os.Unsetenv("CORS_ALLOW_METHODS")
	defer os.Unsetenv("CORS_ALLOW_HEADERS")
	defer os.Unsetenv("CORS_EXPOSE_HEADERS")
	defer os.Unsetenv("ENABLE_CORS")

	expectedCorsConfig := config.CorsConfig{
		AllowOrigin:   []string{"*"},
		AllowMethods:  "POST,OPTIONS",
		AllowHeaders:  "keep-alive,user-agent,cache-control,content-type,content-transfer-encoding,x-accept-content-transfer-encoding,x-accept-response-streaming,x-user-agent,x-grpc-web",
		ExposeHeaders: "grpc-status,grpc-message",
		Enabled:       true,
	}

	cfg := config.Load()

	assert.Equal(t, &expectedCorsConfig, cfg.GetCorsConfig())
}

func TestDefaultValuesForCorsConfig(t *testing.T) {
	expectedCorsConfig := config.CorsConfig{
		AllowOrigin:   []string{""},
		AllowMethods:  "",
		AllowHeaders:  "",
		ExposeHeaders: "",
		Enabled:       false,
	}

	cfg := config.Load()

	assert.Equal(t, &expectedCorsConfig, cfg.GetCorsConfig())
}
