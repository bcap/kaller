package plan

type Call struct {
	HTTP          HTTP      `json:"http,omitempty" yaml:"http,omitempty"`
	Delay         Delay     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Execution     Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
	PostExecution Execution `json:"post-execution,omitempty" yaml:"post-execution,omitempty"`
}

func (Call) Type() StepType {
	return CallType
}
