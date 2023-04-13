package handler

import (
	"fmt"
	"strconv"

	ptype "github.com/bcap/caller/plan"
	"golang.org/x/sync/errgroup"
)

func (h *handler) processSteps(concurrency int, stepIdxOffset int, execution ptype.Execution, location string) error {
	if concurrency == 1 {
		for stepIdx, step := range execution {
			if err := h.processStep(stepIdxOffset+stepIdx, step, location); err != nil {
				return err
			}
		}
		return nil
	}

	if concurrency <= 0 {
		concurrency = len(execution)
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
	case *ptype.Parallel:
		err = h.parallel(*v, nextLocation())
	case *ptype.Loop:
		err = h.loop(*v, nextLocation())
	case *ptype.Delay:
		err = h.delay(*v)
	case *ptype.Call:
		err = h.call(*v, nextLocation())
	default:
		return fmt.Errorf("unrecognized step type %T", step)
	}
	if err != nil {
		return fmt.Errorf("failed at step %d: %w", stepIdx, err)
	}
	return nil
}
