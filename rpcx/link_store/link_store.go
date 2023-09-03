package link_store

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
)

type Request struct {
	ClientAddr string
}

type Response struct {
}

type LinkService struct {
	serviceAddr string
	clients     map[string]client.XClient
}

func NewLinkService(serviceAddr string) *LinkService {
	return &LinkService{
		serviceAddr: serviceAddr,
		clients:     make(map[string]client.XClient),
	}
}

func (s *LinkService) Link(_ context.Context, arg Request, _ *Response) error {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+arg.ClientAddr, "")
	if err != nil {
		return err
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	s.clients[arg.ClientAddr] = client.NewXClient(share.SendFileServiceName, client.Failtry, client.RandomSelect, d, opt)

	return nil
}
