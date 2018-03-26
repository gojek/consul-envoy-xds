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
func (h ServiceHost) LbEndpoint() *cp.LbEndpoint {
	if uint32(h.Port) > 0 {
		return &cp.LbEndpoint{
			HealthStatus: cp.HealthStatus_HEALTHY,
			Endpoint: &cp.Endpoint{
				Address: &cp.Address{
					Address: &cp.Address_SocketAddress{
						SocketAddress: &cp.SocketAddress{
							Protocol: cp.SocketAddress_TCP,
							Address:  h.IPAddress,
							PortSpecifier: &cp.SocketAddress_PortValue{
								PortValue: uint32(h.Port),
							},
						},
					},
				}}}
	}
	return &cp.LbEndpoint{
		HealthStatus: cp.HealthStatus_HEALTHY,
		Endpoint: &cp.Endpoint{
			Address: &cp.Address{
				Address: &cp.Address_Pipe{
					Pipe: &cp.Pipe{
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
