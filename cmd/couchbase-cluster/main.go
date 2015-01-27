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
  couchbase-cluster -h | --help

Options:
  -h --help     Show this screen.
  --etcd-servers=<server-list>  Comma separated list of etcd servers [default: 127.0.0.1:4001]`

	arguments, _ := docopt.Parse(usage, nil, true, "Couchbase-Cluster", false)

	_, waitUntilRunning := arguments["wait-until-running"]

	etcdServers := extractEtcdServerList(arguments)

	if waitUntilRunning {

		couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)

		numRetries := 10000
		if err := couchbaseCluster.WaitUntilClusterRunning(numRetries); err != nil {
			log.Fatalf("Failed to wait until cluster running: %v", err)
		}

	}

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
