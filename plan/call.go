package plan

type Call struct {
	HTTP      HTTP      `json:"http,omitempty" yaml:"http,omitempty"`
	Delay     Delay     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Execution Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
}

func (Call) Type() StepType {
	return CallType
}
