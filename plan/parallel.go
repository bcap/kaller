package plan

// Parallel instructs that the steps defined in the Execution list should
// be executed in parallel instead of being executed sequentially.
//
// The concurrency parameter defines a limit on  how many steps can be executed concurrently
type Parallel struct {
	Concurrency int       `json:"concurrency" yaml:"concurrency"`
	Execution   Execution `json:"execution" yaml:"execution"`
}

func (Parallel) StepType() StepType {
	return StepTypeParallel
}
