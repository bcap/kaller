package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	ptype "github.com/bcap/caller/plan"
)

const HeaderPlan = "X-caller-plan"
const HeaderLocation = "X-caller-loc"
const HeaderPlanEncoding = "X-caller-plan-encoding"
const HeaderRequestTrace = "X-caller-request-trace"

func WritePlanHeaders(req *http.Request, plan ptype.Plan, location string) error {
	encodedPlan, encoding, err := EncodePlan(plan)
	if err != nil {
		return fmt.Errorf("cannot write plan headers: %w", err)
	}
	req.Header.Set(HeaderPlan, encodedPlan)
	req.Header.Set(HeaderPlanEncoding, encoding)
	req.Header.Set(HeaderLocation, location)
	return nil
}

func ReadPlanHeaders(req *http.Request) (ptype.Plan, string, error) {
	encodedPlan := req.Header.Get(HeaderPlan)
	plan, err := DecodePlan(encodedPlan, req.Header.Get(HeaderPlanEncoding))
	if err != nil {
		return ptype.Plan{}, "", fmt.Errorf("cannot read plan from headers: %w", err)
	}
	location := req.Header.Get(HeaderLocation)
	return plan, location, nil
}

func EncodePlan(plan ptype.Plan) (string, string, error) {
	jsonBytes, err := plan.ToJSON()
	if err != nil {
		return "", "", err
	}
	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(jsonBytes)))
	base64.RawStdEncoding.Encode(encoded, jsonBytes)
	return string(encoded), "json; base64/no-padding", nil
}

func DecodePlan(encodedPlan string, encoding string) (ptype.Plan, error) {
	encodedPlanBytes := []byte(encodedPlan)
	encodingChain := strings.Split(encoding, ";")
	if len(encodingChain) == 0 {
		return ptype.Plan{}, errors.New("empty encoding")
	}

	for idx := len(encodingChain) - 1; idx >= 0; idx-- {
		codec := strings.ToLower(strings.TrimSpace(encodingChain[idx]))
		switch codec {

		// base64 is an intermediary encoding
		case "base64/no-padding":
			buf := make([]byte, base64.RawStdEncoding.DecodedLen(len(encodedPlan)))
			n, err := base64.RawStdEncoding.Decode(buf, encodedPlanBytes)
			if err != nil {
				return ptype.Plan{}, err
			}
			encodedPlanBytes = buf[:n]

		// json is a final encoding
		case "json":
			return ptype.FromJSON(encodedPlanBytes)

		// yaml is a final encoding
		case "yaml":
			return ptype.FromYAML(encodedPlanBytes)
		}
	}

	return ptype.Plan{}, fmt.Errorf("invalid encoding %s", encoding)
}

func WriteRequestTraceHeader(req *http.Request, requestIDs string) {
	req.Header.Set(HeaderRequestTrace, requestIDs)
}

func ReadRequestTraceHeader(req *http.Request) string {
	return req.Header.Get(HeaderRequestTrace)
}
