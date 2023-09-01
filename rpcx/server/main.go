package main

import (
	"flag"
	"github.com/smallnest/rpcx/share"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/smallnest/rpcx/server"
)

const (
	chunkSize = 1024
)

var (
	addr             = flag.String("addr", "localhost:8972", "server address")
	fileTransferAddr = flag.String("transfer-addr", "localhost:8973", "data transfer address")
)

func main() {
	flag.Parse()

	s := server.NewServer()

	p := server.NewFileTransfer(*fileTransferAddr, saveFileHandler, nil, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)

	err := s.Serve("tcp", *addr)
	if err != nil {
		panic(err)
	}
}

func saveFileHandler(conn net.Conn, args *share.FileTransferArgs) {
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
	file, err := os.OpenFile(args.FileName+"_response", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
