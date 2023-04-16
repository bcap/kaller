package plan

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var indentRegular = `
execution:
- call:
    http: GET service1/listing 200 0 10240
    compute: 100ms to 200ms
    execution:
    - compute:
        min: 100ms
        max: 200ms
    - call:
        http: GET service2/product 200 0 1024
        compute: 150ms
- compute:
    min: 150ms
    max: 250ms
- compute: 10ms
`

var indentShorter = `
execution:
- call:
  http: GET service1/listing 200 0 10240
  compute: 100ms to 200ms
  execution:
  - compute:
    min: 100ms
    max: 200ms
  - call:
    http: GET service2/product 200 0 1024
    compute: 150ms
- compute:
  min: 150ms
  max: 250ms
- compute: 10ms
`

func TestDecodeYAMLIndent(t *testing.T) {
	planShorter := load(t, indentShorter)
	planRegular := load(t, indentRegular)
	assert.Equal(t, planRegular, planShorter)
}

var example1 string = `
execution:
- compute: 10ms 1.0 cpu                                    # 0
- call:                                                    # 1
  http: GET service1/listing 200 0 10240
  compute: 100ms to 200ms
  execution:
  - compute: 1ms to 5ms                                    # 1_0
  - call:                                                  # 1_1
    http: GET service3/profile?id=some_user 400 0 100
  - parallel:                                              # 1_2
    concurrency: 2
    execution:
    - call:                                                # 1_2_0
      http: GET service2/product?id=1 200 0 1024
      compute: 50ms to 200ms 1.2cpu +1mb
    - call:                                                # 1_2_1
      async: true
      http: GET service2/product?id=2 200 0 1024
      compute: 51ms to 201ms 1.3 cpu -100kb
    - loop:                                                # 1_2_2
      times: 2
      compute: 10ms -200kb
      execution:
      - call:                                              # 1_2_2_0
        http: GET service2/product?id=3 502 0 1024
        compute: 52ms to 202ms
    - compute:                                             # 1_2_3
      min: 1s
      memory-delta-kb: +10
    - call:                                                # 1_2_4
      http:
        method: GET
        url: service2/product?id=4
        status-code: 200
        gen-request-body: 0
        gen-response-body: 1024
        request-headers:
          A: foo
          Accept: text/plain
        response-headers:
          B: bar
          Content-Type: text/plain
      compute: 53ms to 203ms
  - compute: 10ms to 20ms                                  # 1_3
  post-execution:
  - call:
    http: POST service5/metrics 201 2048 200               # 1_4
    compute: 10ms to 20ms
- compute:                                                 # 2
  min: 150ms
  max: 250ms
- call:                                                    # 3
  http: POST service5/metrics 201 1024 100
  compute: 10ms
`

func TestDecodeYAMLExample1(t *testing.T) {
	plan := load(t, example1)

	execution := plan.Execution
	assert.Equal(t, 4, len(execution))

	compute_0 := execution[0].(*Compute)
	assert.Equal(t,
		&Compute{Min: 10 * time.Millisecond, Max: 10 * time.Millisecond, CPU: 1.0},
		compute_0,
	)

	call_1 := execution[1].(*Call)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service1/listing"),
			StatusCode:      200,
			GenRequestBody:  0,
			GenResponseBody: 10240,
		},
		call_1.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 100 * time.Millisecond, Max: 200 * time.Millisecond},
		call_1.Compute,
	)

	assert.Equal(t, len(call_1.Execution), 4)

	compute_1_0 := call_1.Execution[0].(*Compute)
	assert.Equal(t,
		&Compute{Min: 1 * time.Millisecond, Max: 5 * time.Millisecond},
		compute_1_0,
	)

	call_1_1 := call_1.Execution[1].(*Call)
	assert.Equal(t, call_1_1.Async, false)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service3/profile?id=some_user"),
			StatusCode:      400,
			GenRequestBody:  0,
			GenResponseBody: 100,
		},
		call_1_1.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 0, Max: 0},
		call_1_1.Compute,
	)

	parallel_1_2 := call_1.Execution[2].(*Parallel)
	assert.Equal(t, 2, parallel_1_2.Concurrency)
	assert.Equal(t, 5, len(parallel_1_2.Execution))

	call_1_2_0 := parallel_1_2.Execution[0].(*Call)
	assert.Equal(t, call_1_2_0.Async, false)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service2/product?id=1"),
			StatusCode:      200,
			GenRequestBody:  0,
			GenResponseBody: 1024,
		},
		call_1_2_0.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 50 * time.Millisecond, Max: 200 * time.Millisecond, CPU: 1.2, MemoryDeltaKB: 1024},
		call_1_2_0.Compute,
	)

	call_1_2_1 := parallel_1_2.Execution[1].(*Call)
	assert.Equal(t, call_1_2_1.Async, true)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service2/product?id=2"),
			StatusCode:      200,
			GenRequestBody:  0,
			GenResponseBody: 1024,
		},
		call_1_2_1.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 51 * time.Millisecond, Max: 201 * time.Millisecond, CPU: 1.3, MemoryDeltaKB: -100},
		call_1_2_1.Compute,
	)

	loop_1_2_2 := parallel_1_2.Execution[2].(*Loop)
	assert.Equal(t, 2, loop_1_2_2.Times)
	assert.Equal(t,
		Compute{Min: 10 * time.Millisecond, Max: 10 * time.Millisecond, MemoryDeltaKB: -200},
		loop_1_2_2.Compute,
	)

	call_1_2_2_0 := loop_1_2_2.Execution[0].(*Call)
	assert.Equal(t, call_1_2_2_0.Async, false)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service2/product?id=3"),
			StatusCode:      502,
			GenRequestBody:  0,
			GenResponseBody: 1024,
		},
		call_1_2_2_0.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 52 * time.Millisecond, Max: 202 * time.Millisecond},
		call_1_2_2_0.Compute,
	)

	compute_1_2_3 := parallel_1_2.Execution[3].(*Compute)
	assert.Equal(t,
		&Compute{Min: 1 * time.Second, MemoryDeltaKB: 10},
		compute_1_2_3,
	)

	call_1_2_4 := parallel_1_2.Execution[4].(*Call)
	assert.Equal(t, call_1_2_4.Async, false)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service2/product?id=4"),
			StatusCode:      200,
			GenRequestBody:  0,
			GenResponseBody: 1024,
			RequestHeaders: map[string]string{
				"A":      "foo",
				"Accept": "text/plain",
			},
			ResponseHeaders: map[string]string{
				"B":            "bar",
				"Content-Type": "text/plain",
			},
		},
		call_1_2_4.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 53 * time.Millisecond, Max: 203 * time.Millisecond},
		call_1_2_4.Compute,
	)

	compute_1_3 := call_1.Execution[3].(*Compute)
	assert.Equal(t,
		&Compute{Min: 10 * time.Millisecond, Max: 20 * time.Millisecond},
		compute_1_3,
	)

	call_1_4 := call_1.PostExecution[0].(*Call)
	assert.Equal(t, call_1_4.Async, false)
	assert.Equal(t,
		HTTP{
			Method:          "POST",
			URL:             MustParseURL("service5/metrics"),
			StatusCode:      201,
			GenRequestBody:  2048,
			GenResponseBody: 200,
		},
		call_1_4.HTTP,
	)
	assert.Equal(t,
		Compute{Min: 10 * time.Millisecond, Max: 20 * time.Millisecond},
		call_1_4.Compute,
	)

	compute_2 := execution[2].(*Compute)
	assert.Equal(t,
		&Compute{Min: 150 * time.Millisecond, Max: 250 * time.Millisecond},
		compute_2,
	)
}

var withAnchors string = `
anchors:
- call: &call1
    http: GET service1/listing 200 0 10240
    compute: 100ms to 200ms 0.2 cpu +1mb
- call: &call2
    http: POST service2/metrics 200 1024 2048
    compute: 50ms to 80ms 0.1cpu +100kb

execution:
- call: *call1
- call: *call2
`

func TestDecodeYAMLWithAnchors(t *testing.T) {
	plan := load(t, withAnchors)

	execution := plan.Execution
	require.Equal(t, 2, len(execution))

	call_0 := execution[0].(*Call)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service1/listing"),
			StatusCode:      200,
			GenRequestBody:  0,
			GenResponseBody: 10240,
		},
		call_0.HTTP,
	)
	assert.Equal(t,
		Compute{
			Min:           100 * time.Millisecond,
			Max:           200 * time.Millisecond,
			CPU:           0.2,
			MemoryDeltaKB: 1024},
		call_0.Compute,
	)

	call_1 := execution[1].(*Call)
	assert.Equal(t,
		HTTP{
			Method:          "POST",
			URL:             MustParseURL("service2/metrics"),
			StatusCode:      200,
			GenRequestBody:  1024,
			GenResponseBody: 2048,
		},
		call_1.HTTP,
	)
	assert.Equal(t,
		Compute{
			Min:           50 * time.Millisecond,
			Max:           80 * time.Millisecond,
			CPU:           0.1,
			MemoryDeltaKB: 100},
		call_1.Compute,
	)
}

func TestEncodeDecodeYAML(t *testing.T) {
	plan := load(t, example1)
	encoded, err := plan.ToYAML()
	require.NoError(t, err)
	fmt.Println(string(encoded))
	decoded, err := FromYAML(encoded)
	require.NoError(t, err)
	assert.Equal(t, plan, decoded)
}

func TestEncodeDecodeJSON(t *testing.T) {
	plan := load(t, example1)
	encoded, err := plan.ToJSON()
	require.NoError(t, err)
	fmt.Println(string(encoded))
	decoded, err := FromJSON(encoded)
	require.NoError(t, err)
	assert.Equal(t, plan, decoded)
}

//
// Auxiliary functions
//

func load(t *testing.T, yaml string) Plan {
	data := []byte(strings.TrimSpace(yaml))
	plan, err := FromYAML(data)
	require.NoError(t, err)
	return plan
}
