package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/alexflint/go-arg"

	"github.com/bcap/kaller/cmd"
	"github.com/bcap/kaller/handler"
	srv "github.com/bcap/kaller/server"
)

type Args struct {
	ListenAddress string `arg:"-l,--listen,env:LISTEN_ADDRESS" default:":8080" help:"Which address to listen to"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd.ConfigureLogging()

	args := parseArgs()

	server := srv.Server{}
	addr, err := server.Listen(ctx, args.ListenAddress)
	cmd.PanicOnErr(err)

	log.Printf("Caller server running with pid %v and listening on %v", os.Getpid(), addr.AddrPort())

	cmd.InstallSignalHandler(
		func(signal os.Signal) {
			log.Println("Caller server interrupted, shutting down")
			cancel()
			server.ShutdownWithTimeout(1 * time.Second)
		},
		os.Interrupt,
	)

	err = server.Serve(handler.New(ctx))
	if !srv.IsClosedError(err) {
		cmd.PanicOnErr(err)
	}
	log.Println("Caller server succesfully shutdown")
}

func parseArgs() Args {
	var args Args
	arg.MustParse(&args)
	return args
}
