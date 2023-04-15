package plan

// Step is an item in an execution list
// Check the StepType enum values for types of steps
type Step interface {
	StepType() StepType
}

type StepType string

const (
	StepTypeCall     StepType = "call"
	StepTypeDelay    StepType = "delay"
	StepTypeParallel StepType = "parallel"
	StepTypeLoop     StepType = "loop"
)
