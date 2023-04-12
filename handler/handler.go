package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	ptype "github.com/bcap/caller/plan"
	"github.com/bcap/caller/random"
)

type Handler struct {
	// access log capturing is for unit testing only
	testCaptureAccessLog bool
	testAccessLog        []string
	testAccessLogMutex   sync.Mutex
}

type handler struct {
	*Handler

	Context context.Context

	RequestID   string
	Request     *http.Request
	RequestBody []byte

	Response           http.ResponseWriter
	ResponseStatusCode int
	ResponseBody       []byte

	Plan        ptype.Plan
	RequestedAt time.Time
	RespondedAt time.Time
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	handler := handler{
		Handler:  h,
		Request:  req,
		Response: resp,
	}
	handler.Handle()
}

func (h *handler) Handle() {
	h.RequestedAt = time.Now()
	h.Context = h.Request.Context()

	reqBodyBytes, err := io.ReadAll(h.Request.Body)
	if err != nil {
		h.textResponse(400, "bad request: %v", err)
		return
	}
	h.RequestBody = reqBodyBytes
	h.identifyRequest()

	h.logRequestIn()

	plan, location, err := ReadPlanHeaders(h.Request)
	if err != nil {
		h.textResponse(400, "bad plan: %v", err)
		return
	}
	h.Plan = plan

	step, err := locateInPlan(plan, location)
	if err != nil {
		h.textResponse(400, "bad location in plan: %v", err)
		return
	}
	call, ok := step.(*ptype.Call)
	if !ok {
		h.textResponse(400, "bad location in plan: %s is not a call", location, err)
		return
	}

	h.delay(call.Delay)
	err = h.processSteps(1, 0, call.Execution, location)
	if err != nil {
		h.textResponse(500, "execution failure: %v", err)
	}

	statusCode, respBodyBytes, err := h.respond(call)
	if err != nil {
		h.logResponseWriteErr(err)
		return
	}

	h.ResponseStatusCode = statusCode
	h.ResponseBody = respBodyBytes
	h.logResponseOut()

	//
	// post execution phase (what to run after the response was sent)
	//

	if len(call.PostExecution) == 0 {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.Context = ctx

	err = h.processSteps(1, len(call.Execution), call.PostExecution, location)
	if err != nil {
		h.textResponse(500, "execution failure: %v", err)
	}

	h.logPostResponseOut()
}

func (h *handler) processSteps(concurrency int, stepIdxOffset int, execution ptype.Execution, location string) error {
	if concurrency == 1 {
		for stepIdx, step := range execution {
			if err := h.processStep(stepIdxOffset+stepIdx, step, location); err != nil {
				return err
			}
		}
		return nil
	}

	group, ctx := errgroup.WithContext(h.Context)
	stepsC := make(chan int)
	for i := 0; i < concurrency; i++ {
		group.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case stepIdx, ok := <-stepsC:
					if !ok {
						return nil
					}
					if err := h.processStep(stepIdxOffset+stepIdx, execution[stepIdx], location); err != nil {
						return err
					}
				}
			}
		})
	}
	for i := 0; i < len(execution); i++ {
		stepsC <- i
	}
	close(stepsC)
	return group.Wait()
}

func (h *handler) processStep(stepIdx int, step ptype.Step, location string) error {
	nextLocation := func() string {
		stepIdxStr := strconv.Itoa(stepIdx)
		if location == "" {
			return stepIdxStr
		}
		return location + "." + stepIdxStr
	}

	var err error
	switch v := step.(type) {
	case *ptype.Delay:
		err = h.delay(*v)
	case *ptype.Call:
		err = h.call(*v, nextLocation())
	case *ptype.Parallel:
		err = h.processSteps(v.Concurrency, 0, v.Execution, nextLocation())
	default:
		return fmt.Errorf("unrecognized step type %T", step)
	}
	if err != nil {
		return fmt.Errorf("failed at step %d: %w", stepIdx, err)
	}
	return nil
}

func (h *handler) respond(call *ptype.Call) (int, []byte, error) {
	statusCode := call.HTTP.StatusCode
	if statusCode == 0 {
		statusCode = 200
	}
	h.Response.WriteHeader(statusCode)
	var body []byte
	if call.HTTP.ResponseBody != "" {
		body = []byte(call.HTTP.ResponseBody)
	} else if call.HTTP.GenResponseBody > 0 {
		body = []byte(random.RandString(call.HTTP.GenResponseBody))
	} else {
		body = []byte{}
	}
	_, err := h.Response.Write(body)
	h.RespondedAt = time.Now()
	return statusCode, body, err
}

func (h *handler) textResponse(statusCode int, msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	log.Print(msg)
	h.Response.Header().Set("Content-type", "text/plain")
	h.Response.WriteHeader(statusCode)
	h.Response.Write([]byte(msg))
}

func (h *handler) identifyRequest() {
	h.RequestID = ReadRequestTraceHeader(h.Request)
	newID := random.RandString(3)
	if h.RequestID == "" {
		h.RequestID = newID
	} else {
		h.RequestID = h.RequestID + "." + newID
	}
}

func locateInPlan(plan ptype.Plan, location string) (ptype.Step, error) {
	var step ptype.Step = &ptype.Call{Execution: plan.Execution}
	if location == "" {
		return step, nil
	}
	path := strings.Split(location, ".")
	for idx, stepIdxStr := range path {
		stepIdx, err := strconv.Atoi(stepIdxStr)
		if err != nil {
			return nil, fmt.Errorf("bad location %s: step #%d (%s) is not an integer", location, idx, stepIdxStr)
		}
		switch v := step.(type) {
		case *ptype.Call:
			if stepIdx < len(v.Execution) {
				step = v.Execution[stepIdx]
			} else {
				step = v.PostExecution[stepIdx-len(v.Execution)]
			}
		case *ptype.Parallel:
			step = v.Execution[stepIdx]
		default:
			return nil, fmt.Errorf("bad location %s: step #%d is of unrecognized type %T", location, idx, v)
		}
	}
	return step, nil
}
