package main

import (
	"context"
	"file-transfer-test/generator"
	"file-transfer-test/rpcx/types"
	"flag"
	"github.com/smallnest/rpcx/protocol"
	"log"
	"log/slog"
	"time"

	"github.com/smallnest/rpcx/client"
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

	ch := make(chan *protocol.Message)
	xClient := client.NewBidirectionalXClient("Service", client.Failtry, client.RandomSelect, d, client.DefaultOption, ch)
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
		if err != nil {
			panic(err)
		}
		slog.Info("send ok", slog.String("send_end_time", time.Now().Format(time.StampMilli)))

		args := &types.Args{}
		reply := &types.Reply{}
		err = xClient.Call(context.Background(), "ConnectServer", args, reply)
		if err != nil {
			log.Fatalf("failed to call: %v", err)
		}

		// log args and FILLED reply from server

		msg := <-ch
		slog.Info("receive msg from server", slog.String("message", string(msg.Payload)))
		if string(msg.Payload) == "success" {
			break
		} else {
			slog.Warn("send file again...")
		}
	}
	slog.Info("work completed")
}
