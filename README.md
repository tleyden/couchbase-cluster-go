[![Build Status](https://drone.io/github.com/tleyden/couchbase-cluster-go/status.png)](https://drone.io/github.com/tleyden/couchbase-cluster-go/latest)
[![Join the chat at https://gitter.im/tleyden/couchbase-cluster-go](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tleyden/couchbase-cluster-go?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

This is a Go library that helps initialize and manage a Couchbase Server cluster running under CoreOS.

## Requirements

* Couchbase Server
* Etcd present on all nodes (this ships by default with CoreOS)

## Issue Tracker

Please file issues to the [couchbase-server-docker](https://github.com/couchbaselabs/couchbase-server-docker) repo.  

## Rebalance-during-startup

### Problem

Currently what happens is:

* Nodes race to trigger rebalances
* Losers block until no rebalance running, then trigger

The net effect is that in a 3 node cluster coming up, two rebalances happen making it slow.

### Desired - case: inside time window

* 1st node comes up and establishes cluster
* 2nd node comes up, followed by 3rd node in 15 seconds (inside 30 sec time window)
   * 2nd node publishes a rebalance-intent at t0
   * 3rd node joins the rebalance-intent at t1 (+15 seconds later)
   * At t2 (+30 seconds later), 2nd node triggers rebalance
       * Waits until rebalance finishes
       * Publishes node state as active
   * 3rd node 
       * Waits for next rebalance since it expects one coming
       * Waits until next rebalance finishes
       * Publishes node state as active
   
### Desired - case: outside time window

* 1st node comes up and establishes cluster
* 3rd node comes up, followed by 2nd node in 90 seconds (outside 30 sec time window)
   * 3rd node publishes a rebalance-intent at t0
   * At t1 (+30 seconds later), 3rd node triggers rebalance
       * Waits until rebalance finishes
       * Publishes node state as active
* At t2 (+90), 2nd node comes up 
  * Sees no rebalance intent
  * Publishes a rebalance intent that expires at t3 (+120)
* At t3 (+120), rebalance intent expires
  * 2nd node triggers rebalance
  * Waits until rebalance finishes
  * Publishes node state as active

### Rebalance Algorithm

* Add self to cluster
* Is there a pending rebalance?
  * If yes, join it
    * Wait for the rebalance to happen
    * Wait until rebalance finishes
    * Verify cluster membership == active
    * Publish node state as active
  * If no, create one
    * Publish node to etcd: /couchbase.com/pending-rebalance/<local-ip> with ttl=30
    * Wait for 30 seconds
    * Trigger rebalance
    * Wait until rebalance finishes
    * Publish node state as active
    
### Current node publish state

* Only *after* they join the cluster and rebalance, they publish themselves as: ip -> up
* Until then, they are invisible

## References

* [Running a Sync Gateway Cluster Under CoreOS on AWS](http://tleyden.github.io/blog/2014/12/15/running-a-sync-gateway-cluster-under-coreos-on-aws/)
* [couchbase-server-docker repo](https://github.com/couchbaselabs/couchbase-server-docker)	
