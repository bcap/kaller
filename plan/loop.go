package plan

// Loop is used to repeat a whole list of steps, defined by the Execution parameter.
// Compute can used to define wait time in between repeating such executions
type Loop struct {
	Times       int       `json:"times" yaml:"times"`
	Concurrency int       `json:"concurrency,omitempty" yaml:"concurrency,omitempty"`
	Compute     Compute   `json:"compute,omitempty" yaml:"compute,omitempty"`
	Execution   Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
}

func (Loop) StepType() StepType {
	return StepTypeLoop
}
