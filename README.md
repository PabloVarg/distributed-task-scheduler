# Distributed Task/Job scheduler

Welcome to my Distributed Task Scheduler! This tool helps you efficiently manage and distribute tasks across a cluster of worker nodes. The scheduler supports two primary modes of worker discovery:

- **Heartbeat-based Discovery:** Workers periodically send heartbeats to the scheduler to announce their presence.
- **External Load Balancer:** Use an external load balancer, such as Traefik, to handle worker discovery and routing.

# Running the examples

There are two exaples in this repo, one uses heartbeats for scheduler/worker communication and the other uses an external load balancer like traefik

## Heartbeat example

```
docker compose -f docker-compose.prod.heartbeats.yml up
```

## External load balancer example (Traefik)

```
docker compose -f docker-compose.prod.balancer.yml up
```

# Submitting tasks

To submit a new task you can:

> This will submit a task to be run after 1 second of calling the API, the delay field is parsed my `time.parseDuration` function

```
curl -X POST localhost:8000/tasks -d '{"command": "echo \"hello world\"", "delay": "1s"}'
```

or

> This will submit a task to be run at the exact moment specified in "scheduled_at"

```
curl -X POST localhost:8000/tasks -d '{"command": "echo \"hello world\"", "scheduled_at": "2024-08-18T03:15:52+00:00"}'
```
