package garoa

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

/*
	Mock objects, general setup and a 'base' Suite
*/

var mockedInput = make(chan interface{})
var mockedOutput = make(chan interface{})

type MockedTestFunctionsContainer struct {
	mock.Mock
}

func (m *MockedTestFunctionsContainer) mockFunctionA(input interface{}) (interface{}, error) {
	args := m.Called(input)
	return args.Get(0), nil
}

func (m *MockedTestFunctionsContainer) mockFunctionB(input interface{}) (interface{}, error) {
	args := m.Called(input)
	return args.Get(0), nil
}

var expectedParallelismDegree = 10

type BasePipelineSuite struct {
	suite.Suite

	mockFunctions *MockedTestFunctionsContainer

	sut *PipelineBuilder

	result      *Pipeline
	resultError error
}

/*
	Test suite calling Build() without calling CreateNew() generates error and a nil Pipeline
*/
type MissingStepsBuildingPipelineShouldReturnError struct {
	BasePipelineSuite
}

func (s *MissingStepsBuildingPipelineShouldReturnError) SetupTest() {
	s.mockFunctions = new(MockedTestFunctionsContainer)
	s.sut = new(PipelineBuilder)
	s.result, s.resultError = s.sut.
		ConsumingFrom(mockedInput).
		ThenRunning(s.mockFunctions.mockFunctionA, expectedParallelismDegree).
		OutputtingTo(mockedOutput).
		Build()
}

func (s *MissingStepsBuildingPipelineShouldReturnError) TestCallingBuildWithoutCreateNewReturnsNilPipeline() {
	assert.Nil(s.T(), s.result, "Returning Pipeline should be nil")
}

func (s *MissingStepsBuildingPipelineShouldReturnError) TestCallingBuildWithoutCreateNewReturnsError() {
	assert.Error(s.T(), s.resultError, "Calling Build() without CreateNew() before should return non empty error")
}

/*
	Test suite calling Build() twice should generate a different Pipeline object
*/
type BuildingTwiceShouldGenerateDifferentPipelines struct {
	secondResult      *Pipeline
	secondResultError error
	BasePipelineSuite
}

type MockEmptyObject struct{}

func (s *BuildingTwiceShouldGenerateDifferentPipelines) SetupTest() {

	s.mockFunctions = new(MockedTestFunctionsContainer)
	s.mockFunctions.On("mockFunctionA", MockEmptyObject{}).Return(MockEmptyObject{})
	s.mockFunctions.On("mockFunctionB", MockEmptyObject{}).Return(MockEmptyObject{})

	s.sut = new(PipelineBuilder)

	s.result, s.resultError = s.sut.CreateNew().
		ConsumingFrom(mockedInput).
		ThenRunning(s.mockFunctions.mockFunctionA, expectedParallelismDegree).
		OutputtingTo(mockedOutput).
		Build()
	s.secondResult, s.secondResultError = s.sut.CreateNew().
		ConsumingFrom(mockedInput).
		ThenRunning(s.mockFunctions.mockFunctionA, expectedParallelismDegree).
		OutputtingTo(mockedOutput).
		Build()
}

func (s *BuildingTwiceShouldGenerateDifferentPipelines) TestPipelinesShouldBeDifferent() {
	assert.NoError(s.T(), s.resultError, "Calling twice should not raise errors")
	assert.NoError(s.T(), s.secondResultError, "Calling twice should generate not raise errors")
	assert.True(s.T(), s.result != s.secondResult, "Calling twice should generate different objects")
}

/*
	Building empty pipeline test suite. A nil Pipeline and a error should return
*/
type BuildingEmptyPipelineShouldReturnError struct {
	BasePipelineSuite
}

func (s *BuildingEmptyPipelineShouldReturnError) SetupTest() {
	s.sut = new(PipelineBuilder)
	s.result, s.resultError = s.sut.CreateNew().ConsumingFrom(mockedInput).OutputtingTo(mockedOutput).Build()
}

func (s *BuildingEmptyPipelineShouldReturnError) TestBuildShouldReturnANilPipeline() {
	assert.Nil(s.T(), s.result, "Returning Pipeline should not be nil")
}

func (s *BuildingEmptyPipelineShouldReturnError) TestResultShouldReturnError() {
	assert.Error(s.T(), s.resultError, "Error must not be nil when creating empty pipeline")
}

/*
	Building a pipeline without input readers should return nil and error
*/
type BuildingPipelineWithoutInputShouldReturnError struct {
	BasePipelineSuite
}

func (s *BuildingPipelineWithoutInputShouldReturnError) SetupTest() {

	s.mockFunctions = new(MockedTestFunctionsContainer)
	s.mockFunctions.On("mockFunctionA", MockEmptyObject{}).Return(MockEmptyObject{})

	s.sut = new(PipelineBuilder)
	s.result, s.resultError = s.sut.CreateNew().ThenRunning(s.mockFunctions.mockFunctionA, expectedParallelismDegree).Build()
}

func (s *BuildingPipelineWithoutInputShouldReturnError) TestBuildShouldReturnANilPipeline() {
	assert.Nil(s.T(), s.result, "Returning Pipeline should not be nil")
}

func (s *BuildingPipelineWithoutInputShouldReturnError) TestResultShouldReturnError() {
	assert.Error(s.T(), s.resultError, "Error must not be nil when trying to create a pipeline without input")
}

/*
	Building a pipeline without output appenders should return nil and error
*/
type BuildingPipelineWithoutOutputShouldReturnError struct {
	BasePipelineSuite
}

func (s *BuildingPipelineWithoutOutputShouldReturnError) SetupTest() {

	s.mockFunctions = new(MockedTestFunctionsContainer)
	s.mockFunctions.On("mockFunctionA", MockEmptyObject{}).Return(MockEmptyObject{})

	s.sut = new(PipelineBuilder)
	s.result, s.resultError = s.sut.CreateNew().ThenRunning(s.mockFunctions.mockFunctionA, expectedParallelismDegree).Build()
}

func (s *BuildingPipelineWithoutOutputShouldReturnError) TestBuildShouldReturnANilPipeline() {
	assert.Nil(s.T(), s.result, "Returning Pipeline should not be nil")
}

func (s *BuildingPipelineWithoutOutputShouldReturnError) TestResultShouldReturnError() {
	assert.Error(s.T(), s.resultError, "Error must not be nil when trying to create a pipeline without output appender")
}

/*
	Building a non empty Pipeline test suite.
*/
type BuildingNonEmptyPipeline struct {
	BasePipelineSuite
}

func (s *BuildingNonEmptyPipeline) SetupTest() {

	s.mockFunctions = new(MockedTestFunctionsContainer)

	s.mockFunctions.On("mockFunctionA", MockEmptyObject{}).Return(MockEmptyObject{})
	s.mockFunctions.On("mockFunctionB", MockEmptyObject{}).Return(MockEmptyObject{})

	s.sut = new(PipelineBuilder)

	s.result, s.resultError = s.sut.CreateNew().
		ConsumingFrom(mockedInput).
		ThenRunning(s.mockFunctions.mockFunctionB, 7).
		ThenRunning(s.mockFunctions.mockFunctionA, 13).
		OutputtingTo(mockedOutput).
		Build()
}

func (s *BuildingNonEmptyPipeline) TestBuildShouldReturnAValidPipeline() {
	assert.NotNil(s.T(), s.result, "Returning Pipeline should not be null")
}

//
func (s *BuildingNonEmptyPipeline) TestReturnsNoError() {
	assert.NoError(s.T(), s.resultError)
}

func (s *BuildingNonEmptyPipeline) TestPipelineStepsShouldNotBeEmpty() {
	assert.NotEmpty(s.T(), s.result.steps, "steps should not be empty")
}

func (s *BuildingNonEmptyPipeline) TestPipelinestepsShouldHave2Elements() {
	assert.Equal(s.T(), len(s.result.steps), 2, "steps should have 2 elements")
}

func (s *BuildingNonEmptyPipeline) TestPipelineStepsShouldBeInTheSameOrder() {
	// First callback: Act and Assert
	s.result.steps[0].action(MockEmptyObject{})
	s.mockFunctions.AssertCalled(s.T(), "mockFunctionB", MockEmptyObject{})
	// Second callback: Act and Assert
	s.result.steps[1].action(MockEmptyObject{})
	s.mockFunctions.AssertCalled(s.T(), "mockFunctionA", MockEmptyObject{})
}

func (s *BuildingNonEmptyPipeline) TestParallelismDegreeFromFirstStepShouldBe7() {
	assert.Equal(s.T(), s.result.steps[0].degree, 7, "Degree for first step should be 7")
}

func (s *BuildingNonEmptyPipeline) TestParallelismDegreeFromSecondStepShouldBe13() {
	assert.Equal(s.T(), s.result.steps[1].degree, 13, "Degree for second step should be 13")
}

func (s *BuildingNonEmptyPipeline) TestPipelineStepsInputShouldNotBeNil() {
	for i := 0; i < len(s.result.steps); i++ {
		assert.NotNil(s.T(), s.result.steps[i].input, "Input of of pipeline step should not be Nil")
	}
}

func (s *BuildingNonEmptyPipeline) TestPipelineInputShouldNotBeNil() {
	assert.NotNil(s.T(), s.result.input, "Input should not be Nil")
}

func (s *BuildingNonEmptyPipeline) TestPipelineInputShouldBeMockedInput() {
	assert.True(s.T(), s.result.input == mockedInput, "Input should be mockedInput")
}

func (s *BuildingNonEmptyPipeline) TestPipelineOutputShouldNotBeNil() {
	assert.NotNil(s.T(), s.result.output, "Output should not be Nil")
}

func (s *BuildingNonEmptyPipeline) TestPipelineOutputShouldBeMockedOutput() {
	assert.True(s.T(), s.result.output == mockedOutput, "Input should be mockedOutput")
}

func (s *BuildingNonEmptyPipeline) TestPipelineStepsShouldBeCorrectlyChained() {
	for i := 0; i < len(s.result.steps); i++ {
		// Checking inputs
		if i == 0 {
			assert.True(s.T(), s.result.input == s.result.steps[i].input, "Input not assigned correctly to first step")
		}
		// Checking outputs
		if i == (len(s.result.steps) - 1) {
			assert.True(s.T(), s.result.output == s.result.steps[len(s.result.steps)-1].output, "Output not assigned correctly to last step")
		} else {
			assert.True(s.T(), s.result.steps[i].output == s.result.steps[i+1].input, "Output of step "+strconv.Itoa(i)+" not assigned to input of step "+strconv.Itoa(i+1))
		}
	}
}

type BuildingDiscardOutputPipeline struct {
	BasePipelineSuite
}

func (s *BuildingDiscardOutputPipeline) SetupTest() {

	s.mockFunctions = new(MockedTestFunctionsContainer)

	s.mockFunctions.On("mockFunctionA", MockEmptyObject{}).Return(MockEmptyObject{})
	s.mockFunctions.On("mockFunctionB", MockEmptyObject{}).Return(MockEmptyObject{})

	s.sut = new(PipelineBuilder)

	s.result, s.resultError = s.sut.CreateNew().
		ConsumingFrom(mockedInput).
		ThenRunning(s.mockFunctions.mockFunctionB, 7).
		ThenRunning(s.mockFunctions.mockFunctionA, 13).
		DiscardOutput().
		Build()
}

func (s *BuildingDiscardOutputPipeline) TestBuildShouldReturnAValidPipeline() {
	assert.NotNil(s.T(), s.result, "Returning Pipeline should not be null")
}

func (s *BuildingDiscardOutputPipeline) TestReturnsNoError() {
	assert.NoError(s.T(), s.resultError)
}

func (s *BuildingDiscardOutputPipeline) TestPipelineOutputShouldBeNil() {
	assert.Nil(s.T(), s.result.output, "Output should be Nil")
}

/*
	Runs all the test suites
*/

func TestRunAllSuites(t *testing.T) {
	suite.Run(t, new(MissingStepsBuildingPipelineShouldReturnError))
	suite.Run(t, new(BuildingEmptyPipelineShouldReturnError))
	suite.Run(t, new(BuildingPipelineWithoutInputShouldReturnError))
	suite.Run(t, new(BuildingPipelineWithoutOutputShouldReturnError))
	suite.Run(t, new(BuildingNonEmptyPipeline))
	suite.Run(t, new(BuildingTwiceShouldGenerateDifferentPipelines))
	suite.Run(t, new(BuildingDiscardOutputPipeline))
}
