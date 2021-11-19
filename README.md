# Delayed Job Operator

Delaying the start of Jobs until after a Unix Timestamp.

This might be useful if you need to do something at a specific time, once and then cleaned up.

Why not use a CronJob with a specific day / time set ?

> CronJobs would achieve the goal of running at a specific time,  
> but still exist in the Kubernetes API.  
> If you list CronJobs, you could have a high noise vs. signal.  
> If you are running it only once, a CronJob might be a better idea.  
> This Operator exists for cases where many delayed jobs need to be created often or regularly.

Why not us a Job with a `sleep` statement.

> This would create a pod running for the specified amount of time, and then run the job,
> which would solve the problem, but any pod rescheduling or restarts will reset the timer,
> and no guarantees can be made about the specific time when a job will be run.

Essentially a Job is missing the scheduling capability, and a CronJob is missing the TTL capability.

What this Operator tries to achieve is to add a scheduling component to a normal Job.

It will simply delay the execution of a Job until a specified time, 
also allowing extending that time by editing the `.spec.after` field.
 
