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
    delay: 100ms to 200ms
    execution:
    - delay:
        min: 100ms
        max: 200ms
    - call:
        http: GET service2/product 200 0 1024
        delay: 150ms
- delay:
    min: 150ms
    max: 250ms
- delay: 10ms
`

var indentShorter = `
execution:
- call:
  http: GET service1/listing 200 0 10240
  delay: 100ms to 200ms
  execution:
  - delay:
    min: 100ms
    max: 200ms
  - call:
    http: GET service2/product 200 0 1024
    delay: 150ms
- delay:
  min: 150ms
  max: 250ms
- delay: 10ms
`

func TestDecodeYAMLIndent(t *testing.T) {
	planShorter := load(t, indentShorter)
	planRegular := load(t, indentRegular)
	assert.Equal(t, planRegular, planShorter)
}

var example1 string = `
execution:
- delay: 10ms                                               # 0
- call:                                                     # 1
  http: GET service1/listing 200 0 10240
  delay: 100ms to 200ms
  execution:
  - delay: 1ms to 5ms                                       # 1_0
  - call:                                                   # 1_1
    http: GET service3/profile?id=some_user 400 0 100
  - parallel:                                               # 1_2
    concurrency: 2
    execution:
    - call:                                                # 1_2_0
      http: GET service2/product?id=1 200 0 1024
      delay: 50ms to 200ms
    - call:                                                # 1_2_1
      http: GET service2/product?id=2 200 0 1024
      delay: 51ms to 201ms
    - call:                                                # 1_2_2
      http: GET service2/product?id=3 502 0 1024
      delay: 52ms to 202ms
    - delay:                                               # 1_2_3
      min: 1s
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
      delay: 53ms to 203ms
  - delay: 10ms to 20ms                                    # 1_3
- delay:                                                   # 2
  min: 150ms
  max: 250ms
- call:                                                    # 3
  http: POST service5/metrics 201 1024 100
  delay: 10ms
`

func TestDecodeYAMLExample1(t *testing.T) {
	plan := load(t, example1)

	execution := plan.Execution
	assert.Equal(t, 4, len(execution))

	delay_0 := execution[0].(*Delay)
	assert.Equal(t,
		&Delay{Min: 10 * time.Millisecond, Max: 10 * time.Millisecond},
		delay_0,
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
		Delay{Min: 100 * time.Millisecond, Max: 200 * time.Millisecond},
		call_1.Delay,
	)

	assert.Equal(t, len(call_1.Execution), 4)

	delay_1_0 := call_1.Execution[0].(*Delay)
	assert.Equal(t,
		&Delay{Min: 1 * time.Millisecond, Max: 5 * time.Millisecond},
		delay_1_0,
	)

	call_1_1 := call_1.Execution[1].(*Call)
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
		Delay{Min: 0, Max: 0},
		call_1_1.Delay,
	)

	parallel_1_2 := call_1.Execution[2].(*Parallel)
	assert.Equal(t, 2, parallel_1_2.Concurrency)
	assert.Equal(t, 5, len(parallel_1_2.Execution))

	call_1_2_0 := parallel_1_2.Execution[0].(*Call)
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
		Delay{Min: 50 * time.Millisecond, Max: 200 * time.Millisecond},
		call_1_2_0.Delay,
	)

	call_1_2_1 := parallel_1_2.Execution[1].(*Call)
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
		Delay{Min: 51 * time.Millisecond, Max: 201 * time.Millisecond},
		call_1_2_1.Delay,
	)

	call_1_2_2 := parallel_1_2.Execution[2].(*Call)
	assert.Equal(t,
		HTTP{
			Method:          "GET",
			URL:             MustParseURL("service2/product?id=3"),
			StatusCode:      502,
			GenRequestBody:  0,
			GenResponseBody: 1024,
		},
		call_1_2_2.HTTP,
	)
	assert.Equal(t,
		Delay{Min: 52 * time.Millisecond, Max: 202 * time.Millisecond},
		call_1_2_2.Delay,
	)

	delay_1_2_3 := parallel_1_2.Execution[3].(*Delay)
	assert.Equal(t,
		&Delay{Min: 1 * time.Second},
		delay_1_2_3,
	)

	call_1_2_4 := parallel_1_2.Execution[4].(*Call)
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
		Delay{Min: 53 * time.Millisecond, Max: 203 * time.Millisecond},
		call_1_2_4.Delay,
	)

	delay_1_3 := call_1.Execution[3].(*Delay)
	assert.Equal(t,
		&Delay{Min: 10 * time.Millisecond, Max: 20 * time.Millisecond},
		delay_1_3,
	)

	delay_2 := execution[2].(*Delay)
	assert.Equal(t,
		&Delay{Min: 150 * time.Millisecond, Max: 250 * time.Millisecond},
		delay_2,
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
