# 🚀 Purpose
## 🚣 Why are we building this ?
We want to build a simple scheduler that scales without much effort. There are a few cron executors in the market, and task schedulers. But they either have a single point of failure, or not flexible. With this scheduler, we want to support jobs like:
* Payment timeouts
* Game engine timeouts
* User wise bonus, ticket, promo expiry
* User engagement timeouts - User adds something to cart but doesnt checkout.

## 🚜 What does it take ?
To support all the functionalities mentioned above, we need
* Fault tolerance
* Scalability
* Accuracy
* Trigger exactly once guarantee
* Easy and dev friendly API
* Simplicity

## 🤹🏽‍♂️ What this is not ?
While there are a very wide range of applications and usecases for schedulers out there, we want to limit our projects' purpose. The last things we want to timeMachine to turn out to are
* CRON expression based job scheduling
* Messaging queue
* Analytical database

Although the scope of this project is limited in the MVP phase, we will definetly consider adding more features that do not go against the core design principles later on