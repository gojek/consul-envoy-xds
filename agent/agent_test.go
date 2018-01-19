package agent_test

import (
	"github.com/gojektech/consul-envoy-xds/agent"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldUseDatacenterAsRegionForTheControlPlaneEndpoint(t *testing.T) {
	a := agent.NewAgent("localhost:8500", "", "dc1")
	locality := a.Locality()
	assert.Equal(t, "dc1", locality.Region)
}

func TestShouldHaveDCAndTokenInWatcherParams(t *testing.T) {
	a := agent.NewAgent("localhost:8500", "Foo", "dc1")
	assert.Equal(t, map[string]string{"datacenter": "dc1", "token": "Foo"}, a.WatchParams())
}
