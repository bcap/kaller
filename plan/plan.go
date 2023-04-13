package plan

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// Plan represents the whole configuration for this application. It defines the plan on
// how callers should execute steps and call each other
//
// As of now the Plan is serialized and sent to all services that participate in the
// call mesh. For more details on how this is transported check handler.WritePlanHeaders and
// handler.ReadPlanHeaders
type Plan struct {
	Execution Execution `json:"execution" yaml:"execution"`
}

func FromJSON(data []byte) (Plan, error) {
	var plan Plan
	err := json.Unmarshal(data, &plan)
	return plan, err
}

func FromYAML(data []byte) (Plan, error) {
	var plan Plan
	err := yaml.Unmarshal(data, &plan)
	return plan, err
}

func (p *Plan) ToJSON() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Plan) ToYAML() ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(p); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
