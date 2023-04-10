package cmd

import (
	"os"
	"os/signal"
)

func InstallSignalHandler(fn func(os.Signal), signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	go func() {
		signal := <-c
		fn(signal)
	}()
}
