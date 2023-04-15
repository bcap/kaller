package plan

// Loop is used to repeat a whole list of steps, defined by the Execution parameter.
// Delay can used to define wait time in between repeating such executions
type Loop struct {
	Times     int       `json:"times" yaml:"times"`
	Delay     Delay     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Execution Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
}

func (Loop) StepType() StepType {
	return StepTypeLoop
}
