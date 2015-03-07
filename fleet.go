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

	"github.com/coreos/fleet/schema"
	"github.com/coreos/go-systemd/unit"
	"github.com/tleyden/go-etcd/etcd"
)

const (
	UNIT_NAME_NODE     = "couchbase_node"
	UNIT_NAME_SIDEKICK = "couchbase_sidekick"
)

var (
	FLEET_API_ENDPOINT = "http://localhost:49153/fleet/v1"
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

func (c CouchbaseFleet) StopUnits() error {

	// set the /couchbase.com/remove-rebalance-disabled flag in etcd since
	// otherwise, it will try to remove and rebalance the node, which is not
	// what we want when stopping all units.

	// set the ttl to be 5 minutes, since there's nothing in place yet to
	// block until all the units have stopped
	// (TODO: this should get added .. it waits for all units to stop, and then
	// it removes the /couchbase.com/remove-rebalance-disabled flag)
	ttlSeconds := uint64(300)
	_, err := c.etcdClient.Set(KEY_REMOVE_REBALANCE_DISABLED, "true", ttlSeconds)
	if err != nil {
		return err
	}

	// call ManipulateUnits with a function that will stop them
	unitStopper := func(unit *schema.Unit) error {

		// stop the unit by updating desiredState to inactive
		// and posting to fleet api
		endpointUrl := fmt.Sprintf("%v/%v", FLEET_API_ENDPOINT, unit.Name)
		return PUT(endpointUrl, `{"desiredState": "inactive"}`)

	}

	return c.ManipulateUnits(unitStopper)

}

func (c CouchbaseFleet) DestroyUnits() error {

	ttlSeconds := uint64(300)
	_, err := c.etcdClient.Set(KEY_REMOVE_REBALANCE_DISABLED, "true", ttlSeconds)
	if err != nil {
		return err
	}

	// call ManipulateUnits with a function that will stop them
	unitDestroyer := func(unit *schema.Unit) error {

		// stop the unit by updating desiredState to inactive
		// and posting to fleet api
		endpointUrl := fmt.Sprintf("%v/%v", FLEET_API_ENDPOINT, unit.Name)
		return DELETE(endpointUrl)

	}

	return c.ManipulateUnits(unitDestroyer)

}

type UnitManipulator func(unit *schema.Unit) error

func (c CouchbaseFleet) ManipulateUnits(unitManipulator UnitManipulator) error {

	// find all the units
	allUnits, err := c.findAllFleetUnits()
	if err != nil {
		return err
	}

	unitNamePatterns := []string{UNIT_NAME_NODE, UNIT_NAME_SIDEKICK}

	// filter the ones out that have the name pattern we care about (couchbase_node)
	units := c.filterFleetUnits(allUnits, unitNamePatterns)

	for _, unit := range units {
		if err := unitManipulator(unit); err != nil {
			return err
		}
	}

	return nil

}

func (c CouchbaseFleet) findAllFleetUnits() (units []*schema.Unit, err error) {

	endpointUrl := ""
	maxAttempts := 10000
	sleepSeconds := 0
	nextPageToken := ""

	log.Printf("findAllFleetUnits()")

	worker := func() (finished bool, err error) {

		// append a next page token to url if needed
		if len(nextPageToken) > 0 {
			endpointUrl = fmt.Sprintf("%v/units?nextPageToken=%v", FLEET_API_ENDPOINT, nextPageToken)
		} else {
			endpointUrl = fmt.Sprintf("%v/units", FLEET_API_ENDPOINT)
		}

		log.Printf("Getting units from %v", endpointUrl)

		unitPage := schema.UnitPage{}
		if err := getJsonData(endpointUrl, &unitPage); err != nil {
			return true, err
		}

		// add all units to return value
		for _, unit := range unitPage.Units {
			units = append(units, unit)
		}

		// if no more pages, we are finished
		areWeFinished := len(unitPage.NextPageToken) == 0

		return areWeFinished, nil

	}

	sleeper := func(numAttempts int) (bool, int) {
		if numAttempts > maxAttempts {
			return false, -1
		}
		return true, sleepSeconds
	}

	if err := RetryLoop(worker, sleeper); err != nil {
		return nil, err
	}

	return units, nil

}

func (c CouchbaseFleet) filterFleetUnits(units []*schema.Unit, filters []string) (filteredUnits []*schema.Unit) {

	stringContainsAny := func(s string, filters []string) bool {
		for _, filter := range filters {
			if strings.Contains(s, filter) {
				return true
			}
		}
		return false
	}

	for _, unit := range units {

		if stringContainsAny(unit.Name, filters) {
			filteredUnits = append(filteredUnits, unit)
		}
	}

	return filteredUnits

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
		return "", fmt.Errorf("could not find asset: %v.  err: %v", assetName, err)
	}

	params := struct {
		CB_VERSION    string
		CONTAINER_TAG string
	}{
		CB_VERSION:    c.CbVersion,
		CONTAINER_TAG: c.ContainerTag,
	}

	return generateUnitFileFromTemplate(content, params)

}

func (c CouchbaseFleet) generateSidekickFleetUnitFile(unitNumber string) (string, error) {

	assetName := "data/couchbase_sidekick@.service.template"
	content, err := Asset(assetName)
	if err != nil {
		return "", fmt.Errorf("could not find asset: %v.  err: %v", assetName, err)
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

	return generateUnitFileFromTemplate(content, params)

}

func generateUnitFileFromTemplate(templateContent []byte, params interface{}) (string, error) {

	// run through go template engine
	tmpl, err := template.New("Template").Parse(string(templateContent))
	if err != nil {
		return "", err
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

func DELETE(endpointUrl string) error {

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", endpointUrl, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Unexpected status code in response")
	}

	return nil

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
