package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"

	"github.com/bcap/caller/cmd"
	"github.com/bcap/caller/handler"
)

type Args struct {
	ListenAddress     string        `arg:"-l,--listen,env:LISTEN_ADDRESS" default:":8080" help:"Which address to listen to"`
	ReadHeaderTimeout time.Duration `arg:"--read-header-timeout,env:READ_HEADER_TIMEOUT" default:"1s" help:"Once a client opens a connection with the server, how long to wait for request headers to be fully received"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd.ConfigureLogging()

	args := parseArgs()

	server := http.Server{
		Addr:              args.ListenAddress,
		Handler:           &handler.Handler{},
		ReadHeaderTimeout: args.ReadHeaderTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	listener, err := net.Listen("tcp", args.ListenAddress)
	cmd.PanicOnErr(err)

	log.Printf("Caller server running with pid %v and listening on %v", os.Getpid(), args.ListenAddress)

	onInterrupt := func(signal os.Signal) {
		log.Println("Caller server interrupted, shutting down")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer shutdownCancel()
		server.Shutdown(shutdownCtx)
	}

	cmd.InstallSignalHandler(onInterrupt, os.Interrupt)

	err = server.Serve(listener)
	if !(errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed)) {
		cmd.PanicOnErr(err)
	}
	log.Println("Caller server succesfully shutdown")
}

func parseArgs() Args {
	var args Args
	arg.MustParse(&args)
	return args
}
