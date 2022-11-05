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
