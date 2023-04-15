package plan

import (
	"fmt"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

// Delay represents that the service should spend some time before moving to the next step
// The delay can operate in 2 different ways:
//   - A fixed amount of time if only Min is defined, or if both Min and Max have the same value
//   - A random time between Min and Max
//
// TODO: implement delay strategies like: sleep vs cpu burn
type Delay struct {
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

const DelayPattern = `` +
	// Min
	`(\w+)` +
	// Optional Max
	`(?:\s+to\s+(\w+))?`

var delayPattern = regexp.MustCompile(DelayPattern)

// Constructs a Delay from a simple string. Examples:
//   - "10ms" creates a Delay with both Min and Max at 10ms
//   - "10ms to 200ms" creates a Delay with Min set to 10ms and Max set to 200ms
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
	d.Min = min
	d.Max = max
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
