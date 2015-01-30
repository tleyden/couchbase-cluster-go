package main

import (
	"log"
	"strings"

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
	etcdServers := extractEtcdServerList(arguments)

	_, cmdWaitUntilRunning := arguments["wait-until-running"]
	if cmdWaitUntilRunning {
		waitUntilRunning(etcdServers)
		return
	}

	_, cmdStartCouchbaseNode := arguments["start-couchbase-node"]
	if cmdStartCouchbaseNode {
		startCouchbaseNode(etcdServers)
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

func startCouchbaseNode(etcdServers []string) {

	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)

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

// convert from comma separated list to a string slice
func extractEtcdServerList(docOptParsed map[string]interface{}) []string {

	rawServerList, found := docOptParsed["--etcd-servers"]
	if !found {
		return nil
	}

	rawServerListStr, ok := rawServerList.(string)
	if !ok {
		return nil
	}

	return strings.Split(rawServerListStr, ",")

}
