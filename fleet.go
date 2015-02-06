package cbcluster

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

const (
	FLEET_API_ENDPOINT = "http://localhost:49153/v1-alpha"
)

type CouchbaseFleet struct {
	etcdClient          *etcd.Client
	UserPass            string
	NumNodes            int
	CbVersion           string
	EtcdServers         []string
	SkipCleanSlateCheck bool
}

// this is used in the fleet template.
// TODO: should use anon struct
type FleetParams struct {
	CB_VERSION string
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

	for i := 1; i < c.NumNodes+1; i++ {

		fleetUnitJson, err := c.generateFleetUnitJson()
		if err != nil {
			return err
		}
		if err := submitAndLaunchFleetUnitN(i, fleetUnitJson); err != nil {
			return err
		}

	}

	// wait until X nodes are up in cluster

	// let user know its up

	return nil

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
	c.SkipCleanSlateCheck = ExtractSkipCheckCleanState(arguments)

	return nil
}

// call fleetctl list-machines and verify that the number of nodes
// the user asked to kick off is LTE number of machines on cluster
func (c CouchbaseFleet) verifyEnoughMachinesAvailable() error {

	log.Printf("verifyEnoughMachinesAvailable()")

	endpointUrl := fmt.Sprintf("%v/machines", FLEET_API_ENDPOINT)

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

	if c.SkipCleanSlateCheck {
		return nil
	}

	key := path.Join(KEY_NODE_STATE)

	_, err := c.etcdClient.Get(key, false, false)

	// if that key exists, there is residue and we should abort
	if err == nil {
		return fmt.Errorf("Found residue -- key: %v in etcd.  Destroy cluster first", KEY_NODE_STATE)
	}

	// if we get an error with "key not found", then we are starting
	// with a clean slate
	if strings.Contains(err.Error(), "Key not found") {
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

func (c CouchbaseFleet) generateFleetUnitJson() (string, error) {
	// run fleet template through templating engine, passing couchbase version,
	// and save files to temp directory
	fleetUnitJsonTemplate := `
{
    "desiredState":"launched",
    "options":[
        {
            "section":"Service",
            "name":"TimeoutStartSec",
            "value":"0"
        },
        {
            "section":"Service",
            "name":"EnvironmentFile",
            "value":"/etc/environment"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"-/usr/bin/docker kill couchbase"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"-/usr/bin/docker rm couchbase"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"/usr/bin/docker pull tleyden5iwx/couchbase-server-{{ .CB_VERSION }}"
        },
        {
            "section":"Service",
            "name":"ExecStart",
            "value":"/bin/bash -c '/usr/bin/docker run --name couchbase -v /opt/couchbase/var:/opt/couchbase/var --net=host tleyden5iwx/couchbase-server-{{ .CB_VERSION }} couchbase-cluster start-couchbase-node --local-ip=$COREOS_PRIVATE_IPV4'"
        },
        {
            "section":"Service",
            "name":"ExecStop",
            "value":"/usr/bin/docker stop couchbase"
        },
        {
            "section":"X-Fleet",
            "name":"Conflicts",
            "value":"couchbase_node*.service"
        }
    ]
}
`

	tmpl, err := template.New("couchbase_fleet").Parse(fleetUnitJsonTemplate)
	if err != nil {
		return "", err
	}

	params := FleetParams{
		CB_VERSION: c.CbVersion,
	}

	out := &bytes.Buffer{}

	// execute template and write to dest
	err = tmpl.Execute(out, params)
	if err != nil {
		return "", err
	}

	return out.String(), nil

}

func submitAndLaunchFleetUnitN(unitNumber int, fleetUnitJson string) error {

	client := &http.Client{}

	endpointUrl := fmt.Sprintf("%v/units/couchbase_node@%v.service", FLEET_API_ENDPOINT, unitNumber)

	req, err := http.NewRequest("PUT", endpointUrl, bytes.NewReader([]byte(fleetUnitJson)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("response body: %v", string(bodyStr))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Unexpected status code in response")
	}

	return nil

}
