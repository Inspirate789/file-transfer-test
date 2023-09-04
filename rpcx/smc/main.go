package main

import (
	"context"
	"file-transfer-test/rpcx/file_transfer"
	"file-transfer-test/rpcx/link_store"
	"flag"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/smallnest/rpcx/server"
)

const (
	chunkSize        = 1024
	addrSMC          = "localhost:8972"
	fileTransferAddr = "localhost:8973"
)

func saveFileHandler(conn net.Conn, args *share.FileTransferArgs) {
	slog.Info("saveFileHandler called")
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
	file, err := os.OpenFile(
		args.FileName+"_"+conn.LocalAddr().String(),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	_, err = io.CopyBuffer(file, conn, make([]byte, chunkSize))
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("receive ok", slog.String("receive_end_time", time.Now().Format(time.StampMilli)))
}

func onLinkCallback(xClient client.XClient) error {
	slog.Info("onLinkCallback called")
	req := file_transfer.Request{IncidentID: 15}
	resp := file_transfer.Response{}
	err := xClient.Call(context.Background(), "GetIncidentFile", req, &resp)
	if err != nil {
		return err
	}

	slog.Info("file from store was sent",
		slog.String("filename", resp.Filename),
		slog.Int("size", int(resp.Filesize)),
		slog.String("send_end_time", time.Now().Format(time.StampMilli)),
	)

	return nil
}

func main() {
	flag.Parse()

	s := server.NewServer()
	err := s.RegisterName("LinkService", link_store.NewLinkService(addrSMC, onLinkCallback), "")
	if err != nil {
		panic(err)
	}
	p := server.NewFileTransfer(fileTransferAddr, saveFileHandler, nil, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)

	slog.Info("smc started", slog.String("addr", addrSMC))
	err = s.Serve("tcp", addrSMC)
	if err != nil {
		panic(err)
	}
}
