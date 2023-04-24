package plan

type CallType string

const (
	CallTypeSync  CallType = "sync"
	CallTypeAsync CallType = "async"
)

// Call represents that a service call should be invoked and how that service should
// process it
//
// Calls have the following phases:
//  1. Receive and read the passed request in full
//  2. Compute, if a compute period was specified.
//  3. Execute the Execution steps, if provided (eg: do calls to other services)
//  4. Respond the call with the given response delais.
//  5. Execute the PostExecution steps, if provided (eg: do work after response was sent)
//  6. Wait for any pending async Call that was launched in either Exection steps or PostExecution steps
//
// The Compute in step 2 is a way to define calls more simply. The same can be achieved with
// an Execution where the first step is a Compute.
//
// Calls can be sync (the default) or async by setting the Async parameter. In sync calls the client
// will wait for the call result before moving to the next step. In async calls the client will
// not wait for the call result and move to the next step.
//
// NOTE: As of now only HTTP calls are supported
type Call struct {
	Async         bool      `json:"async" yaml:"async"`
	HTTP          HTTP      `json:"http,omitempty" yaml:"http,omitempty"`
	Compute       Compute   `json:"compute,omitempty" yaml:"compute,omitempty"`
	Execution     Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
	PostExecution Execution `json:"post-execution,omitempty" yaml:"post-execution,omitempty"`
}

func (Call) StepType() StepType {
	return StepTypeCall
}
