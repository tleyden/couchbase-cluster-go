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
	"time"

	"github.com/tleyden/go-etcd/etcd"
)

const (
	FLEET_API_ENDPOINT = "http://localhost:49153/v1-alpha"
)

type CouchbaseFleet struct {
	etcdClient          *etcd.Client
	UserPass            string
	NumNodes            int
	CbVersion           string
	ContainerTag        string // Docker tag
	EtcdServers         []string
	SkipCleanSlateCheck bool
}

// this is used in the fleet template.
// TODO: should use anon struct
type FleetParams struct {
	CB_VERSION    string
	CONTAINER_TAG string
}
type SidekickFleetParams struct {
	CONTAINER_TAG string
	UNIT_NUMBER   int
}

func NewCouchbaseFleet(etcdServers []string) *CouchbaseFleet {

	c := &CouchbaseFleet{}

	if len(etcdServers) > 0 {
		c.EtcdServers = etcdServers
		log.Printf("Connect to explicit etcd servers: %v", c.EtcdServers)
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

	nodeFleetUnitJson, err := c.generateNodeFleetUnitJson()
	if err != nil {
		return err
	}
	for i := 1; i < c.NumNodes+1; i++ {

		if err := submitAndLaunchFleetUnitN(
			i,
			"couchbase_node",
			nodeFleetUnitJson,
		); err != nil {
			return err
		}

		sidekickFleetUnitJson, err := c.generateSidekickFleetUnitJson(i)
		if err != nil {
			return err
		}

		if err := submitAndLaunchFleetUnitN(
			i,
			"couchbase_sidekick",
			sidekickFleetUnitJson,
		); err != nil {
			return err
		}

	}

	// wait until X nodes are up in cluster
	log.Printf("Waiting for cluster to be up ..")
	WaitUntilNumNodesRunning(c.NumNodes, c.EtcdServers)

	// wait until no rebalance running
	cb := NewCouchbaseCluster(c.EtcdServers)

	if err := cb.LoadAdminCredsFromEtcd(); err != nil {
		return err
	}
	liveNodeIp, err := cb.FindLiveNode()
	if err != nil {
		return err
	}

	// dirty hack to solve problem: the cluster might have
	// 2 nodes which just finished rebalancing, and a third node
	// that joins and triggers another rebalance.  thus, it will briefly
	// go into "no rebalances happening" state, followed by a rebalance.
	// if we see the "no rebalances happening state", we'll be tricked and
	// think we're done when we're really not.
	// workaround: check twice, and sleep in between the check
	for i := 0; i < c.NumNodes; i++ {
		if err := cb.WaitUntilNoRebalanceRunning(liveNodeIp, 30); err != nil {
			return err
		}
		log.Printf("No rebalance running, sleeping 15s. (%v/%v)", i+1, c.NumNodes)
		<-time.After(time.Second * 15)

	}
	log.Println("No rebalance running after several checks")

	// let user know its up

	log.Printf("Cluster is up!")

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
	c.ContainerTag = ExtractDockerTagOrLatest(arguments)
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
		return fmt.Errorf("Found residue -- key: %v in etcd.  You should destroy the cluster first, then try again.", KEY_NODE_STATE)
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

func (c CouchbaseFleet) generateNodeFleetUnitJson() (string, error) {

	fleetUnitJsonTemplate := `
{
    "desiredState":"inactive",
    "options":[
        {
            "section":"Unit",
            "name":"Description",
            "value":"couchbase_node"
        },
        {
            "section":"Unit",
            "name":"After",
            "value":"docker.service"
        },
        {
            "section":"Unit",
            "name":"Requires",
            "value":"docker.service"
        },
        {
            "section":"Unit",
            "name":"After",
            "value":"etcd.service"
        },
        {
            "section":"Unit",
            "name":"Requires",
            "value":"etcd.service"
        },
        {
            "section":"Service",
            "name":"TimeoutStartSec",
            "value":"0"
        },
        {
            "section":"Service",
            "name":"TimeoutStopSec",
            "value":"600"
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
            "value":"/usr/bin/docker pull tleyden5iwx/couchbase-server-{{ .CB_VERSION }}:{{ .CONTAINER_TAG }}"
        },
        {
            "section":"Service",
            "name":"ExecStart",
            "value":"/bin/bash -c '/usr/bin/docker run --name couchbase -v /opt/couchbase/var:/opt/couchbase/var --net=host tleyden5iwx/couchbase-server-{{ .CB_VERSION }}:{{ .CONTAINER_TAG }} couchbase-start'"
        },
        {
            "section":"Service",
            "name":"ExecStop",
            "value":"/bin/bash -c 'curl localhost:4001/version 2>&1 | tee /home/core/curl-etcd.txt;  echo stopping via remove-and-rebalance | tee /home/core/echo-out.txt; /usr/bin/wget https://drone.io/github.com/tleyden/couchbase-cluster-go/files/cmd/couchbase-cluster/couchbase-cluster; /usr/bin/chmod +x couchbase-cluster; ./couchbase-cluster remove-and-rebalance --local-ip $COREOS_PRIVATE_IPV4 2>&1 | tee /home/core/out2.txt;  echo stopping docker container | tee /home/core/echo-out2.txt; sudo docker ps 2>&1 | tee /home/core/docker-beforestop.txt;  sudo docker stop couchbase; sudo docker ps 2>&1 | tee /home/core/docker-afterstop.txt; echo ExecStop done!'"
        },
        {
            "section":"X-Fleet",
            "name":"Conflicts",
            "value":"couchbase_node*.service"
        }
    ]
}
`

	tmpl, err := template.New("couchbase_node").Parse(fleetUnitJsonTemplate)
	if err != nil {
		return "", err
	}

	params := FleetParams{
		CB_VERSION:    c.CbVersion,
		CONTAINER_TAG: c.ContainerTag,
	}

	out := &bytes.Buffer{}

	// execute template and write to dest
	err = tmpl.Execute(out, params)
	if err != nil {
		return "", err
	}

	log.Printf("Couchbase node fleet unit: %v", out.String())

	return out.String(), nil

}

func (c CouchbaseFleet) generateSidekickFleetUnitJson(unitNumber int) (string, error) {

	fleetUnitJsonTemplate := `
{
    "desiredState":"inactive",
    "options":[
        {
            "section":"Unit",
            "name":"Description",
            "value":"couchbase_sidekick"
        },
        {
            "section":"Unit",
            "name":"After",
            "value":"docker.service"
        },
        {
            "section":"Unit",
            "name":"Requires",
            "value":"docker.service"
        },
        {
            "section":"Unit",
            "name":"BindsTo",
            "value":"couchbase_node@{{ .UNIT_NUMBER }}.service"
        },
        {
            "section":"Unit",
            "name":"After",
            "value":"couchbase_node@{{ .UNIT_NUMBER }}.service"
        },
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
            "value":"-/usr/bin/docker kill couchbase-sidekick"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"-/usr/bin/docker rm couchbase-sidekick"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"/usr/bin/docker pull tleyden5iwx/couchbase-cluster-go:{{ .CONTAINER_TAG }}"
        },
        {
            "section":"Service",
            "name":"ExecStart",
            "value":"/bin/bash -c '/usr/bin/docker run --name couchbase-sidekick -v /opt/couchbase/var:/opt/couchbase/var --net=host tleyden5iwx/couchbase-cluster-go:{{ .CONTAINER_TAG }} update-wrapper couchbase-cluster start-couchbase-sidekick --local-ip=$COREOS_PRIVATE_IPV4'"
        },
        {
            "section":"Service",
            "name":"ExecStop",
            "value":"/usr/bin/docker stop couchbase-sidekick"
        },
        {
            "section":"X-Fleet",
            "name":"MachineOf",
            "value":"couchbase_node@{{ .UNIT_NUMBER }}.service"
        }
    ]
}
`

	tmpl, err := template.New("couchbase_fleet").Parse(fleetUnitJsonTemplate)
	if err != nil {
		return "", err
	}

	params := SidekickFleetParams{
		CONTAINER_TAG: c.ContainerTag,
		UNIT_NUMBER:   unitNumber,
	}

	out := &bytes.Buffer{}

	// execute template and write to dest
	err = tmpl.Execute(out, params)
	if err != nil {
		return "", err
	}

	log.Printf("Couchbase sidekick fleet unit %v: %v", unitNumber, out.String())

	return out.String(), nil

}

func submitAndLaunchFleetUnitN(unitNumber int, unitName, fleetUnitJson string) error {

	/* temporarily disable this, since it may be causing issues
	           that make "systemctl cat <unit>" stop working
		if err := submitFleetUnit(unitName, fleetUnitJson); err != nil {
			return err
		}
	*/

	if err := launchFleetUnitN(unitName, unitNumber, fleetUnitJson); err != nil {
		return err
	}

	return nil

}

func submitFleetUnit(unitName, fleetUnitJson string) error {

	log.Printf("Submit fleet unit for %v", unitName)

	endpointUrl := fmt.Sprintf("%v/units/%v@.service", FLEET_API_ENDPOINT, unitName)

	return PUT(endpointUrl, fleetUnitJson)

}

func launchFleetUnitN(unitName string, unitNumber int, fleetUnitJson string) error {

	// this is clunky due to:
	// https://github.com/coreos/fleet/issues/1118

	// modify the json so that the state is "launched" rather than
	// "active"
	fleetUnitJsonModified := strings.Replace(
		fleetUnitJson,
		"inactive",
		"launched",
		-1,
	)

	log.Printf("Launch fleet unit %v (%v)", unitName, unitNumber)

	endpointUrl := fmt.Sprintf("%v/units/%v@%v.service", FLEET_API_ENDPOINT, unitName, unitNumber)

	return PUT(endpointUrl, fleetUnitJsonModified)

}

func PUT(endpointUrl, json string) error {

	client := &http.Client{}

	req, err := http.NewRequest("PUT", endpointUrl, bytes.NewReader([]byte(json)))
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
