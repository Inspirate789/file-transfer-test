package main

import (
	"context"
	"file-transfer-test/rpcx/file_transfer"
	"file-transfer-test/rpcx/link_store"
	"flag"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"go.uber.org/multierr"
	"log/slog"
)

const (
	addrSMC = "localhost:8972"
)

var (
	addrStore = flag.String("addr", "localhost:8800", "server address")
)

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
	err = s.RegisterName("FileService", file_transfer.NewFileService(xClient), "")
	if err != nil {
		panic(err)
	}

	slog.Info("store started", slog.String("addr", *addrStore))
	go func() {
		err = multierr.Combine(s.Serve("tcp", *addrStore))
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
}
