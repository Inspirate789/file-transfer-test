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
	reqLimit         = 1000
	chunkSize        = 1024
	retriesOnFailure = 5
)

var (
	portSMC          = flag.String("port_smc", ":8972", "SMC port")
	portFileTransfer = flag.String("port_file_transfer", ":8973", "file transfer port")
)

func main() {
	flag.Parse()
	log.SetLogger(log.NewDefaultLogger(os.Stdout, "SMC", 0, log.LvMax))
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)).WithGroup("SMC"))

	s := server.NewServer()
	incidentService, saveFileHandler := incident_service.NewService(reqLimit, chunkSize, retriesOnFailure)
	p := server.NewFileTransfer(*portFileTransfer, saveFileHandler, nil, reqLimit)
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

	slog.Info("smc started", slog.String("addr", *portSMC))
	err = s.Serve("tcp", *portSMC)
	if err != nil {
		panic(err)
	}
}
