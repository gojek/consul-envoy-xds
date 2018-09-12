package instrument

import (
	"log"

	"github.com/gojektech/consul-envoy-xds/config"
	"github.com/newrelic/go-agent"
)

var newRelicApp newrelic.Application

func init() {
	cfg := config.Load()
	newRelicCfg := newrelic.NewConfig(cfg.AppName(), cfg.NewRelicLicenseKey())
	newRelicCfg.Enabled = cfg.NewRelicEnabled()
	var err error
	newRelicApp, err = newrelic.NewApplication(newRelicCfg)
	if err != nil {
		log.Fatal("failed to initialize new relic application", err)
	}
}

func NewRelicApp() newrelic.Application {
	return newRelicApp
}
