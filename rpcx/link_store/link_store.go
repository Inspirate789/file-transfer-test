package link_store

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"log/slog"
)

type Request struct {
	ClientAddr string
}

type Response struct {
}

type LinkService struct {
	serviceAddr    string
	onLinkCallback func(client.XClient) error
	clients        map[string]client.XClient
}

func NewLinkService(serviceAddr string, onLinkCallback func(client.XClient) error) *LinkService {
	return &LinkService{
		serviceAddr:    serviceAddr,
		onLinkCallback: onLinkCallback,
		clients:        make(map[string]client.XClient),
	}
}

func (s *LinkService) Link(_ context.Context, arg Request, _ *Response) error {
	slog.Info("LinkService.Link called")
	d, err := client.NewPeer2PeerDiscovery("tcp@"+arg.ClientAddr, "")
	if err != nil {
		return err
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	s.clients[arg.ClientAddr] = client.NewXClient("FileService", client.Failtry, client.RandomSelect, d, opt)

	return s.onLinkCallback(s.clients[arg.ClientAddr])
}
