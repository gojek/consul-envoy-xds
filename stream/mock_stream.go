package stream

import (
	cp "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type MockXDSStream struct {
	mock.Mock
	Ctx context.Context
}

func (s *MockXDSStream) Capture() *cp.DiscoveryResponse {
	return s.TestData()["capture"].(*cp.DiscoveryResponse)
}

func (s *MockXDSStream) Send(r *cp.DiscoveryResponse) error {
	args := s.Called(r)
	s.TestData()["capture"] = r

	return args.Error(0)
}

func (*MockXDSStream) Recv() (*cp.DiscoveryRequest, error) { return nil, nil }
func (*MockXDSStream) SetHeader(metadata.MD) error         { return nil }
func (*MockXDSStream) SendHeader(metadata.MD) error        { return nil }
func (*MockXDSStream) SetTrailer(metadata.MD)              {}
func (m *MockXDSStream) Context() context.Context          { return m.Ctx }
func (*MockXDSStream) SendMsg(m interface{}) error         { return nil }
func (*MockXDSStream) RecvMsg(m interface{}) error         { return nil }
