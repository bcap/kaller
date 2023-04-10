package cmd

import (
	"log"
	"os"
)

func ConfigureLogging() {
	log.SetPrefix("")
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}
