package eds_test

import (
	"github.com/gojektech/consul-envoy-xds/eds"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestShouldSetCPLBAEndpoint(t *testing.T) {
	sh := eds.NewServiceHost(&api.CatalogService{ServiceAddress: "someip", ServicePort: 9091})
	ep := sh.LbEndpoint()
	assert.Equal(t, "socket_address:<address:\"someip\" port_value:9091 > ", ep.Endpoint.Address.String())
}

func TestShouldSetPipeCPLBAEndpoint(t *testing.T) {
	sh := eds.NewServiceHost(&api.CatalogService{ServiceAddress: "/tmp/tsd.26333.sock", ServicePort: 0})
	ep := sh.LbEndpoint()
	assert.Equal(t, "pipe:<path:\"/tmp/tsd.26333.sock\" > ", ep.Endpoint.Address.String())
}
