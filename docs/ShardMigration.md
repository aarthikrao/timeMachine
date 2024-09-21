# Shard migration
Shard migration is essential for scaling the cluster up or down. It involves distributing the shards among new nodes. This technique can also be used to consolidate shards into fewer physical nodes to scale down during periods of low traffic. 

Since we prioritize consistency over availability,
* A shard will be unavailable for writes during the migration
* All the jobs recieved during this time will not be added into the database. Instead, it will be added to a seperate queue
* The job is added to the queue only if it lies after the acceptable time duration specified in the config (TODO). This is done to ensure that we maintain a consistent view of the database snapshot during migraion. 
* Once the migraion is complete, the jobs in the queue will be replayed and stored in the actual database
* There may be times when the timeMachine service returns an error response and does not accept a job. However, there will never be a situation where a job is accepted but never triggered, or triggered more than once.

### Limitations
* Migration should only be performed during periods of low traffic. This is because a larger amount of delta data during migration complicates the transfer process.
