package cbcluster

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

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

	userpass, err := ExtractUserPass(arguments)
	if err != nil {
		return err
	}
	numnodes, err := ExtractNumNodes(arguments)
	if err != nil {
		return err
	}
	cbVersion, err := ExtractCbVersion(arguments)
	if err != nil {
		return err
	}

	c.UserPass = userpass
	c.NumNodes = numnodes
	c.CbVersion = cbVersion

	return nil
}

func (c *CouchbaseFleet) LaunchCouchbaseServer() error {

	if err := c.verifyEnoughMachinesAvailable(); err != nil {
		return err
	}

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

// call fleetctl list-machines and verify that the number of nodes
// the user asked to kick off is LTE number of machines on cluster
func (c CouchbaseFleet) verifyEnoughMachinesAvailable() error {

	log.Printf("verifyEnoughMachinesAvailable()")

	endpointUrl := "http://localhost:49153/v1-alpha/machines"

	// {"machines":[{"id":"a91c394439734375aa256d7da1410132","primaryIP":"172.17.8.101"}]}
	jsonMap := map[string]interface{}{}
	if err := getJsonData(endpointUrl, &jsonMap); err != nil {
		log.Printf("getJsonData error: %v", err)
		return err
	}

	machineListRaw := jsonMap["machines"]
	machineList, ok := machineListRaw.([]interface{})
	if !ok {
		return fmt.Errorf("Unexpected value for machines: %v", jsonMap)
	}

	if len(machineList) < c.NumNodes {
		return fmt.Errorf("User requested %v nodes, only %v available", c.NumNodes, len(machineList))
	}

	log.Printf("/verifyEnoughMachinesAvailable()")

	return nil
}

// Make sure that /couchbase.com/couchbase-node-state is empty
func (c CouchbaseFleet) verifyCleanSlate() error {

	key := path.Join(KEY_NODE_STATE)

	_, err := c.etcdClient.Get(key, false, false)

	// if that key exists, there is residue and we should abort
	if err == nil {
		return fmt.Errorf("Found residue -- key: %v in etcd.  Destroy cluster first", KEY_NODE_STATE)
	}

	// if we get an error with "key not found", then we are starting
	// with a clean slate
	if strings.Contains(err.String(), "Key not found") {
		return nil
	}

	// if we got a different error rather than "Key not found", treat that as
	// an error as well.
	return fmt.Errorf("Unexpected error trying to get key: %v: %v", KEY_NODE_STATE, err)

}

func (c CouchbaseFleet) setUserNamePassEtcd() error {

	_, err := c.etcdClient.Set(KEY_USER_PASS, c.UserPass, 0)

	return err

}
