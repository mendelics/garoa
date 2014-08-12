garoa
========

`garoa` provides an extensible way to plug different functions into several steps.

Each step has its own parallellism degree, and the communication is done through channels.

The only thing necessary for those functions is to accept an `interface{}` (aka anything) and return also `interface{}` with an error. This can encapsulate more complex domain operations.

Garoa means light rain in Portuguese. The software was inspired by Storm (distributed computaional framework).
Conceived by Igor Correa & Vitor De Mario at Mendelics.
