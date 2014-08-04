package pipeline

import "errors"

func constructPipeline() *Pipeline {
	newPipeline := new(Pipeline)
	newPipeline.steps = make([]step, 0, 32)
	return newPipeline
}

// Builder contains an unfinished pipeline.
//
// Builder methods return a pointer in order to allow chaining
type Builder struct {
	buildingPipeline *Pipeline
}

// CreateNew starts the construction of a new Pipeline through a Builder
func (builder *Builder) CreateNew() *Builder {
	builder.buildingPipeline = constructPipeline()
	return builder
}

// ConsumingFrom specifies the initial channel that will be used for the Pipeline
func (builder *Builder) ConsumingFrom(input chan interface{}) *Builder {
	if builder.buildingPipeline != nil {
		builder.buildingPipeline.input = input
	}
	return builder
}

type pipelineFunc func(interface{}) (interface{}, error)

// ThenRunning specifies a new action to be run under the pipeline, with its corresponding parallelism degree. This
// degree specifies how many goroutines can be started to run the action.
//
// Input/ouput among the actions is determined by the order they are specified on the builder.
func (builder *Builder) ThenRunning(action pipelineFunc, parallelismDegree int) *Builder {

	if builder.buildingPipeline != nil {

		numberOfSteps := len(builder.buildingPipeline.steps)

		var input chan interface{}

		if numberOfSteps == 0 {
			input = builder.buildingPipeline.input
		} else {
			builder.buildingPipeline.steps[numberOfSteps-1].output = make(chan interface{}, 1)
			input = builder.buildingPipeline.steps[numberOfSteps-1].output
		}

		step := step{action: action, degree: parallelismDegree, input: input, output: nil}

		builder.buildingPipeline.steps = append(builder.buildingPipeline.steps, step)
	}
	return builder
}

// OutputtingTo specifies the last channel which will receive the output from the Pipeline, after all the actions are applied.
func (builder *Builder) OutputtingTo(output chan interface{}) *Builder {
	if builder.buildingPipeline != nil {
		builder.buildingPipeline.output = output
		if len(builder.buildingPipeline.steps) > 0 {
			builder.buildingPipeline.steps[len(builder.buildingPipeline.steps)-1].output = builder.buildingPipeline.output
		}
	}
	return builder
}

// Build verifies the construction of the Pipeline is complete and no pieces are missing and returns the finished Pipeline
// in a state where it can be run.
func (builder *Builder) Build() (*Pipeline, error) {

	if builder.buildingPipeline == nil {
		return nil, errors.New("should call CreateNew() before build")
	}

	if builder.buildingPipeline.input == nil {
		return nil, errors.New("should call ConsumingFrom() passing an already initialized channel before callig Build()")
	}

	if builder.buildingPipeline.output == nil {
		return nil, errors.New("should call OutputtingTo() passing an already initialized channel before callig Build()")
	}

	if len(builder.buildingPipeline.steps) == 0 {
		return nil, errors.New("cannot build a pipeline without actions")
	}

	return builder.buildingPipeline, nil
}
