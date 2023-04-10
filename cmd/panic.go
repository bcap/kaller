package cmd

import "log"

func PanicOnErr(err error) {
	if err != nil {
		log.Printf("fatal error: %v", err)
		panic(err)
	}
}
