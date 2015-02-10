package garoa

import "log"

type step struct {
	degree int
	input  chan interface{}
	output chan interface{}
	action PipelineFunc
	done   chan signal
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
func (pipeline Pipeline) Run() chan signal {
	numSteps := len(pipeline.steps)

	pipeline.done = make(chan signal)
	pipeline.stepsDone = make(chan signal, numSteps)

	for i := 0; i < numSteps; i++ {
		pipeline.runStep(pipeline.steps[i])
	}

	go func() {
		acked := 0
		for _ = range pipeline.stepsDone {
			acked++
			if acked == numSteps {
				close(pipeline.stepsDone)
				pipeline.done <- signal{}
			}
		}
	}()

	return pipeline.done
}

func (pipeline Pipeline) runStep(step step) {

	step.done = make(chan signal, step.degree)

	for i := 0; i < step.degree; i++ {
		go func() {
			for value := range step.input {
				v, err := step.action(value)
				if err == nil && step.output != nil && v != nil { // there are no errors, the return value is non-nil and there is a valid output channel
					step.output <- v
				} else if err != nil {
					log.Printf("Error: '%v' applying action to %v\n", err, value)
				} else {
					// silently discard, either there is no output channel or a nil value has been received from the action
				}
			}
			step.done <- signal{}
		}()
	}

	go func() {
		acked := 0
		for _ = range step.done {
			acked++
			if acked == step.degree {
				if step.output != nil {
					close(step.output)
				}
				close(step.done)
				pipeline.stepsDone <- signal{}
			}
		}
	}()
}
