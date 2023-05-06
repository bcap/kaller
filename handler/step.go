package handler

import (
	"bytes"
	"log"
	"net/http"

	ptype "github.com/bcap/caller/plan"
	"github.com/bcap/caller/random"
	"golang.org/x/sync/errgroup"
)

func (h *handler) parallel(parallel ptype.Parallel, location string) error {
	return h.processSteps(parallel.Concurrency, 0, parallel.Execution, location)
}

func (h *handler) loop(loop ptype.Loop, location string) error {
	do := func() error {
		if err := h.processSteps(1, 0, loop.Execution, location); err != nil {
			return err
		}
		h.compute(loop.Compute)
		return nil
	}

	concurrency := loop.Concurrency
	if concurrency <= 1 {
		for i := 0; i < loop.Times; i++ {
			if err := do(); err != nil {
				return err
			}
		}
		return nil
	}

	if concurrency > loop.Times {
		concurrency = loop.Times
	}
	group, ctx := errgroup.WithContext(h.Context)
	runCh := make(chan struct{})
	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case _, ok := <-runCh:
					if !ok {
						return nil
					}
					if err := do(); err != nil {
						return err
					}
				}
			}
		})
	}
	for i := 0; i < loop.Times; i++ {
		runCh <- struct{}{}
	}
	close(runCh)
	return group.Wait()
}

func (h *handler) compute(compute ptype.Compute) error {
	compute.Do(h.Context, &h.Fill)
	return nil
}

func (h *handler) call(call ptype.Call, location string) error {
	execute := func() error {
		client := http.Client{}
		var body *bytes.Buffer
		if call.HTTP.RequestBody != "" {
			body = bytes.NewBufferString(call.HTTP.RequestBody)
		} else if call.HTTP.GenRequestBody > 0 {
			str := random.String(call.HTTP.GenRequestBody)
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
		if h.EncodedPlan != nil {
			WriteEncodedPlanHeaders(req, h.EncodedPlan, location)
		} else if err := WritePlanHeaders(req, h.Plan, location); err != nil {
			return err
		}
		WriteRequestTraceHeader(req, h.RequestID)
		_, err = client.Do(req)
		return err
	}

	if call.Async {
		h.pendingAsyncCalls.Add(1)
		go func() {
			err := execute()
			if err != nil {
				log.Printf("!! async call failed: %v", err)
			}
			h.pendingAsyncCalls.Done()
		}()
		return nil
	} else {
		return execute()
	}
}
