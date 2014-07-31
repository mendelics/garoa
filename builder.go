package pipeline

import "errors"

func constructPipeline() *Pipeline {
	newPipeline := new(Pipeline)
	newPipeline.steps = make([]Step, 0, 32)
	return newPipeline
}

type Builder struct {
	buildingPipeline *Pipeline
}

func (builder *Builder) CreateNew() *Builder {
	builder.buildingPipeline = constructPipeline()
	return builder
}

func (builder *Builder) ConsumingFrom(input chan interface{}) *Builder {
	if builder.buildingPipeline != nil {
		builder.buildingPipeline.input = input
	}
	return builder
}

type pipelineFunc func(interface{}) (interface{}, error)

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

		step := Step{action: action, degree: parallelismDegree, input: input, output: nil}

		builder.buildingPipeline.steps = append(builder.buildingPipeline.steps, step)
	}
	return builder
}

func (builder *Builder) OutputtingTo(output chan interface{}) *Builder {
	if builder.buildingPipeline != nil {
		builder.buildingPipeline.output = output
		if len(builder.buildingPipeline.steps) > 0 {
			builder.buildingPipeline.steps[len(builder.buildingPipeline.steps)-1].output = builder.buildingPipeline.output
		}
	}
	return builder
}

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
