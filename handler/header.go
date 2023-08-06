package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	ptype "github.com/bcap/kaller/plan"
)

const HeaderPlan = "X-kaller-plan"
const HeaderLocation = "X-kaller-loc"
const HeaderPlanEncoding = "X-kaller-plan-encoding"
const HeaderRequestTrace = "X-kaller-request-trace"

func WritePlanHeaders(req *http.Request, plan ptype.Plan, location string) error {
	encodedPlan, err := EncodePlan(plan)
	if err != nil {
		return fmt.Errorf("cannot write plan headers: %w", err)
	}
	WriteEncodedPlanHeaders(req, encodedPlan, location)
	return nil
}

func WriteEncodedPlanHeaders(req *http.Request, encodedPlan *EncodedPlan, location string) {
	req.Header.Set(HeaderPlan, encodedPlan.Content)
	req.Header.Set(HeaderPlanEncoding, encodedPlan.Encoding)
	req.Header.Set(HeaderLocation, location)
}

func ReadPlanHeaders(req *http.Request) (ptype.Plan, *EncodedPlan, string, error) {
	encodedPlan := req.Header.Get(HeaderPlan)
	encoding := req.Header.Get(HeaderPlanEncoding)
	plan, err := DecodePlan(encodedPlan, encoding)
	if err != nil {
		return ptype.Plan{}, nil, "", fmt.Errorf("cannot read plan from headers: %w", err)
	}
	location := req.Header.Get(HeaderLocation)
	return plan, &EncodedPlan{Content: encodedPlan, Encoding: encoding}, location, nil
}

func EncodePlan(plan ptype.Plan) (*EncodedPlan, error) {
	jsonBytes, err := plan.ToJSON()
	if err != nil {
		return nil, err
	}
	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(jsonBytes)))
	base64.RawStdEncoding.Encode(encoded, jsonBytes)
	return &EncodedPlan{Content: string(encoded), Encoding: "json; base64/no-padding"}, nil
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
