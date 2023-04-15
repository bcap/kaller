package handler

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ptype "github.com/bcap/caller/plan"
)

var plan1 = `
execution:
- call:
  http: GET {{addr}}/service1 200 0 10240
  execution:
  - call:
    http: GET {{addr}}/service2 200 0 1024
    delay: 100ms
  post-execution:
  - call:
    http: POST {{addr}}/service3 200 1024 10240
    delay: 100ms
`

func TestHandlerPlan1(t *testing.T) {
	ctx, cancel, handler, addr := launchServer(t)
	defer cancel()

	execPlan(t, ctx, handler, addr, plan1)
	time.Sleep(300 * time.Millisecond)

	// assert the access log
	accessLog := handler.testAccessLog
	assertInLog(t, accessLog, "GET / 0 -> 200 0", 1)
	assertInLog(t, accessLog, "GET /service2 0 -> 200 1024", 1)
	assertInLog(t, accessLog, "POST /service3 1024 -> 200 10240", 1)
}

var plan2 = `
execution:
- call:
  http: POST {{addr}}/service4/login 200 1024 2048
  delay: 100ms to 200ms
- call:
  http: GET {{addr}}/service1/listing 200 0 10240
  execution:
  - delay: 100ms to 200ms
  - parallel:
    concurrency: 2
    execution:
    - call:
      http: GET {{addr}}/service2/product?id=1 200 0 1024
      delay: 500ms
    - loop:
      times: 3
      delay: 20ms
      execution:
      - call:
        http: GET {{addr}}/service2/product?id=2 404 0 1024
        delay: 50ms
    - call:
      http: GET {{addr}}/service2/product?id=3 200 0 1024
      delay: 500ms
    - call:
      http: GET {{addr}}/service2/product?id=4 200 0 1024
      delay: 500ms
    - call:
      async: true
      http: POST {{addr}}/service4/viewed 200 1024 2048
      delay: 2000ms
  post-execution:
  - call:
    http: POST {{addr}}/service3/metrics 200 1024 10240
    delay: 100ms
`

func TestHandlerPlan2(t *testing.T) {
	ctx, cancel, handler, addr := launchServer(t)
	defer cancel()

	execPlan(t, ctx, handler, addr, plan2)

	// assert the access log
	accessLog := handler.testAccessLog
	assertInLog(t, accessLog, "GET / 0 -> 200 0", 1)
	assertInLog(t, accessLog, "GET /service1/listing 0 -> 200 10240", 1)
	assertInLog(t, accessLog, "GET /service2/product?id=1 0 -> 200 1024", 1)
	assertInLog(t, accessLog, "GET /service2/product?id=2 0 -> 404 1024", 3)
	assertInLog(t, accessLog, "GET /service2/product?id=3 0 -> 200 1024", 1)
	assertInLog(t, accessLog, "GET /service2/product?id=4 0 -> 200 1024", 1)
	assertInLog(t, accessLog, "POST /service3/metrics 1024 -> 200 10240", 1)
}

func assertInLog(t *testing.T, accessLog []string, msg string, times int) {
	found := 0
	for _, entry := range accessLog {
		if strings.Contains(entry, msg) {
			found++
		}
	}
	assert.Equal(
		t, times, found,
		"access log should have %d entries for %q, but has %d instead",
		times, msg, found,
	)
}

//
// Helper functions
//

func launchServer(t *testing.T) (context.Context, context.CancelFunc, *Handler, *net.TCPAddr) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := Handler{testCaptureAccessLog: true}
	server := http.Server{
		Handler:     &handler,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", ":0")
	require.NoError(t, err)
	go func() {
		server.Serve(listener)
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	return ctx, cancel, &handler, listener.Addr().(*net.TCPAddr)
}

func execPlan(t *testing.T, ctx context.Context, handler *Handler, addr *net.TCPAddr, planString string) {
	request, err := http.NewRequestWithContext(ctx, "GET", "http://"+addr.AddrPort().String(), nil)
	require.NoError(t, err)

	plan := preparePlan(t, planString, addr)
	err = WritePlanHeaders(request, plan, "")
	require.NoError(t, err)

	client := http.Client{}
	_, err = client.Do(request)
	require.NoError(t, err)

	waitRequestsHandled(handler)
}

func preparePlan(t *testing.T, planStr string, addr *net.TCPAddr) ptype.Plan {
	planStr = strings.ReplaceAll(planStr, "{{addr}}", "http://"+addr.AddrPort().String())
	plan, err := ptype.FromYAML([]byte(planStr))
	require.NoError(t, err)
	return plan
}

func waitRequestsHandled(handler *Handler) {
	for handler.Outstanding() > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
