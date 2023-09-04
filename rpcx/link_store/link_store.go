package link_store

import (
	"context"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Request struct {
	ClientAddr string
}

type Response struct {
}

type LinkService struct {
	serviceAddr string
	mx          sync.RWMutex
	clients     map[string]*client.OneClient
}

func NewLinkService(serviceAddr string) *LinkService {
	return &LinkService{
		serviceAddr: serviceAddr,
		clients:     make(map[string]*client.OneClient),
	}
}

func DeleteLinkService(s *LinkService) error {
	s.mx.Lock()
	for _, cl := range s.clients {
		err := cl.Close()
		if err != nil {
			return err
		}
	}
	s.mx.Unlock()

	return nil
}

func (s *LinkService) receiveFile(clientAddr string) error {
	slog.Info("LinkService.receiveFile called")

	filename := clientAddr + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	slog.Info("receive file from store ...",
		slog.String("receive_start_time", time.Now().Format(time.StampMilli)),
	)
	s.mx.RLock()
	cl := s.clients[clientAddr]
	s.mx.RUnlock()
	err = cl.DownloadFile(context.Background(), "file.txt", file, map[string]string{
		"incident_id": "12345",
	})
	if err != nil {
		return err
	}
	slog.Info("file from store received",
		slog.String("receive_end_time", time.Now().Format(time.StampMilli)),
	)

	return nil
}

func (s *LinkService) Link(_ context.Context, arg Request, _ *Response) error {
	slog.Info("LinkService.Link called")
	d, err := client.NewPeer2PeerDiscovery("tcp@"+arg.ClientAddr, "")
	if err != nil {
		return err
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	s.mx.Lock()
	s.clients[arg.ClientAddr] = client.NewOneClient(client.Failtry, client.RandomSelect, d, opt)
	s.mx.Unlock()

	return s.receiveFile(arg.ClientAddr)
}
