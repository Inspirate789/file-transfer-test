package main

import (
	"context"
	"file-transfer-test/rpcx/file_transfer"
	"file-transfer-test/rpcx/link_store"
	"flag"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
	"go.uber.org/multierr"
	"log/slog"
	"sync"
)

const (
	addrSMC = "localhost:8972"
)

var (
	addrStore = flag.String("addr", "localhost:8800", "server address")
)

func closeXClient(xClient client.XClient, err error) {
	err = multierr.Combine(err, xClient.Close())
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	d, err := client.NewPeer2PeerDiscovery("tcp@"+addrSMC, "")
	if err != nil {
		panic(err)
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	xLinkClient := client.NewXClient("LinkService", client.Failtry, client.RandomSelect, d, opt)
	defer closeXClient(xLinkClient, err)

	xFileClient := client.NewXClient(share.SendFileServiceName, client.Failtry, client.RandomSelect, d, opt)
	defer closeXClient(xFileClient, err)

	s := server.NewServer()
	err = s.RegisterName("FileService", file_transfer.NewFileService(xFileClient), "")
	if err != nil {
		panic(err)
	}

	slog.Info("store started", slog.String("addr", *addrStore))
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		err = multierr.Combine(s.Serve("tcp", *addrStore))
		if err != nil {
			panic(err)
		}
	}()

	slog.Info("call to smc...")
	req := link_store.Request{ClientAddr: *addrStore}
	err = xLinkClient.Call(context.Background(), "Link", req, &link_store.Response{})
	if err != nil {
		panic(err)
	}
	slog.Info("linked to smc")
	wg.Wait()
}
