package incident_service

import (
	"context"
	"file-transfer-test/rpcx/file_service"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

type LinkRequest struct {
	ClientAddr string
}

type IncidentRequest struct {
	ClientAddr string
	IncidentID string
}

type Response struct {
}

type Service struct {
	incidentsQueue chan IncidentRequest
	fileChunkSize  uint
	mx             sync.RWMutex
	wg             sync.WaitGroup
	clients        map[string]client.XClient
}

func NewService(reqLimit uint, chunkSize uint) (*Service, server.FileTransferHandler) {
	s := &Service{
		incidentsQueue: make(chan IncidentRequest, reqLimit),
		fileChunkSize:  chunkSize,
		clients:        make(map[string]client.XClient),
	}

	go func() {
		s.wg.Add(1)
		s.processIncidents()
	}()

	return s, s.saveFileHandler
}

func DeleteService(s *Service) error {
	close(s.incidentsQueue)
	slog.Info("waiting for processing all incidents")
	s.wg.Wait()

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

func (s *Service) Link(_ context.Context, arg LinkRequest, _ *Response) error {
	d, err := client.NewPeer2PeerDiscovery("tcp@"+arg.ClientAddr, "")
	if err != nil {
		return err
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	cl := client.NewXClient("FileService", client.Failtry, client.RandomSelect, d, opt)
	s.mx.Lock()
	s.clients[arg.ClientAddr] = cl
	s.mx.Unlock()

	slog.Info("store linked to smc", slog.String("store_addr", arg.ClientAddr))

	return nil
}

func (s *Service) SendIncident(_ context.Context, arg IncidentRequest, _ *Response) error {
	slog.Debug("incident in pending", slog.Any("incident", arg))
	s.incidentsQueue <- arg
	return nil
}

func (s *Service) processIncidents() {
	for incident := range s.incidentsQueue {
		slog.Debug("incident in progress", slog.Any("incident", incident))
		s.mx.RLock()
		cl, inMap := s.clients[incident.ClientAddr]
		s.mx.RUnlock()
		if !inMap {
			slog.Error(fmt.Sprintf("failed to call %s: there are no xClient for this address", incident.ClientAddr))
			continue
		}
		args := file_service.FileRequest{
			FileServiceAddr: incident.ClientAddr,
			IncidentID:      incident.IncidentID,
		}
		err := cl.Call(context.Background(), "RequestFile", args, &file_service.Response{}) // TODO: goroutine for waiting to recall?
		if err != nil {
			slog.Error(fmt.Sprintf("failed to call %s: %v", incident.ClientAddr, err)) // TODO: resend
		}
	}
}

func (s *Service) saveFileHandler(conn net.Conn, args *share.FileTransferArgs) {
	defer func() {
		err := conn.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	slog.Info("receive file",
		slog.String("filename", args.FileName),
		slog.Int("size", int(args.FileSize)),
		slog.Any("meta", args.Meta),
		slog.String("receive_start_time", time.Now().Format(time.StampMilli)),
	)

	filename := "out/" + conn.RemoteAddr().String() + ".txt"
	file, err := os.Create(filename)
	if err != nil {
		slog.Error(err.Error())
		return // TODO: resend
	}
	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	_, err = io.CopyBuffer(file, conn, make([]byte, s.fileChunkSize))
	if err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("receive ok", slog.String("receive_end_time", time.Now().Format(time.StampMilli)))
}
