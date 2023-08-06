package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"

	"github.com/bcap/kaller/cmd"
	"github.com/bcap/kaller/handler"
	"github.com/bcap/kaller/plan"
	srv "github.com/bcap/kaller/server"
)

type Args struct {
	Plan    string `arg:"positional,required" help:"The plan yaml file to use. Use \"-\" to read the plan from stdin"`
	Port    int    `arg:"-p,--port" help:"control the tcp port for the localhost server that is used to execute the plan" default:"0"`
	Profile string `arg:"--profile" help:"Enables profiling for the given mode. Available modes at cmd/profile.go"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd.ConfigureLogging()

	args := parseArgs()

	if args.Profile != "" {
		defer cmd.ProfileStart(args.Profile).Stop()
	}

	server := srv.Server{}
	addr, err := server.Listen(ctx, fmt.Sprintf(":%d", args.Port))
	cmd.PanicOnErr(err)

	go func() {
		err := server.Serve(handler.New(ctx))
		if !srv.IsClosedError(err) {
			cmd.PanicOnErr(err)
		}
	}()

	plan := readPlan(args.Plan)

	localRunURL := fmt.Sprintf("http://%s/run-plan", addr.AddrPort())
	req, err := http.NewRequestWithContext(ctx, "POST", localRunURL, nil)
	cmd.PanicOnErr(err)
	handler.WritePlanHeaders(req, plan, "")

	http.DefaultClient.Do(req)

	handler := server.Handler().(*handler.Handler)
	for handler.Outstanding() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

func parseArgs() Args {
	var args Args
	arg.MustParse(&args)
	return args
}

func readPlan(location string) plan.Plan {
	var input io.Reader = os.Stdin
	if location != "-" {
		var err error
		input, err = os.OpenFile(location, os.O_RDONLY, 0)
		cmd.PanicOnErr(err)
	}
	data, err := io.ReadAll(input)
	cmd.PanicOnErr(err)
	plan, err := plan.FromYAML(data)
	cmd.PanicOnErr(err)
	return plan
}
