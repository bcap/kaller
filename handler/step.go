package handler

import (
	"bytes"
	"math/rand"
	"net/http"
	"time"

	ptype "github.com/bcap/caller/plan"
	"github.com/bcap/caller/random"
)

var delayRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (h *handler) delay(delay ptype.Delay) error {
	if delay.IsZero() {
		return nil
	}
	sleep := delay.Min
	if delay.Min != delay.Max {
		delta := int64(delay.Max - delay.Min)
		sleep = delay.Min + time.Duration(delayRand.Int63n(delta))
	}
	select {
	case <-time.After(sleep):
	case <-h.Context.Done():
	}
	return nil
}

func (h *handler) call(call ptype.Call, location string) error {
	client := http.Client{}
	var body *bytes.Buffer
	if call.HTTP.RequestBody != "" {
		body = bytes.NewBufferString(call.HTTP.RequestBody)
	} else if call.HTTP.GenRequestBody > 0 {
		str := random.RandString(call.HTTP.GenRequestBody)
		body = bytes.NewBufferString(str)
	} else {
		body = &bytes.Buffer{}
	}
	req, err := http.NewRequestWithContext(
		h.Context, call.HTTP.Method, call.HTTP.URL.String(), body,
	)
	if err != nil {
		return err
	}
	for key, value := range call.HTTP.RequestHeaders {
		req.Header.Set(key, value)
	}
	if err := WritePlanHeaders(req, h.Plan, location); err != nil {
		return err
	}
	WriteRequestTraceHeader(req, h.RequestID)
	_, err = client.Do(req)
	return err
}

func (h *handler) loop(loop ptype.Loop, location string) error {
	for i := 0; i < loop.Times; i++ {
		if err := h.processSteps(1, 0, loop.Execution, location); err != nil {
			return err
		}
		h.delay(loop.Delay)
	}
	return nil
}
