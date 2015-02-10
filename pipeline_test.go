package garoa

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PipelineInterruptSuite struct {
	suite.Suite

	builder *PipelineBuilder

	pipe       *Pipeline
	buildError error
}

func (s *PipelineInterruptSuite) SetupTest() {
	s.builder = new(PipelineBuilder)

	mockedInput := make(chan interface{})

	runForever := func(i interface{}) (interface{}, error) {
		for {
			time.Sleep(time.Second) // sleep for one second so busy wait doesn't kill the system
		}
	}

	s.pipe, s.buildError = s.builder.CreateNew().
		ConsumingFrom(mockedInput).
		ThenRunning(runForever, 2).
		DiscardOutput().
		Build()
}

func (s *PipelineInterruptSuite) TestBuildShouldReturnAValidPipeline() {
	assert.NotNil(s.T(), s.pipe, "Pipeline should not be nil")
}

func (s *PipelineInterruptSuite) TestReturnsNoError() {
	assert.NoError(s.T(), s.buildError)
}

func (s *PipelineInterruptSuite) TestRunsForeverWithoutInterrupt() {
	fmt.Println("Entering long test. Expected to run for 3 seconds if successful.")
	pipeFinished := s.pipe.Run()
	select {
	case <-pipeFinished:
		assert.Fail(s.T(), "pipe finished, it should stay on an infinite loop forever")
	case <-time.After(3 * time.Second):
		// run for 3 seconds, it's enough
		return
	}
}

func (s *PipelineInterruptSuite) TestInterrupt() {
	fmt.Println("Entering long test. Will take from 2 to 5 seconds.")
	pipeFinished := s.pipe.Run()
	time.Sleep(2 * time.Second) // let it run a little bit before interrupting
	s.pipe.Interrupt()
	select {
	case <-pipeFinished:
		return
	case <-time.After(3 * time.Second):
		assert.Fail(s.T(), "pipe didn't get interrupted")
	}
}

func TestRunPipelineSuites(t *testing.T) {
	suite.Run(t, new(PipelineInterruptSuite))
}
