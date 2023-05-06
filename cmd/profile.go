package cmd

import (
	"fmt"

	"github.com/pkg/profile"
)

type Stoppper interface {
	Stop()
}

func ProfileStart(mode string) Stoppper {
	return profile.Start(parseProfileMode(mode), profile.ProfilePath("."))
}

func parseProfileMode(mode string) func(*profile.Profile) {
	switch mode {
	case "cpu":
		return profile.CPUProfile
	case "memHeap":
		return func(p *profile.Profile) {
			profile.MemProfileRate(1)
			profile.MemProfileHeap(p)
		}
	case "memAllocs":
		return func(p *profile.Profile) {
			profile.MemProfileRate(1)
			profile.MemProfileAllocs(p)
		}
	case "goroutines":
		return profile.GoroutineProfile
	case "mutex":
		return profile.MutexProfile
	case "block":
		return profile.BlockProfile
	case "threadCreation":
		return profile.ThreadcreationProfile
	case "trace":
		return profile.TraceProfile
	case "clock":
		return profile.ClockProfile
	default:
		panic(fmt.Sprintf("unknown profile mode %q", mode))
	}
}
