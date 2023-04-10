package plan

type Parallel struct {
	Concurrency int       `json:"concurrency" yaml:"concurrency"`
	Execution   Execution `json:"execution" yaml:"execution"`
}

func (Parallel) Type() StepType {
	return ParallelType
}
