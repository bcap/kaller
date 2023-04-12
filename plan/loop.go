package plan

type Loop struct {
	Times     int       `json:"times" yaml:"times"`
	Delay     Delay     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Execution Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
}

func (Loop) Type() StepType {
	return LoopType
}
