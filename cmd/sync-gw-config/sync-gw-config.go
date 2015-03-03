package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Sync-Gw-Config.

Usage:
  sync-gw-config rewrite --destination=<config-dest> [--etcd-servers=<server-list>]
  sync-gw-config -h | --help

Options:
  -h --help     Show this screen.
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
  --destination=<config-dest> The path where the updated config should be written
`

	arguments, err := docopt.Parse(usage, nil, true, "Sync-Gw-Config", false)
	log.Printf("args: %v.  err: %v", arguments, err)

	if cbcluster.IsCommandEnabled(arguments, "rewrite") {
		if err := retwriteConfig(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	log.Printf("Nothing to do!")

}

func rewriteConfig(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	syncGwCluster := cbcluster.NewSyncGwCluster(etcdServers)

	// get the sync gw config from etcd (cbcluster.KEY_SYNC_GW_CONFIG)
	syncGwConfig, err := syncGwCluster.FetchSyncGwConfig()

	// get a couchbase live node
	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)
	liveNodeIp, err := couchbaseCluster.FindLiveNode()
	if err != nil {
		return err
	}

	// run the sync gw config through go templating engine
	updatedConfig, err := syncGwCluster.UpdateConfig(liveNodeIp, syncGwConfig)
	if err != nil {
		return err
	}

	// write the new config to the dest file

	/*
		syncGwCluster := cbcluster.NewSyncGwCluster(etcdServers)
		if err := syncGwCluster.ExtractDocOptArgs(arguments); err != nil {
			return err
		}

		return syncGwCluster.LaunchSyncGateway()

	*/

}
