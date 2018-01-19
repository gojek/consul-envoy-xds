package eds

import (
	cp "github.com/envoyproxy/go-control-plane/api"
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

// LbEndpoint translates a consul agent service endpoint to an envoy control plane LbEndpoint
func (h ServiceHost) LbEndpoint() *cp.LbEndpoint {
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
