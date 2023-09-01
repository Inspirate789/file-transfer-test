package main

import (
	"context"
	"file-transfer-test/rpcx/types"
	"flag"
	"fmt"
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
	clientConn       net.Conn
	connected        = false
	sendErrors       = make(chan error, 10)
)

type Service struct {
}

func (t *Service) TransferFile(ctx context.Context, _ *types.Args, _ *types.Reply) error {
	clientConn = ctx.Value(server.RemoteConnContextKey).(net.Conn)
	// fill reply fields
	connected = true
	return nil
}

func main() {
	flag.Parse()

	s := server.NewServer()

	p := server.NewFileTransfer(*fileTransferAddr, saveFileHandler, nil, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)

	err := s.Register(new(Service), "")
	if err != nil {
		panic(err)
	}
	go func() {
		err := s.Serve("tcp", *addr)
		if err != nil {
			panic(err)
		}
	}()

	for !connected {
		time.Sleep(time.Second)
	}

	slog.Info("start to send messages", slog.String("addr", clientConn.RemoteAddr().String()))
	for {
		select {
		case err := <-sendErrors:
			slog.Info("receive ended")
			if err != nil {
				slog.Error(err.Error())
				err = s.SendMessage(clientConn, "test_service_path", "test_service_method", nil, []byte(err.Error()))
				if err != nil {
					slog.Error(fmt.Sprintf("failed to send messsage to %s: %s\n", clientConn.RemoteAddr().String(), err))
					panic(err)
				}
			}
			err = s.SendMessage(clientConn, "test_service_path", "test_service_method", nil, []byte("success"))
			if err != nil {
				slog.Error(fmt.Sprintf("failed to send messsage to %s: %s\n", clientConn.RemoteAddr().String(), err))
				panic(err)
			}

			sendErrors = make(chan error, 10)
		}
	}
}

func saveFileHandler(conn net.Conn, args *share.FileTransferArgs) {
	defer func() {
		err := conn.Close()
		if err != nil {
			slog.Error(err.Error())
			sendErrors <- err
		} else {
			sendErrors <- nil
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
		sendErrors <- err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(err.Error())
			sendErrors <- err
		}
	}()

	_, err = io.CopyBuffer(file, conn, make([]byte, chunkSize))
	if err != nil {
		slog.Error(err.Error())
		sendErrors <- err
	}
	slog.Info("receive ok", slog.String("receive_end_time", time.Now().Format(time.StampMilli)))
}
