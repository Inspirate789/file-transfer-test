package main

import (
	"file-transfer-test/rpcx/link_store"
	"flag"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	"go.uber.org/multierr"
	"log/slog"
	"os"
)

const (
	addrSMC = "localhost:8972"
)

func main() {
	flag.Parse()
	log.SetLogger(log.NewDefaultLogger(os.Stdout, "SMC", 0, log.LvMax))
	slog.SetDefault(slog.Default().WithGroup("SMC"))

	s := server.NewServer()
	linkService := link_store.NewLinkService(addrSMC)
	err := s.RegisterName("LinkService", linkService, "")
	if err != nil {
		panic(err)
	}
	defer func() {
		err = multierr.Combine(err, link_store.DeleteLinkService(linkService))
		if err != nil {
			panic(err)
		}
	}()

	slog.Info("smc started", slog.String("addr", addrSMC))
	err = s.Serve("tcp", addrSMC)
	if err != nil {
		panic(err)
	}
}
