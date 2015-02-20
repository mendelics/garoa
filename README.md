# Garoa

`garoa` is a library that provides an extensible way to plug different functions into several steps.

Each step has its own parallellism degree (aka number of workers), and the communication is done internally through channels.

The only thing necessary for those functions is to accept an `interface{}` (aka anything) and return also `interface{}` with an error. This can encapsulate more complex domain operations.

Garoa means light rain in Portuguese. The software was inspired by Apache Storm (distributed computaional framework).

Conceived by Igor Correa & Vitor De Mario at Mendelics.

## Usage example

    convertJSON := func(i interface{}) (interface{}, error) {
        converted, err := json.Marshal(i)
        if err != nil {
            log.Println("json err", err)
            return i, err
        }
        return converted, nil
    }

    accumulator := 0
    count := func(i interface{}) (interface{}, error) {
        accumulator++ // unsafe operation, but there will be only one worker
        if accumulator%1000 == 0 {
            log.Printf("%d reached end of pipeline\n", accumulator)
        }
        return nil, nil // means skip
    }

    input := make(chan interface{})
    pipelineBuilder := new(garoa.PipelineBuilder)
    pipe, err := pipelineBuilder.
	    CreateNew().
	    ConsumingFrom(input).
	    ThenRunning(convertJSON, 10). // 10 workers converting inputs to JSON
	    ThenRunning(count, 1). // only one worker updating the count
	    Build()

	// send as much data as necessary into the input channel

    log.Println("Starting pipeline")
    pipeFinished := pipe.Run()
    <-pipeFinished // wait for the pipe to signal it has finished
    log.Printf("Pipeline finished, processed %d inputs", accumulator)
