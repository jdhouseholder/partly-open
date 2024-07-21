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

Insight into understanding parameters
---
A way to understand the parameters is to think about the extremes, where the partly-open system resembles a closed system or an open system.

With `StayProbability=1`, the load resembles a closed system where we have a fixed set of users (Max Workers) making requests indefinitely.
The clients in the closed system make a new request after the completion of their previous request, thus request completion causes a new request (after some think time).
We can ignore `MeanNewWorkersPerSecond`, as we will hit steady state and have `MaxWorkers` forever.

With `StayProbability=0`, the load resembles an open system where users arrive according to some arrival process, think website traffic.
Request completion is independent of a new request.
We will make exactly `MaxWorkers` requests. 

Succinctly, as we increase the `StayProbability` we increase the average client session length.
