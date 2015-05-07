package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Sync-Gw-Cluster.

Usage:
  sync-gw-cluster launch-sgw --num-nodes=<num_nodes> --config-url=<config_url> [--in-memory-db] [--launch-nginx] [--create-bucket=<bucket-name>] [--create-bucket-size=<bucket-size-mb>] [--create-bucket-replicas=<replica-count>] [--etcd-servers=<server-list>] [--docker-tag=<dt>]
  sync-gw-cluster launch-sidekick --local-ip=<ip> [--etcd-servers=<server-list>]
  sync-gw-cluster -h | --help

Options:
  -h --help     Show this screen.
  --num-nodes=<num_nodes> number of sync gw nodes to start
  --config-url=<config_url> the url where the sync gw config json is stored
  --create-bucket=<bucket-name> create a bucket on couchbase server with the given name 
  --create-bucket-size=<bucket-size-mb> if creating a bucket, use this size in MB
  --create-bucket-replicas=<replica-count> if creating a bucket, use this replica count (defaults to 1)
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
  --docker-tag=<docker-tag>  if present, use this docker tag for spawned containers, otherwise, default to "latest"
  --local-ip=<ip> the ip address (no port) to publish in etcd
`

	arguments, err := docopt.Parse(usage, nil, true, "Sync-Gw-Cluster", false)
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	if cbcluster.IsCommandEnabled(arguments, "launch-sgw") {
		if err := launchSyncGateway(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "launch-sidekick") {
		if err := launchSyncGatewaySidekick(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	log.Printf("Nothing to do!")

}

func launchSyncGateway(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	syncGwCluster := cbcluster.NewSyncGwCluster(etcdServers)
	if err := syncGwCluster.ExtractDocOptArgs(arguments); err != nil {
		return err
	}

	return syncGwCluster.LaunchSyncGateway()

}

func launchSyncGatewaySidekick(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	syncGwCluster := cbcluster.NewSyncGwCluster(etcdServers)

	localIp, _ := cbcluster.ExtractStringArg(arguments, "--local-ip")
	if localIp != "" {
		syncGwCluster.LocalIp = localIp
	}

	return syncGwCluster.LaunchSyncGatewaySidekick()

}
