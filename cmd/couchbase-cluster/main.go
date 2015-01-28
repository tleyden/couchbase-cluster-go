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

		if err := couchbaseCluster.LoadAdminCredsFromEtcd(); err != nil {
			log.Fatalf("Failed to get admin credentials from etc: %v", err)
		}

		// stupid hack needed because we aren't storing the live node ports
		// in etcd.  for ecample, in etcd we have:
		//   /couchbase.com/couchbase-node-state/10.153.167.148
		// but we should have:
		//   /couchbase.com/couchbase-node-state/10.153.167.148:8091
		couchbaseCluster.LocalCouchbasePort = 8091

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
