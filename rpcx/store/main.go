package main

import (
	"context"
	"file-transfer-test/rpcx/link_store"
	"flag"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
	"go.uber.org/multierr"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

const (
	addrSMC          = "localhost:8972"
	fileTransferAddr = "localhost:8973"
)

var (
	addrStore  = flag.String("addr", "localhost:8800", "server address")
	fileBuffer = make([]byte, 1024)
)

func downloadFileHandler(conn net.Conn, args *share.DownloadFileArgs) {
	slog.Info("send file to smc",
		slog.Any("args", *args),
		slog.String("send_start_time", time.Now().Format(time.StampMilli)),
	)
	defer func() {
		err := conn.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	file, err := os.Open(args.FileName)
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

	_, err = io.CopyBuffer(conn, file, fileBuffer)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("send ok", slog.String("send_end_time", time.Now().Format(time.StampMilli)))
}

func main() {
	flag.Parse()

	d, err := client.NewPeer2PeerDiscovery("tcp@"+addrSMC, "")
	if err != nil {
		panic(err)
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	xClient := client.NewXClient("LinkService", client.Failtry, client.RandomSelect, d, opt)
	defer func(xClient client.XClient) {
		err = multierr.Combine(err, xClient.Close())
		if err != nil {
			panic(err)
		}
	}(xClient)

	s := server.NewServer()
	p := server.NewFileTransfer(fileTransferAddr, nil, downloadFileHandler, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)

	slog.Info("store started", slog.String("addr", *addrStore))
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		err = multierr.Combine(err, s.Serve("tcp", *addrStore))
		if err != nil {
			panic(err)
		}
	}()

	slog.Info("call to smc...")
	req := link_store.Request{ClientAddr: *addrStore}
	err = xClient.Call(context.Background(), "Link", req, &link_store.Response{})
	if err != nil {
		panic(err)
	}
	slog.Info("linked to smc")
	wg.Wait()
}
