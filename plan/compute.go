package plan

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/bcap/kaller/memory"
	"gopkg.in/yaml.v3"
)

// Compute represents that the service should simulate some computation before moving to the next step
//
// The time it will take to simulate the computation can be defined in 2 different ways:
//   - A fixed amount of time if only Min is defined, or if both Min and Max have the same value
//   - A random time will be chosen between Min and Max if both have different values
//
// To generate CPU load, a CPU float parameter can be passed to Compute in order:
//   - A CPU value of 0.0 means no cpu usage. Compute will only sleep
//   - A CPU value of 1.0 means that we should load a single core completely
//   - A CPU value of 2.3 means that we should load 2 cores completely and 30% of a 3rd core
//
// To simulate memory usage, a MemoryDeltaKB parameter can be passed. The parameter can be either
// positive (memory allocated) as well negative (memory fred).
//
// NOTE: While Compute simulates memory usage (amount of allocated bytes), it does *not* simulate
// memory pressure (constantly accessing memory and potentially causing paging)
type Compute struct {
	Min           time.Duration `json:"min" yaml:"min"`
	Max           time.Duration `json:"max" yaml:"max"`
	CPU           float64       `json:"cpu" yaml:"cpu"`
	MemoryDeltaKB int           `json:"memory-delta-kb" yaml:"memory-delta-kb"`
}

func (Compute) StepType() StepType {
	return StepTypeCompute
}

func (d *Compute) String() string {
	if d.Min == d.Max {
		return d.Min.String()
	}
	return fmt.Sprintf("%s to %s", d.Min, d.Max)
}

func (d Compute) IsZero() bool {
	return d.Min == 0 && d.Max == 0
}

func (d Compute) Validate() error {
	if d.IsZero() {
		return nil
	}
	if d.Min < 0 || d.Max < 0 {
		return fmt.Errorf("invalid compute: min and/or max are negative (min: %v, max: %v)", d.Min, d.Max)
	}
	if d.Min > d.Max && d.Max > 0 {
		return fmt.Errorf("invalid compute: min is higher than max (min: %v, max: %v)", d.Min, d.Max)
	}
	return nil
}

var computeRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (d Compute) Do(ctx context.Context, fill *memory.Fill) {
	if err := d.Validate(); err != nil {
		return
	}
	if d.IsZero() {
		return
	}
	duration := d.Min
	if d.Min < d.Max {
		delta := int64(d.Max - d.Min)
		duration = d.Min + time.Duration(computeRand.Int63n(delta))
	}

	d.do(ctx, duration, fill)
}

func (d Compute) do(ctx context.Context, duration time.Duration, fill *memory.Fill) {
	start := time.Now()
	d.memory(fill)
	d.compute(ctx, duration-time.Since(start))
}

func (d Compute) memory(fill *memory.Fill) {
	if d.MemoryDeltaKB != 0 {
		fill.Add(d.MemoryDeltaKB * 1024)
	}
}

func (d Compute) compute(ctx context.Context, duration time.Duration) {
	if d.CPU > 0.0 {
		start := time.Now()
		// Compute.CPU is capped at the number of cores.
		// For instance if Compute.CPU is 8.5 on a 4 core system, effectivelly
		// a load of CPU 4.0 will be generated
		for cpu := 0; cpu < runtime.NumCPU(); cpu++ {
			ratio := d.CPU - float64(cpu)
			if ratio <= 0.0 {
				break
			}
			if ratio > 1.0 {
				ratio = 1.0
			}
			workUnit := 1 * time.Millisecond
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

const ComputePattern = `` +
	// Min
	`(\w+)` +
	// Optional Max
	`(?:\s+to\s+(\w+))?` +
	// Optional CPU
	`(?:\s+([\d\.]+)\s*cpu)?` +
	// Optional Memory delta
	`(?:\s+(` +
	`((?:-|\+)[\d\.]+)` + // signal (+ or -) plus numeric value
	`((?:k|m)b)` + // unit: kb or mb
	`))?`

var computePattern = regexp.MustCompile(ComputePattern)

// Constructs a Compute from a simple string. Examples:
//   - "10ms" creates a Compute that will run for 10ms with no cpu load nor synthetic memory growth
//   - "10ms to 200ms" creates a Compute that will run for 10ms to 200ms with no cpu load nor synthetic memory growth
//   - "10ms 2.3 cpu" creates a Compute that will run for 10ms loading 2 cores completely and 30% of another core
//   - "10ms to 100ms 1.5 cpu" creates a Compute with that will run for 10ms to 100ms and will use 1.3 cores (load 1 core by 100% and another one by 30%)
//   - "10ms +10mb" creates a Compute that will run for 10ms and increase memory usage by 10mb
//   - "10ms to 50ms 1.3 cpu -100kb" creates a Compute that will run for 10ms to 50ms, use 1.3 cpus and decrease memory usage by 100kb
func (d *Compute) Parse(s string) error {
	parts := computePattern.FindStringSubmatch(s)
	if parts == nil {
		return fmt.Errorf("cannot parse compute definition %q", s)
	}
	min, err := time.ParseDuration(parts[1])
	if err != nil {
		return fmt.Errorf("invalid compute time %q: %w", parts[1], err)
	}
	max := min
	if parts[2] != "" {
		max, err = time.ParseDuration(parts[2])
		if err != nil {
			return fmt.Errorf("invalid compute time %q: %w", parts[2], err)
		}
	}
	cpu := 0.0
	if parts[3] != "" {
		cpu, err = strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return fmt.Errorf("invalid compute cpu %q: %w", parts[3], err)
		}
	}
	memDelta := 0
	if parts[4] != "" {
		memDelta, err = strconv.Atoi(parts[5])
		if err != nil {
			return fmt.Errorf("invalid memory delta %q: %w", parts[4], err)
		}
		unit := parts[6]
		switch unit {
		case "kb":
		case "mb":
			memDelta *= 1024
		default:
			return fmt.Errorf("invalid memory delta %q: %w", parts[4], err)
		}
	}
	d.Min = min
	d.Max = max
	d.CPU = cpu
	d.MemoryDeltaKB = memDelta
	return nil
}

func (d *Compute) UnmarshalYAML(node *yaml.Node) error {
	if node.Value != "" {
		if err := d.Parse(node.Value); err != nil {
			return fmt.Errorf("invalid compute definition at line %d: %w", node.Line, err)
		}
		return nil
	}
	type raw Compute
	return node.Decode((*raw)(d))
}
