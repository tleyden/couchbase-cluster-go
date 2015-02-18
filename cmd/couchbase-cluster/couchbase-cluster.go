package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Couchbase-Cluster.

Usage:
  couchbase-cluster wait-until-running [--etcd-servers=<server-list>] 
  couchbase-cluster start-couchbase-sidekick --local-ip=<ip> [--etcd-servers=<server-list>] 
  couchbase-cluster failover --local-ip=<ip> [--etcd-servers=<server-list>] 
  couchbase-cluster remove-and-rebalance --local-ip=<ip> [--etcd-servers=<server-list>] 
  couchbase-cluster -h | --help

Options:
  -h --help     Show this screen.
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost`

	arguments, _ := docopt.Parse(usage, nil, true, "Couchbase-Cluster", false)
	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	if cbcluster.IsCommandEnabled(arguments, "wait-until-running") {
		cbcluster.WaitUntilCBClusterRunning(etcdServers)
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "start-couchbase-sidekick") {

		localIp, found := arguments["--local-ip"]
		if !found {
			log.Fatalf("Required argument missing")
		}
		localIpString := localIp.(string)
		startCouchbaseSidekick(etcdServers, localIpString)
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "failover") {

		localIp, found := arguments["--local-ip"]
		if !found {
			log.Fatalf("Required argument missing")
		}
		localIpString := localIp.(string)
		failover(etcdServers, localIpString)
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "remove-and-rebalance") {

		localIp, found := arguments["--local-ip"]
		if !found {
			log.Fatalf("Required argument missing")
		}
		localIpString := localIp.(string)
		removeAndRebalance(etcdServers, localIpString)
		return
	}

	log.Fatalf("Nothing to do!")

}

func initCluster(etcdServers []string, localIp string) *cbcluster.CouchbaseCluster {

	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)
	couchbaseCluster.LocalCouchbaseIp = localIp

	if err := couchbaseCluster.LoadAdminCredsFromEtcd(); err != nil {
		log.Fatalf("Failed to get admin credentials from etc: %v", err)
	}

	return couchbaseCluster

}

func startCouchbaseSidekick(etcdServers []string, localIp string) {

	couchbaseCluster := initCluster(etcdServers, localIp)

	if err := couchbaseCluster.StartCouchbaseSidekick(); err != nil {
		log.Fatal(err)
	}

}

func failover(etcdServers []string, localIp string) {

	couchbaseCluster := initCluster(etcdServers, localIp)

	if err := couchbaseCluster.Failover(); err != nil {
		log.Fatal(err)
	}

}

func removeAndRebalance(etcdServers []string, localIp string) {

	couchbaseCluster := initCluster(etcdServers, localIp)

	if err := couchbaseCluster.RemoveAndRebalance(); err != nil {
		log.Fatal(err)
	}

}
