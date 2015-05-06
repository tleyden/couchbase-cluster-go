package main

import (
	"log"

	"github.com/tleyden/couchbase-cluster-go"
)

// func (c CouchbaseCluster) CreateDefaultBucket() error {

func main() {

	couchbaseCluster := cbcluster.NewCouchbaseCluster([]string{})
	couchbaseCluster.LocalCouchbaseIp = "127.0.0.1"
	couchbaseCluster.AdminUsername = "user"
	couchbaseCluster.AdminPassword = "passw0rd"

	if err := couchbaseCluster.CreateDefaultBucket(); err != nil {
		log.Fatalf("Error creating default bucket: %v", err)
	}

}
