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

	ptype "github.com/bcap/kaller/plan"
	"github.com/bcap/kaller/server"
)

var plan1 = `
execution:
- call:
  http: GET {{addr}}/service1 200 0 10240
  execution:
  - call:
    http: GET {{addr}}/service2 200 0 1024
    compute: 100ms
  post-execution:
  - call:
    http: POST {{addr}}/service3 200 1024 10240
    compute: 100ms
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
  compute: 100ms to 200ms
- call:
  http: GET {{addr}}/service1/listing 200 0 10240
  execution:
  - compute: 100ms to 200ms
  - parallel:
    concurrency: 2
    execution:
    - call:
      http: GET {{addr}}/service2/product?id=1 200 0 1024
      compute: 500ms
    - loop:
      times: 3
      concurrency: 2
      compute: 20ms
      execution:
      - call:
        http: GET {{addr}}/service2/product?id=2 404 0 1024
        compute: 50ms
    - call:
      http: GET {{addr}}/service2/product?id=3 200 0 1024
      compute: 500ms
    - call:
      http: GET {{addr}}/service2/product?id=4 200 0 1024
      compute: 500ms
    - call:
      async: true
      http: POST {{addr}}/service4/viewed 200 1024 2048
      compute: 2000ms
  post-execution:
  - call:
    http: POST {{addr}}/service3/metrics 200 1024 10240
    compute: 100ms
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
	handler := New(ctx)
	handler.testCaptureAccessLog = true
	srv := server.Server{}
	addr, err := srv.Listen(ctx, ":0")
	require.NoError(t, err)
	go func() {
		err := srv.Serve(handler)
		if !server.IsClosedError(err) {
			assert.Fail(t, "server exit error: %v", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	return ctx, cancel, handler, addr
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
