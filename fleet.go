package cbcluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/go-systemd/unit"
	"github.com/tleyden/go-etcd/etcd"
)

const (
	FLEET_API_ENDPOINT = "http://localhost:49153/fleet/v1"
	UNIT_NAME_NODE     = "couchbase_node"
	UNIT_NAME_SIDEKICK = "couchbase_sidekick"
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
			UNIT_NAME_NODE,
			nodeFleetUnitJson,
		); err != nil {
			return err
		}

		sidekickFleetUnitJson, err := c.generateSidekickFleetUnitJson(fmt.Sprintf("%v", i))
		if err != nil {
			return err
		}

		if err := submitAndLaunchFleetUnitN(
			i,
			UNIT_NAME_SIDEKICK,
			sidekickFleetUnitJson,
		); err != nil {
			return err
		}

	}

	if err := c.WaitForFleetLaunch(); err != nil {
		log.Printf("Error waiting for couchbase cluster launch: %v", err)
		return err
	}

	return nil

}

func (c CouchbaseFleet) GenerateUnits(outputDir string) error {

	// generate node unit
	nodeFleetUnit, err := c.generateNodeFleetUnitFile()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%v@.service", UNIT_NAME_NODE)
	path := filepath.Join(outputDir, filename)

	if err := ioutil.WriteFile(path, []byte(nodeFleetUnit), 0644); err != nil {
		return err
	}

	// generate sidekick unit
	sidekickFleetUnit, err := c.generateSidekickFleetUnitFile("%i")
	if err != nil {
		return err
	}

	filename = fmt.Sprintf("%v@.service", UNIT_NAME_SIDEKICK)
	path = filepath.Join(outputDir, filename)

	if err := ioutil.WriteFile(path, []byte(sidekickFleetUnit), 0644); err != nil {
		return err
	}

	return nil

}

func (c CouchbaseFleet) WaitForFleetLaunch() error {

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

	unitFile, err := c.generateNodeFleetUnitFile()
	if err != nil {
		return "", err
	}
	log.Printf("Couchbase node fleet unit: %v", unitFile)

	jsonBytes, err := unitFileToJson(unitFile)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), err

}

func (c CouchbaseFleet) generateSidekickFleetUnitJson(unitNumber string) (string, error) {

	unitFile, err := c.generateSidekickFleetUnitFile(unitNumber)
	if err != nil {
		return "", err
	}
	log.Printf("Couchbase sidekick fleet unit: %v", unitFile)

	jsonBytes, err := unitFileToJson(unitFile)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), err

}

func unitFileToJson(unitFileContent string) ([]byte, error) {

	// deserialize to units
	opts, err := unit.Deserialize(strings.NewReader(unitFileContent))
	if err != nil {
		return nil, err
	}

	fleetUnit := struct {
		Options      []*unit.UnitOption `json:"options"`
		DesiredState string             `json:"desiredState"`
	}{
		Options:      opts,
		DesiredState: "launched",
	}

	bytes, err := json.Marshal(fleetUnit)
	return bytes, err

}

func (c CouchbaseFleet) generateNodeFleetUnitFile() (string, error) {

	assetName := "data/couchbase_node@.service.template"
	content, err := Asset(assetName)
	if err != nil {
		return "", fmt.Errorf("could not find asset: %v.  err: %v", err)
	}

	// run through go template engine
	tmpl, err := template.New("NodeUnitFile").Parse(string(content))
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

	return out.String(), nil

}

func (c CouchbaseFleet) generateSidekickFleetUnitFile(unitNumber string) (string, error) {

	content := `
[Unit]
Description=couchbase_sidekick
After=docker.service
Requires=docker.service
After=etcd.service
Requires=etcd.service
After=fleet.service
Requires=fleet.service
BindsTo=couchbase_node@{{ .UNIT_NUMBER }}.service
After=couchbase_node@{{ .UNIT_NUMBER }}.service

[Service]
TimeoutStartSec=0
EnvironmentFile=/etc/environment
ExecStartPre=-/usr/bin/docker kill couchbase-sidekick
ExecStartPre=-/usr/bin/docker rm couchbase-sidekick
ExecStartPre=/usr/bin/docker pull tleyden5iwx/couchbase-cluster-go:{{ .CONTAINER_TAG }}
ExecStart=/bin/bash -c '/usr/bin/docker run --name couchbase-sidekick --net=host tleyden5iwx/couchbase-cluster-go:{{ .CONTAINER_TAG }} update-wrapper couchbase-cluster start-couchbase-sidekick --local-ip=$COREOS_PRIVATE_IPV4'
ExecStop=/usr/bin/docker stop couchbase-sidekick

[X-Fleet]
MachineOf=couchbase_node@{{ .UNIT_NUMBER }}.service
`
	// run through go template engine
	tmpl, err := template.New("SidekickUnitFile").Parse(content)
	if err != nil {
		return "", err
	}

	params := struct {
		CB_VERSION    string
		CONTAINER_TAG string
		UNIT_NUMBER   string
	}{
		CB_VERSION:    c.CbVersion,
		CONTAINER_TAG: c.ContainerTag,
		UNIT_NUMBER:   unitNumber,
	}

	out := &bytes.Buffer{}

	// execute template and write to dest
	err = tmpl.Execute(out, params)
	if err != nil {
		return "", err
	}

	return out.String(), nil

}

func submitAndLaunchFleetUnitN(unitNumber int, unitName, fleetUnitJson string) error {

	if err := launchFleetUnitN(unitName, unitNumber, fleetUnitJson); err != nil {
		return err
	}

	return nil

}

func launchFleetUnitN(unitName string, unitNumber int, fleetUnitJson string) error {

	log.Printf("Launch fleet unit %v (%v)", unitName, unitNumber)

	endpointUrl := fmt.Sprintf("%v/units/%v@%v.service", FLEET_API_ENDPOINT, unitName, unitNumber)

	return PUT(endpointUrl, fleetUnitJson)

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
