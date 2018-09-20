package agent

import (
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/testutil"
)

func StartConsulTestServer() (*testutil.TestServer, *consulapi.Client) {
	consulSvr, _ := testutil.NewTestServer()
	consulClientCfg := consulapi.DefaultConfig()
	consulClientCfg.Address = consulSvr.HTTPAddr
	consulClient, _ := consulapi.NewClient(consulClientCfg)
	return consulSvr, consulClient
}
