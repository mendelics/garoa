package pipeline

import "log"

type Step struct {
	degree int
	input  chan interface{}
	output chan interface{}
	action pipelineFunc
	done   chan signal
}

type Pipeline struct {
	input     chan interface{}
	output    chan interface{}
	steps     []Step
	done      chan signal
	stepsDone chan signal
}

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

func (pipeline Pipeline) runStep(step Step) {

	step.done = make(chan signal, step.degree)

	for i := 0; i < step.degree; i++ {
		go func() {
			for value := range step.input {
				v, err := step.action(value)
				if err == nil {
					step.output <- v
				} else {
					log.Printf("Error: '%v' applying action to %v\n", err, value)
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
				close(step.output)
				close(step.done)
				pipeline.stepsDone <- signal{}
			}
		}
	}()
}
