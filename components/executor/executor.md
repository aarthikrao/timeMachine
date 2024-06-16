# Executor
An Executor is a component that manages a queue of tasks (in this case, jobs) with specific timing constraints. It ensures that each task is executed at the right moment in time, based on its trigger time defined by `TriggerMS`

**How does it work?**

The Executor has three main components:

1. **Job Queue**: A data structure that stores tasks (jobs) waiting to be executed. It is implemented with a min heap and a hashmap.
2. **Dispatcher**: A go-routine that continuously fetches jobs from the queue and dispatches them for execution when their trigger time is reached.
3. **Outbound Job Channel**: The triggered jobs are sent via this channel. 

Here's a step-by-step overview of the process:

1. The Executor creates a new job entry in its internal data structure, which includes information about the job (e.g., ID, trigger time, version). The jobs are sorted by the `TriggerMS` in a min heap
2. When the job's trigger time arrives, the Dispatcher(triggered by a `time.Ticker`) fetches that job from the queue and calls the `dispatchJob` function.
3. In `dispatchJob`, the Executor checks if the job is still valid (i.e., not deleted). If so it dispatches the job for execution by sending it to a channel (`outboundJobs`) for further processing.
4. Step 3 is repeated until the dispatcher finds a job that is in the future, returns, and waits for the next tick

**Timing parameters**
The Executor has two timing-related parameters:

1. **Grace period**: Its the time after calling the close function by which the Executor force shuts down. This time is to allow for the already queued jobs to be executed before closing.
2. **Accuracy**: This is the time interval for ticks in the dispatcher. The smaller the accuracy, the more accurate the job will be executed at the actual trigger time. 
