package plan

import (
	"fmt"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

type Delay struct {
	Min time.Duration `json:"min" yaml:"min"`
	Max time.Duration `json:"max" yaml:"max"`
}

func (Delay) Type() StepType {
	return DelayType
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

var delayPattern = regexp.MustCompile(`(\w+)(?:\s+to\s+(\w+))?`)

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
	type rawDelay Delay
	raw := (*rawDelay)(d)
	return node.Decode(raw)
}
