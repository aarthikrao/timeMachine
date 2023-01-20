# Time Machine DB 🐓
[![Slack](https://img.shields.io/badge/Slack-4A154B?style=for-the-badge&logo=slack&logoColor=white)](https://join.slack.com/t/timemachinedb/shared_invite/zt-1nnti899g-6XppaC~5kqF0QAqALBgxqw) 
![Status](https://img.shields.io/badge/Status-Ideation-ffb3ff?style=for-the-badge)

A distributed, fault tolerant scheduler database that can potentially scale to millions of jobs. 

The idea is to build it with a storage layer based on B+tree or LSM-tree implementation, consistent hashing for load balancing, and raft for consensus.

## 🧬 Documentation
- [Purpose](./docs/Purpose.md)
- [Architecture](./docs/Architecture.md) • [Components of a node](/components/Components.md) • [Also read](./docs/Refer.md)
- [Developer APIs](./docs/DevAPI.md) • [Job APIs](./docs/DevAPI.md#-job-apis) • [Route APIs](./docs/DevAPI.md#-route-apis)
- [TODO](./docs/TODO.md)

![Cluster animation](/docs/images/cluster_animation.gif)

## 🎯 Quick start

```bash
# Build 
❯ go build

# Clean and create 5 data folders
❯ ./scripts/clean-create.sh 5

# Spawn 5 instances
❯ ./scripts/spawn.sh 5 true

# Create a cluster
❯ ./scripts/join.sh 5

# Check status
❯ ./scripts/status.sh 5
```
Checkout the [detailed guide](/docs/Setup.md)

## 🎬 Roadmap
You can find the [roadmap here](/docs/Roadmap.md)

## 🛺 Tech Stack
Time machine is built on 
* [BBoltDB](https://github.com/etcd-io/bbolt)
* [Raft](https://raft.github.io/)
* [Consistent hashing](https://en.wikipedia.org/wiki/Consistent_hashing)

For more details checkout our [Tech stack](/docs/Refer.md#🛺-tech-stack)

## ⚽ Contribute
* Choose a component to work on.
* Research the component thoroughly.
* Reach out to me, so that I can mark it as "Work in Progress" to prevent duplication of efforts.
* Build, code, and test the component.
* Submit a pull request (PR) when you are ready to have your changes reviewed.


Refer [Contributing](./CONTRIBUTING.md) for more