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
  couchbase-cluster start-couchbase-node --local-ip=<ip> [--etcd-servers=<server-list>] 
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

	if cbcluster.IsCommandEnabled(arguments, "start-couchbase-node") {

		localIp, found := arguments["--local-ip"]
		if !found {
			log.Fatalf("Required argument missing")
		}
		localIpString := localIp.(string)
		startCouchbaseNode(etcdServers, localIpString)
		return
	}

}

func startCouchbaseNode(etcdServers []string, localIp string) {

	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)
	couchbaseCluster.LocalCouchbaseIp = localIp

	if err := couchbaseCluster.LoadAdminCredsFromEtcd(); err != nil {
		log.Fatalf("Failed to get admin credentials from etc: %v", err)
	}

	if err := couchbaseCluster.StartCouchbaseNode(); err != nil {
		log.Fatal(err)
	}

}
