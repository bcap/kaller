package plan

// Call represents that a service call should be invoked and how that service should
// process it
//
// Calls have the following phases:
//  1. Receive and read the passed request in full
//  2. Delay, if a delay period was specified.
//  3. Execute the Execution steps, if provided (eg: do calls to other services)
//  4. Respond the call with the given response delais
//  5. Execute the PostExecution steps, if provided (eg: do work after response was sent)
//
// NOTE: The Delay in step 2 is a way to define calls more simply. The same can be achieved with
// an Execution where the first step is a Delay.
//
// NOTE: As of now only HTTP calls are supported
type Call struct {
	HTTP          HTTP      `json:"http,omitempty" yaml:"http,omitempty"`
	Delay         Delay     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Execution     Execution `json:"execution,omitempty" yaml:"execution,omitempty"`
	PostExecution Execution `json:"post-execution,omitempty" yaml:"post-execution,omitempty"`
}

func (Call) Type() StepType {
	return CallType
}
