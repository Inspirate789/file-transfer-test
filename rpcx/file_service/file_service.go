package file_service

import (
	"context"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
	"log/slog"
	"sync"
	"time"
)

const (
	incidentFilename = "out/file.txt"
)

type FileRequest struct {
	FileServiceAddr string
	IncidentID      string
}

type Response struct {
}

type Service struct {
	addr          string
	requestsQueue chan FileRequest
	xClient       client.XClient
	wg            sync.WaitGroup
}

func NewService(serviceAddr, receiverAddr string, reqLimit uint) (*Service, error) {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+receiverAddr, "")
	if err != nil {
		return nil, err
	}
	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack
	xClient := client.NewXClient(share.SendFileServiceName, client.Failtry, client.RandomSelect, d, opt)

	s := &Service{
		addr:          serviceAddr,
		requestsQueue: make(chan FileRequest, reqLimit),
		xClient:       xClient,
	}

	go func() {
		s.wg.Add(1)
		s.processRequests()
	}()

	return s, nil
}

func DeleteService(s *Service) error {
	close(s.requestsQueue)
	slog.Info("waiting for processing all file requests")
	s.wg.Wait()

	return s.xClient.Close()
}

func (s *Service) RequestFile(_ context.Context, arg FileRequest, _ *Response) error {
	if arg.FileServiceAddr != s.addr {
		slog.Warn("file service cannot process request for another server; ignore",
			slog.String("request_addr", arg.FileServiceAddr),
			slog.String("service_addr", s.addr),
		)
		return fmt.Errorf("file service (%s) cannot process request for another server (%s); ignore",
			s.addr, arg.FileServiceAddr,
		)
	}
	s.requestsQueue <- arg
	slog.Debug("file request in pending", slog.Any("request", arg))

	return nil
}

func (s *Service) getFilenameByID(_ string) string { // TODO: external call?
	return incidentFilename
}

func (s *Service) processRequests() {
	for request := range s.requestsQueue {
		slog.Debug("file request in progress", slog.Any("request", request))
		filename := s.getFilenameByID(request.IncidentID)
		slog.Info("send file to smc",
			slog.String("send_start_time", time.Now().Format(time.StampMilli)),
			slog.String("filename", filename),
		)
		err := s.xClient.SendFile(context.Background(), filename, 0, nil) // TODO: context timeout?
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Info("send ok", slog.String("send_end_time", time.Now().Format(time.StampMilli)))
	}
}
