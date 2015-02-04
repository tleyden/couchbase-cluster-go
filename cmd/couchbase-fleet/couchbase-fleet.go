package main

import (
	"log"

	"github.com/coreos/go-etcd/etcd"
	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Couchbase-Fleet.

Usage:
  couchbase-fleet launch-cbs --version=<cb-version> --num-nodes=<num_nodes> --userpass=<user:pass> [--etcd-servers=<server-list>] 
  couchbase-fleet -h | --help

Options:
  -h --help     Show this screen.
  --version=<cb-version> Couchbase Server version (3.0.1 or 2.2) 
  --num-nodes=<num_nodes> number of couchbase nodes to start
  --userpass <user:pass> the username and password as a single string, delimited by a colon (:)
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
`
	arguments, _ := docopt.Parse(usage, nil, true, "Couchbase-Fleet", false)
	log.Printf("args: %v", arguments)

	if cbcluster.IsCommandEnabled(arguments, "launch-cbs") {
		if err := launchCouchbaseServer(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

}

func launchCouchbaseServer(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)

	couchbaseFleet := NewCouchbaseFleet(etcdServers)
	if err := couchbaseFleet.ExtractDocOptArgs(arguments); err != nil {
		return err
	}

	return couchbaseFleet.LaunchCouchbaseServer()

}

type CouchbaseFleet struct {
	etcdClient  *etcd.Client
	UserPass    string
	NumNodes    int
	CbVersion   string
	EtcdServers []string
}

func NewCouchbaseFleet(etcdServers []string) *CouchbaseFleet {

	c := &CouchbaseFleet{}

	if len(etcdServers) > 0 {
		c.EtcdServers = etcdServers
		log.Printf("Connect to explict etcd servers: %v", c.EtcdServers)
	} else {
		c.EtcdServers = []string{}
		log.Printf("Connect to etcd on localhost")
	}
	c.ConnectToEtcd()
	return c

}

func (c *CouchbaseFleet) ConnectToEtcd() {

	c.etcdClient = etcd.NewClient(c.EtcdServers)
	c.etcdClient.SetConsistency(etcd.STRONG_CONSISTENCY)
}

func (c *CouchbaseFleet) ExtractDocOptArgs(arguments map[string]interface{}) error {

	userpass, err := cbcluster.ExtractUserPass(arguments)
	if err != nil {
		return err
	}
	numnodes, err := cbcluster.ExtractNumNodes(arguments)
	if err != nil {
		return err
	}
	cbVersion, err := cbcluster.ExtractCbVersion(arguments)
	if err != nil {
		return err
	}

	c.UserPass = userpass
	c.NumNodes = numnodes
	c.CbVersion = cbVersion

	return nil
}

func (c *CouchbaseFleet) LaunchCouchbaseServer() error {

	// call fleetctl list-machines and verify that the number of nodes
	// the use asked to kick off is LTE number of machines on cluster

	// create an etcd client

	// this need to check:
	//   no etcd key for /couchbase.com
	//   what else?
	if err := c.verifyCleanSlate(); err != nil {
		return err
	}

	if err := c.setUserNamePassEtcd(); err != nil {
		return err
	}

	// run fleet template through templating engine, passing couchbase version,
	// and save files to temp directory

	// call fleetctl submit on the fleet files
	// 	   fleetctl submit couchbase_node@.service

	// call fleetctl start to start the N servers
	// 	   fleetctl start "couchbase_node@$i.service"

	// wait until X nodes are up in cluster

	// let user know its up

	return nil

}

func (c CouchbaseFleet) verifyCleanSlate() error {
	return nil
}

func (c CouchbaseFleet) setUserNamePassEtcd() error {
	return nil
}
