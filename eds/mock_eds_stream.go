package eds

import (
	cp "github.com/envoyproxy/go-control-plane/api"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type MockEDSStream struct {
	mock.Mock
	Ctx context.Context
}

func (s *MockEDSStream) Capture() *cp.DiscoveryResponse {
	return s.TestData()["capture"].(*cp.DiscoveryResponse)
}

func (s *MockEDSStream) Send(r *cp.DiscoveryResponse) error {
	args := s.Called(r)
	s.TestData()["capture"] = r

	return args.Error(0)
}

func (*MockEDSStream) Recv() (*cp.DiscoveryRequest, error) { return nil, nil }
func (*MockEDSStream) SetHeader(metadata.MD) error         { return nil }
func (*MockEDSStream) SendHeader(metadata.MD) error        { return nil }
func (*MockEDSStream) SetTrailer(metadata.MD)              {}
func (m *MockEDSStream) Context() context.Context          { return m.Ctx }
func (*MockEDSStream) SendMsg(m interface{}) error         { return nil }
func (*MockEDSStream) RecvMsg(m interface{}) error         { return nil }
