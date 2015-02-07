package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Sync-Gw-Cluster.

Usage:
  sync-gw-cluster launch-sgw --num-nodes=<num_nodes> --config-url=<config_url> [--sync-gw-commit=<branch-or-commit>] [--create-bucket=<bucket-name>] [--create-bucket-size=<bucket-size-mb>] 
  sync-gw-cluster -h | --help

Options:
  -h --help     Show this screen.
  --num-nodes=<num_nodes> number of sync gw nodes to start
  --config-url=<config_url> the url where the sync gw config json is stored
  --sync-gw-commit=<branch-or-commit> the branch or commit of sync gw to use, defaults to "image", which is the master branch at the time the docker image was built.
  --create-bucket=<bucket-name> create a bucket on couchbase server with the given name 
  --create-bucket-size=<bucket-size-mb> if creating a bucket, use this size in MB
`

	arguments, err := docopt.Parse(usage, nil, true, "Sync-Gw-Cluster", false)
	log.Printf("args: %v.  err: %v", arguments, err)

	if cbcluster.IsCommandEnabled(arguments, "launch-sgw") {
		if err := launchSyncGateway(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	log.Printf("Nothing to do!")

}

func launchSyncGateway(arguments map[string]interface{}) error {

	// create bucket (if user asked for this)

	// # add values to etcd
	// etcdctl set /couchbase.com/sync-gateway/config "$configFileOrURL"
	// etcdctl set /couchbase.com/sync-gateway/commit "$commit"

	// kick off fleet units

}
