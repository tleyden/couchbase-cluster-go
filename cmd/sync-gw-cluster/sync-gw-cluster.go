package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Sync-Gw-Cluster.

Usage:
  sync-gw-cluster launch-sgw --num-nodes=<num_nodes> --config-url=<config_url> [--sync-gw-commit=<branch-or-commit>] [--in-memory-db] [--create-bucket=<bucket-name>] [--create-bucket-size=<bucket-size-mb>] [--create-bucket-replicas=<replica-count>] [--etcd-servers=<server-list>] [--docker-tag=<dt>]
  sync-gw-cluster launch-sidekick --local-ip=<ip> [--etcd-servers=<server-list>]
  sync-gw-cluster -h | --help

Options:
  -h --help     Show this screen.
  --num-nodes=<num_nodes> number of sync gw nodes to start
  --config-url=<config_url> the url where the sync gw config json is stored
  --sync-gw-commit=<branch-or-commit> the branch or commit of sync gw to use, defaults to "image", which is the master branch at the time the docker image was built.
  --in-memory-db add this flag if you don't need to wait for Couchbase Server to launch
  --create-bucket=<bucket-name> create a bucket on couchbase server with the given name 
  --create-bucket-size=<bucket-size-mb> if creating a bucket, use this size in MB
  --create-bucket-replicas=<replica-count> if creating a bucket, use this replica count (defaults to 1)
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
  --docker-tag=<docker-tag>  if present, use this docker tag for spawned containers, otherwise, default to "latest"
  --local-ip=<ip> the ip address (no port) to publish in etcd
`

	arguments, err := docopt.Parse(usage, nil, true, "Sync-Gw-Cluster", false)
	log.Printf("args: %v.  err: %v", arguments, err)

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

	inMemoryDb := cbcluster.ExtractBoolArg(arguments, "--in-memory-db")
	syncGwCluster.RequiresCouchbaseServer = !inMemoryDb

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
