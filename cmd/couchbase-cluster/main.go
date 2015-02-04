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
  couchbase-cluster start-couchbase-node --local-ip=<ip>
  couchbase-cluster -h | --help

Options:
  -h --help     Show this screen.
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost`

	arguments, _ := docopt.Parse(usage, nil, true, "Couchbase-Cluster", false)
	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	if cbcluster.IsCommandEnabled(arguments, "wait-until-running") {
		waitUntilRunning(etcdServers)
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

func waitUntilRunning(etcdServers []string) {

	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)

	if err := couchbaseCluster.LoadAdminCredsFromEtcd(); err != nil {
		log.Fatalf("Failed to get admin credentials from etc: %v", err)
	}

	stupidPortHack(couchbaseCluster)

	numRetries := 10000
	if err := couchbaseCluster.WaitUntilClusterRunning(numRetries); err != nil {
		log.Fatalf("Failed to wait until cluster running: %v", err)
	}

}

func startCouchbaseNode(etcdServers []string, localIp string) {

	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)
	couchbaseCluster.LocalCouchbaseIp = localIp

	if err := couchbaseCluster.LoadAdminCredsFromEtcd(); err != nil {
		log.Fatalf("Failed to get admin credentials from etc: %v", err)
	}

	stupidPortHack(couchbaseCluster)

	if err := couchbaseCluster.StartCouchbaseNode(); err != nil {
		log.Fatal(err)
	}

}

func stupidPortHack(cluster *cbcluster.CouchbaseCluster) {

	// stupid hack needed because we aren't storing the live node ports
	// in etcd.  for ecample, in etcd we have:
	//   /couchbase.com/couchbase-node-state/10.153.167.148
	// but we should have:
	//   /couchbase.com/couchbase-node-state/10.153.167.148:8091
	cluster.LocalCouchbasePort = "8091"

}
