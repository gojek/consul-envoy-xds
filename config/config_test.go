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
