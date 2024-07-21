Partly-Open, a simple work load generator
---
To create a `LoadGenerator` you'll bring your own `doWork` implementation and specify the 
* Mean New Workers Per Second
* Jitter
* Max Workers
* Stay Probability
* Think Time

We'll then spawn a total of `MaxWorkers` new workers with mean spawn rate of `MeanNewWorkersPerSecond` with random `Jitter`.
Each worker will `doWork` and then stop with probability `1 - StayProbability`.
When a worker decides to stay it will think for `ThinkTime` and then `doWork` again, this will repeat until the worker has decided to stop.

The `LoadGenerator` will shutdown on first error, and return the error.

Based on "Open Versus Closed: A Cautionary Tale"
<https://www.usenix.org/legacy/event/nsdi06/tech/full_papers/schroeder/schroeder.pdf>.
