package main

import (
	"context"
	"file-transfer-test/generator"
	"flag"
	"github.com/smallnest/rpcx/protocol"
	"log/slog"
	"time"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"
)

const (
	filename = "file.txt"
	filesize = 1024 * 1024 * 1024
)

var (
	addr = flag.String("addr", "localhost:8972", "server address")
)

func main() {
	flag.Parse()

	d, err := client.NewPeer2PeerDiscovery("tcp@"+*addr, "")
	if err != nil {
		panic(err)
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	xClient := client.NewXClient(share.SendFileServiceName, client.Failtry, client.RandomSelect, d, opt)
	defer func(xClient client.XClient) {
		err = xClient.Close()
		if err != nil {
			panic(err)
		}
	}(xClient)

	err = generator.GenerateFile(filename, filesize)
	if err != nil {
		panic(err)
	}

	for {
		err = xClient.SendFile(context.Background(), filename, 0, map[string]string{
			"my_name":         "client1",
			"send_start_time": time.Now().Format(time.StampMilli),
		})
		if err == nil {
			break
		}
		slog.Error(err.Error())
		slog.Warn("send file again...")
	}
	slog.Info("send ok", slog.String("send_end_time", time.Now().Format(time.StampMilli)))
}
