package config

type StatsDConfig struct {
	appName              string
	flushPeriodInSeconds int
	port                 int
	enabled              bool
}

func NewStatsDConfig(appName string, flushPeriodInSeconds, port int, enabled bool) *StatsDConfig {
	return &StatsDConfig{
		appName:              appName,
		flushPeriodInSeconds: flushPeriodInSeconds,
		port:                 port,
		enabled:              enabled,
	}
}

func (sdc *StatsDConfig) AppName() string {
	return sdc.appName
}

func (sdc *StatsDConfig) FlushPeriodInSeconds() int {
	return sdc.flushPeriodInSeconds
}

func (sdc *StatsDConfig) Port() int {
	return sdc.port
}

func (sdc *StatsDConfig) Enabled() bool {
	return sdc.enabled
}
