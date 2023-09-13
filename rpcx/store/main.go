package main

import (
	"context"
	"file-transfer-test/rpcx/file_service"
	"file-transfer-test/rpcx/incident_service"
	"flag"
	"fmt"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"go.uber.org/multierr"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

const (
	hostStore = "localhost"
	addrSMC   = "localhost:8972"
	reqLimit  = 1000
)

var (
	portStore = flag.String("port", ":8800", "store port")
	sleepTime = flag.Uint("sleep", 0, "store sleep time between linking and sending incident (in seconds)")
)

func main() {
	flag.Parse()
	log.SetLogger(log.NewDefaultLogger(os.Stdout, fmt.Sprintf("Store (%s)", *portStore), 0, log.LvMax))
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)).WithGroup(fmt.Sprintf("Store (%s)", *portStore)))

	s := server.NewServer()
	fileService, err := file_service.NewService(hostStore+(*portStore), addrSMC, reqLimit)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = multierr.Combine(err, file_service.DeleteService(fileService))
		if err != nil {
			panic(err)
		}
	}()
	err = s.RegisterName("FileService", fileService, "")
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		err = multierr.Combine(err, s.Serve("tcp", *portStore))
		if err != nil {
			panic(err)
		}
	}()
	slog.Info("store started", slog.String("addr", *portStore))

	d, err := client.NewPeer2PeerDiscovery("tcp@"+addrSMC, "")
	if err != nil {
		panic(err)
	}

	opt := client.DefaultOption
	opt.SerializeType = protocol.MsgPack

	xClient := client.NewXClient("IncidentService", client.Failtry, client.RandomSelect, d, opt)
	defer func(xClient client.XClient) {
		err = multierr.Combine(err, xClient.Close())
		if err != nil {
			panic(err)
		}
	}(xClient)

	slog.Info("call to smc...")
	host, _, _ := net.SplitHostPort(addrSMC)
	rtt, err := client.Ping(host)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info("Ping to smc", slog.Int("rtt (ms)", rtt))

	slog.Info("link to smc...")
	linkReq := incident_service.LinkRequest{
		ClientAddr: hostStore + *portStore,
	}
	err = xClient.Call(context.Background(), "Link", linkReq, &file_service.Response{})
	if err != nil {
		panic(err)
	}
	slog.Info("linked to smc")

	slog.Info("waiting for incident...")
	time.Sleep(time.Duration(*sleepTime) * time.Second)
	incidentReq := incident_service.IncidentRequest{
		ClientAddr: hostStore + *portStore,
		IncidentID: "12345",
	}
	err = xClient.Call(context.Background(), "SendIncident", incidentReq, &file_service.Response{})
	if err != nil {
		panic(err)
	}

	wg.Wait()
}
