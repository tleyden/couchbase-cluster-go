package main

import (
	"log"

	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	couchbaseCluster := &cbcluster.CouchbaseCluster{}
	log.Printf("couchbase cluster: %v", couchbaseCluster)

}
