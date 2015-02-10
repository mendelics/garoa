package garoa

import "log"

type step struct {
	degree    int
	input     chan interface{}
	output    chan interface{}
	interrupt chan signal
	action    PipelineFunc
	done      chan signal
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
	interrupt chan signal
}

// Run starts the pipeline. All the steps start communicating and the values from the input flow through
// the steps of the pipeline.
//
// Run returns a `chan signal` which returns a single `signal` value when all the input has been consumed
// and all steps have finished processing. Notice it is still possible for the final output channel
// to be fully filled which will cause the Pipeline to block. Values are not discarded after the last step
// runs, they must be treated.
func (pipeline *Pipeline) Run() chan signal {
	numSteps := len(pipeline.steps)

	pipeline.done = make(chan signal)
	pipeline.stepsDone = make(chan signal, numSteps)
	pipeline.interrupt = make(chan signal)

	for i := 0; i < numSteps; i++ {
		pipeline.steps[i].interrupt = make(chan signal)
		pipeline.runStep(pipeline.steps[i])
	}

	go func() {
		<-pipeline.interrupt            // wait for an interrupt to the whole pipeline
		for i := 0; i < numSteps; i++ { // if it comes, tell all steps to stop
			step := pipeline.steps[i]
			for j := 0; j < step.degree; j++ {
				pipeline.steps[i].interrupt <- signal{}
			}
		}
	}()

	go func() {
		acked := 0
		for _ = range pipeline.stepsDone { // wait for the done signal from each step, should come even on interrupt
			acked++
			if acked == numSteps {
				close(pipeline.stepsDone)
				pipeline.done <- signal{}
			}
		}
	}()

	return pipeline.done
}

func (pipeline *Pipeline) Interrupt() {
	pipeline.interrupt <- signal{}
}

func (pipeline *Pipeline) runStep(step step) {

	step.done = make(chan signal, step.degree)

	for i := 0; i < step.degree; i++ {
		go func() {
		loop:
			for {
				select {
				case <-step.interrupt:
					// stop consuming step.input channel, break out of the loop
					break loop
				case value, ok := <-step.input:
					if !ok { // channel closed
						break loop
					}
					v, err := step.action(value)
					if err == nil && step.output != nil && v != nil { // there are no errors, the return value is non-nil and there is a valid output channel
						step.output <- v
					} else if err != nil {
						log.Printf("Error: '%v' applying action to %v\n", err, value)
					} else {
						// silently discard, either there is no output channel or a nil value has been received from the action
					}
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
