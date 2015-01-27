package main

import (
	"log"

	"github.com/alecthomas/kingpin"
	"github.com/tleyden/couchbase-cluster-go"
)

var (
	app = kingpin.New(
		"couchbase-cluster",
		"CLI util for managing a Couchbase Server cluster under CoreOS.",
	)

	blockUntilRunning = app.Flag(
		"block-until-running",
		"Block until all couchbase nodes are detected to be healthy",
	).Bool()

	etcdServers = app.Flag(
		"etcd-servers",
		"Comma separated list of etcd servers (ip/port)",
	).Default("127.0.0.1:4001").String()
)

func main() {
	kingpin.Parse()
	log.Printf("blockUntilRunning: %v, etcdServers: %v", *blockUntilRunning, *etcdServers)

	couchbaseCluster := &cbcluster.CouchbaseCluster{}
	log.Printf("couchbase cluster: %v", couchbaseCluster)

}
