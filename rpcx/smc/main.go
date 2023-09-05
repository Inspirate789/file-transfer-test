package main

import (
	"file-transfer-test/rpcx/incident_service"
	"flag"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
	"go.uber.org/multierr"
	"log/slog"
	"os"
)

const (
	reqLimit  = 1000
	chunkSize = 1024
)

var (
	addrSMC          = flag.String("addr_smc", "localhost:8972", "SMC address")
	addrFileTransfer = flag.String("addr_file_transfer", "localhost:8973", "file transfer address")
)

func main() {
	flag.Parse()
	log.SetLogger(log.NewDefaultLogger(os.Stdout, "SMC", 0, log.LvMax))
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)).WithGroup("SMC"))

	s := server.NewServer()
	incidentService, saveFileHandler := incident_service.NewService(reqLimit, chunkSize)
	p := server.NewFileTransfer(*addrFileTransfer, saveFileHandler, nil, 1000)
	s.EnableFileTransfer(share.SendFileServiceName, p)
	err := s.RegisterName("IncidentService", incidentService, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = multierr.Combine(err, incident_service.DeleteService(incidentService))
		if err != nil {
			panic(err)
		}
	}()

	slog.Info("smc started", slog.String("addr", *addrSMC))
	err = s.Serve("tcp", *addrSMC)
	if err != nil {
		panic(err)
	}
}
