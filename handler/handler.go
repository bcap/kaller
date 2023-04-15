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
	"sync/atomic"
	"time"

	ptype "github.com/bcap/caller/plan"
	"github.com/bcap/caller/random"
	syncx "github.com/bcap/caller/sync"
)

type Handler struct {
	requestsHandled     int64
	requestsOutstanding int32

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

	pendingAsyncCalls syncx.WaitGroup
}

// Main HTTP handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	handler := handler{
		Handler:  h,
		Request:  req,
		Response: resp,
	}
	atomic.AddInt32(&h.requestsOutstanding, 1)
	handler.Handle()
	atomic.AddInt64(&h.requestsHandled, 1)
	atomic.AddInt32(&h.requestsOutstanding, -1)
}

func (h *Handler) Outstanding() int32 {
	return atomic.LoadInt32(&h.requestsOutstanding)
}

func (h *Handler) Handled() int64 {
	return atomic.LoadInt64(&h.requestsHandled)
}

func (h *handler) Handle() {
	h.RequestedAt = time.Now()

	if h.Context == nil {
		h.Context = context.Background()
	}

	reqBodyBytes, err := io.ReadAll(h.Request.Body)
	if err != nil {
		h.textResponse(400, "bad request: %v", err)
		return
	}
	h.RequestBody = reqBodyBytes
	h.identifyRequest()

	plan, location, err := ReadPlanHeaders(h.Request)
	if err != nil {
		h.textResponse(400, "bad plan: %v", err)
		return
	}
	h.Plan = plan

	h.logRequestIn(location)

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

	defer h.waitAsyncCalls()

	err = h.processSteps(1, 0, call.Execution, location)
	if err != nil {
		h.textResponse(500, "execution failure: %v", err)
	}

	statusCode, respBodyBytes, err := h.respond(call)
	if err != nil {
		h.logResponseWriteErr(location, err)
		return
	}

	h.ResponseStatusCode = statusCode
	h.ResponseBody = respBodyBytes
	h.logResponseOut(location)

	//
	// post execution phase (executed after the response was sent)
	//

	if len(call.PostExecution) == 0 {
		return
	}

	err = h.processSteps(1, len(call.Execution), call.PostExecution, location)
	if err != nil {
		h.textResponse(500, "execution failure: %v", err)
	}

	h.logPostResponseOut(location)
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
		body = []byte(random.String(call.HTTP.GenResponseBody))
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
	h.RespondedAt = time.Now()
}

func (h *handler) identifyRequest() {
	h.RequestID = ReadRequestTraceHeader(h.Request)
	newID := random.String(3)
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
		case *ptype.Loop:
			step = v.Execution[stepIdx]
		default:
			return nil, fmt.Errorf("bad location %s: step #%d is of unrecognized type %T", location, idx, v)
		}
	}
	return step, nil
}

const AsyncCallWaitReportTime = 10 * time.Second

func (h *handler) waitAsyncCalls() {
	start := time.Now()
	done := make(chan struct{})
	go func() {
		h.pendingAsyncCalls.Wait()
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		case <-h.Context.Done():
			return
		case <-time.After(AsyncCallWaitReportTime):
			timeTaken := time.Since(start)
			log.Printf(
				"! Waiting on %d async calls for %v and counting",
				h.pendingAsyncCalls.Current(), timeTaken,
			)
		}
	}
}
