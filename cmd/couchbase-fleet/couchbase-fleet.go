package main

import (
	"log"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Couchbase-Fleet.

Usage:
  couchbase-fleet launch-cbs --version=<cb-version> --num-nodes=<num_nodes> --userpass=<user:pass> [--etcd-servers=<server-list>] [--docker-tag=<dt>] [--skip-clean-slate-check]
  couchbase-fleet stop [--etcd-servers=<server-list>]
  couchbase-fleet destroy [--etcd-servers=<server-list>]
  couchbase-fleet generate-units --version=<cb-version> --num-nodes=<num_nodes> --userpass=<user:pass> [--etcd-servers=<server-list>] [--docker-tag=<dt>] --output-dir=<output_dir>
  couchbase-fleet -h | --help

Options:
  -h --help     Show this screen.
  --version=<cb-version> Couchbase Server version (3.0.1 or 2.2) 
  --num-nodes=<num_nodes> number of couchbase nodes to start
  --userpass <user:pass> the username and password as a single string, delimited by a colon (:)
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
  --docker-tag=<dt>  if present, use this docker tag for spawned containers, otherwise, default to "latest"
  --skip-clean-slate-check  if present, will skip the check that we are starting from clean state
  --output-dir=<output_dir>

`

	arguments, err := docopt.Parse(usage, nil, true, "Couchbase-Fleet", false)
	if err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	if cbcluster.IsCommandEnabled(arguments, "launch-cbs") {
		if err := launchCouchbaseServer(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "generate-units") {
		if err := generateUnits(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "stop") {
		if err := stopUnits(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	if cbcluster.IsCommandEnabled(arguments, "destroy") {
		if err := destroyUnits(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	log.Printf("Nothing to do!")

}

func launchCouchbaseServer(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	couchbaseFleet := cbcluster.NewCouchbaseFleet(etcdServers)
	if err := couchbaseFleet.ExtractDocOptArgs(arguments); err != nil {
		return err
	}

	return couchbaseFleet.LaunchCouchbaseServer()

}

func generateUnits(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	couchbaseFleet := cbcluster.NewCouchbaseFleet(etcdServers)
	if err := couchbaseFleet.ExtractDocOptArgs(arguments); err != nil {
		return err
	}

	// get the output dir from args
	outputDir, err := cbcluster.ExtractStringArg(arguments, "--output-dir")
	if err != nil {
		return err
	}

	if err := couchbaseFleet.GenerateUnits(outputDir); err != nil {
		return err
	}

	log.Printf("Unit files written to %v", outputDir)

	return nil

}

func stopUnits(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	couchbaseFleet := cbcluster.NewCouchbaseFleet(etcdServers)

	return couchbaseFleet.StopUnits()

}

func destroyUnits(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	couchbaseFleet := cbcluster.NewCouchbaseFleet(etcdServers)

	return couchbaseFleet.DestroyUnits()

}
