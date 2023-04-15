package plan

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Delay represents that the service should spend some time before moving to the next step
// The delay can operate in 2 different ways:
//   - A fixed amount of time if only Min is defined, or if both Min and Max have the same value
//   - A random time between Min and Max
//
// Also a CPU float parameter can be passed to Delay in order to generate CPU load:
//   - A CPU value of 0.0 means no cpu usage. Delay will only sleep
//   - A CPU value of 1.0 means that we should load a single core completely
//   - A CPU value of 2.3 means that we should load 2 cores completely and 30% of a 3rd core
type Delay struct {
	CPU float64
	Min time.Duration `json:"min" yaml:"min"`
	Max time.Duration `json:"max" yaml:"max"`
}

func (Delay) StepType() StepType {
	return StepTypeDelay
}

func (d *Delay) String() string {
	if d.Min == d.Max {
		return d.Min.String()
	}
	return fmt.Sprintf("%s to %s", d.Min, d.Max)
}

func (d Delay) IsZero() bool {
	return d.Min == 0 && d.Max == 0
}

func (d Delay) Validate() error {
	if d.IsZero() {
		return nil
	}
	if d.Min < 0 || d.Max < 0 {
		return fmt.Errorf("invalid delay: min and/or max are negative (min: %v, max: %v)", d.Min, d.Max)
	}
	if d.Min > d.Max && d.Max > 0 {
		return fmt.Errorf("invalid delay: min is higher than max (min: %v, max: %v)", d.Min, d.Max)
	}
	return nil
}

var delayRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (d Delay) Do(ctx context.Context) {
	if err := d.Validate(); err != nil {
		return
	}
	if d.IsZero() {
		return
	}
	duration := d.Min
	if d.Min < d.Max {
		delta := int64(d.Max - d.Min)
		duration = d.Min + time.Duration(delayRand.Int63n(delta))
	}

	d.do(ctx, duration)
}

func (d Delay) do(ctx context.Context, duration time.Duration) {
	if d.CPU > 0.0 {
		start := time.Now()
		// Delay.CPU is capped at the number of cores.
		// For instance if Delay.CPU is 8.5 on a 4 core system, effectivelly
		// a load of CPU 4.0 will be generated
		for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
			ratio := d.CPU - float64(cpu)
			if ratio <= 0.0 {
				break
			}
			if ratio > 1.0 {
				ratio = 1.0
			}
			workUnit := 100 * time.Millisecond
			if duration < workUnit {
				workUnit = duration
			}
			toRun := time.Duration(float64(workUnit) * ratio)
			toSleep := workUnit - toRun
			go func() {
				runtime.LockOSThread()
				for {
					unitStart := time.Now()
					// this tight loop should take 100% of a core
					for {
						if time.Since(start) >= duration {
							return
						} else if time.Since(unitStart) >= toRun {
							break
						}
					}
					time.Sleep(toSleep)
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
			}()
		}
	}
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
}

const DelayPattern = `` +
	// Min
	`(\w+)` +
	// Optional Max
	`(?:\s+to\s+(\w+))?` +
	// Optional CPU
	`(?:\s+([\d\.]+)\s*cpu)?`

var delayPattern = regexp.MustCompile(DelayPattern)

// Constructs a Delay from a simple string. Examples:
//   - "10ms" creates a Delay with a fixed time of 10ms (both Min and Max at 10ms)
//   - "10ms to 200ms" creates a Delay with Min set to 10ms and Max set to 200ms
//   - "10ms 2.3 cpu" creates a Delay a fixed time of 10ms where we will try to run 100% of 2 cpu cores and 30% of another cpu core
//   - "10ms to 100ms 1.5 cpu" creates a Delay with Min set to 10ms and Max set to 100ms, where we will try to overload a single core and 50% of another one
func (d *Delay) Parse(s string) error {
	parts := delayPattern.FindStringSubmatch(s)
	if parts == nil {
		return fmt.Errorf("cannot parse delay definition %q", s)
	}
	min, err := time.ParseDuration(parts[1])
	if err != nil {
		return fmt.Errorf("invalid delay time %q: %w", parts[1], err)
	}
	max := min
	if parts[2] != "" {
		max, err = time.ParseDuration(parts[2])
		if err != nil {
			return fmt.Errorf("invalid delay time %q: %w", parts[2], err)
		}
	}
	cpu := 0.0
	if parts[3] != "" {
		cpu, err = strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return fmt.Errorf("invalid delay cpu %q: %w", parts[3], err)
		}
	}
	d.Min = min
	d.Max = max
	d.CPU = cpu
	return nil
}

func (d *Delay) UnmarshalYAML(node *yaml.Node) error {
	if node.Value != "" {
		if err := d.Parse(node.Value); err != nil {
			return fmt.Errorf("invalid delay definition at line %d: %w", node.Line, err)
		}
		return nil
	}
	type raw Delay
	return node.Decode((*raw)(d))
}
