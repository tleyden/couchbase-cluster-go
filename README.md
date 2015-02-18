[![Build Status](https://drone.io/github.com/tleyden/couchbase-cluster-go/status.png)](https://drone.io/github.com/tleyden/couchbase-cluster-go/latest)
[![Join the chat at https://gitter.im/tleyden/couchbase-cluster-go](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/tleyden/couchbase-cluster-go?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

This is a Go library that helps initialize and manage a Couchbase Server cluster running under CoreOS.

## Requirements

* Couchbase Server
* Etcd present on all nodes (this ships by default with CoreOS)

## Issue Tracker

Please file issues to the [couchbase-server-docker](https://github.com/couchbaselabs/couchbase-server-docker) repo.  

## Notes regarding node restart

### Approach 1 - remove and re-add

The less risky /slower way:

1. remove - rebalance
1. Reboot the OS
1. add node back
1. rebalance

### Approach 2 - failover

Risky way - fast

1. Graceful failover on the node I need to restart the OS
1. Reboot the OS
1. Delta node recovery
1. Rebalance

## References

* [Running a Sync Gateway Cluster Under CoreOS on AWS](http://tleyden.github.io/blog/2014/12/15/running-a-sync-gateway-cluster-under-coreos-on-aws/)
* [couchbase-server-docker repo](https://github.com/couchbaselabs/couchbase-server-docker)	
