package garoa

import (
	"log"
	"sync"
)

type step struct {
	degree int
	input  chan interface{}
	output chan interface{}
	action PipelineFunc
}

// Pipeline represents a series of steps, interconnected by channels. Values will be passed through channels
// interconnecting the steps, as long as a step doesn't return an error while processing a value.
//
// Errors are logged and the value is dropped from the pipeline, not being passed to the following step.
//
// When the output from the last step is discarded, the `output` channel is nil.
type Pipeline struct {
	input     chan interface{}
	output    chan interface{}
	steps     []step
	done      chan signal
	stepsDone chan signal
}

// Run starts the pipeline. All the steps start communicating and the values from the input flow through
// the steps of the pipeline.
//
// Run returns a `chan signal` which returns a single `signal` value when all the input has been consumed
// and all steps have finished processing. Notice it is still possible for the final output channel
// to be fully filled which will cause the Pipeline to block. Values are not discarded after the last step
// runs, they must be treated.
func (pipeline Pipeline) Run() <-chan signal {
	numSteps := len(pipeline.steps)

	wg := sync.WaitGroup{}
	wg.Add(numSteps)

	for i := 0; i < numSteps; i++ {
		pipeline.runStep(pipeline.steps[i], &wg)
	}

	pipeline.done = make(chan signal)

	go func() {
		wg.Wait()
		pipeline.done <- signal{}
	}()

	return pipeline.done
}

func (pipeline Pipeline) runStep(step step, externalWg *sync.WaitGroup) {
	internalWg := sync.WaitGroup{}
	internalWg.Add(step.degree)

	for i := 0; i < step.degree; i++ {
		go func() {
			for value := range step.input {
				v, err := step.action(value)
				if err == nil && step.output != nil && v != nil { // there are no errors, the return value is non-nil and there is a valid output channel
					step.output <- v
				} else if err != nil {
					log.Printf("Error: '%v' applying action to %v\n", err, value)
				} else {
					// not an error, but either there's no output channel (discard) or there's no return value to send into it (skip)
				}
			}
			internalWg.Done()
		}()
	}

	go func() {
		internalWg.Wait()
		if step.output != nil {
			close(step.output)
		}
		externalWg.Done()
	}()
}
