package plan

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

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
