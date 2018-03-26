package eds

import (
	cpcore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	eds "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/hashicorp/consul/api"
)

//ServiceHost represents a single host for a service
type ServiceHost struct {
	Service     string
	IPAddress   string
	Port        int
	Tags        []string
	CreateIndex uint64
	ModifyIndex uint64
}

// LbEndpoint translates a consul agent service endpoint to an envoy control plane LbEndpoint.
// If the ServiceHost's port is < 1, the endpoint is assumed to be a Pipe and the IPAddress
// represents the Pipe's Path.
func (h ServiceHost) LbEndpoint() eds.LbEndpoint {
	if uint32(h.Port) > 0 {
		return eds.LbEndpoint{
			HealthStatus: cpcore.HealthStatus_HEALTHY,
			Endpoint: &eds.Endpoint{
				Address: &cpcore.Address{
					Address: &cpcore.Address_SocketAddress{
						SocketAddress: &cpcore.SocketAddress{
							Protocol: cpcore.TCP,
							Address:  h.IPAddress,
							PortSpecifier: &cpcore.SocketAddress_PortValue{
								PortValue: uint32(h.Port),
							},
						},
					},
				}}}
	}
	return eds.LbEndpoint{
		HealthStatus: cpcore.HealthStatus_HEALTHY,
		Endpoint: &eds.Endpoint{
			Address: &cpcore.Address{
				Address: &cpcore.Address_Pipe{
					Pipe: &cpcore.Pipe{
						Path: h.IPAddress,
					},
				},
			}}}

}

//NewServiceHost creates a new service host from a consul catalog service
func NewServiceHost(s *api.CatalogService) ServiceHost {
	return ServiceHost{
		IPAddress:   s.ServiceAddress,
		Port:        s.ServicePort,
		Tags:        s.ServiceTags,
		Service:     s.ServiceName,
		CreateIndex: s.CreateIndex,
		ModifyIndex: s.ModifyIndex,
	}
}
