package cbcluster

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"text/template"

	"github.com/coreos/go-etcd/etcd"
)

const (
	KEY_SYNC_GW_CONFIG = "/couchbase.com/sync-gateway/config"
	KEY_SYNC_GW_COMMIT = "/couchbase.com/sync-gateway/commit"
)

type SyncGwCluster struct {
	etcdClient               *etcd.Client
	EtcdServers              []string
	NumNodes                 int
	ContainerTag             string // Docker tag
	ConfigUrl                string
	CommitOrBranch           string
	CreateBucketName         string
	CreateBucketSize         int
	CreateBucketReplicaCount int
}

func NewSyncGwCluster(etcdServers []string) *SyncGwCluster {

	s := &SyncGwCluster{}

	if len(etcdServers) > 0 {
		s.EtcdServers = etcdServers
		log.Printf("Connect to explicit etcd servers: %v", s.EtcdServers)
	} else {
		s.EtcdServers = []string{}
		log.Printf("Connect to etcd on localhost")
	}
	s.ConnectToEtcd()
	return s

}

func (s *SyncGwCluster) ConnectToEtcd() {

	s.etcdClient = etcd.NewClient(s.EtcdServers)
	s.etcdClient.SetConsistency(etcd.STRONG_CONSISTENCY)
}

func (s *SyncGwCluster) ExtractDocOptArgs(arguments map[string]interface{}) error {

	numnodes, err := ExtractNumNodes(arguments)
	if err != nil {
		return err
	}
	s.NumNodes = numnodes

	configUrl, err := ExtractStringArg(arguments, "--config-url")
	if err != nil {
		return err
	}
	if configUrl == "" {
		return fmt.Errorf("Missing or empty config url")
	}
	s.ConfigUrl = configUrl

	commitOrBranch, _ := ExtractStringArg(arguments, "--sync-gw-commit")
	if commitOrBranch != "" {
		s.CommitOrBranch = commitOrBranch
	} else {
		// "image" means: use master branch commit when docker image built
		s.CommitOrBranch = "image"
	}

	createBucketName, _ := ExtractStringArg(arguments, "--create-bucket")
	if createBucketName != "" {
		s.CreateBucketName = createBucketName
	}

	createBucketSize, _ := ExtractIntArg(arguments, "--create-bucket-size")
	s.CreateBucketSize = createBucketSize
	if s.CreateBucketSize == 0 {
		s.CreateBucketSize = 512
	}

	createBucketReplicaCount, _ := ExtractIntArg(arguments, "--create-bucket-replicas")
	s.CreateBucketReplicaCount = createBucketReplicaCount
	if s.CreateBucketReplicaCount <= 0 {
		s.CreateBucketReplicaCount = 1
	}

	s.ContainerTag = ExtractDockerTagOrLatest(arguments)

	return nil
}

func (s SyncGwCluster) LaunchSyncGateway() error {

	// create bucket (if user asked for this)
	if err := s.createBucketIfNeeded(); err != nil {
		return err
	}

	// stash some values into etcd
	if err := s.addValuesEtcd(); err != nil {
		return err
	}

	// kick off fleet units
	if err := s.kickOffFleetUnits(); err != nil {
		return err
	}

	log.Printf("Your sync gateway cluster has been launched successfully!")

	return nil
}

func (s SyncGwCluster) kickOffFleetUnits() error {

	fleetUnitJson, err := s.generateFleetUnitJson()
	if err != nil {
		return err
	}

	for i := 1; i < s.NumNodes+1; i++ {

		if err := submitAndLaunchFleetUnitN(i, "sync_gw_node", fleetUnitJson); err != nil {
			return err
		}

	}

	return nil
}

func (s SyncGwCluster) generateFleetUnitJson() (string, error) {

	fleetUnitJsonTemplate := `
{
    "desiredState":"inactive",
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
            "value":"-/usr/bin/docker kill sync_gw"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"-/usr/bin/docker rm sync_gw"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"/usr/bin/docker pull tleyden5iwx/sync-gateway-coreos:{{ .CONTAINER_TAG }}"
        },
        {
            "section":"Service",
            "name":"ExecStartPre",
            "value":"/usr/bin/docker run --net=host tleyden5iwx/sync-gateway-coreos:{{ .CONTAINER_TAG }} couchbase-cluster-wrapper wait-until-running"
        },

        {
            "section":"Service",
            "name":"ExecStart",
            "value":"/bin/bash -c 'SYNC_GW_COMMIT=$(etcdctl get /couchbase.com/sync-gateway/commit);  SYNC_GW_CONFIG=$(etcdctl get /couchbase.com/sync-gateway/config); /usr/bin/docker run --name sync_gw --net=host tleyden5iwx/sync-gateway-coreos update-wrapper sync-gw-start -c $SYNC_GW_COMMIT -g $SYNC_GW_CONFIG'"
        },
        {
            "section":"Service",
            "name":"ExecStop",
            "value":"/usr/bin/docker stop sync_gw"
        },
        {
            "section":"X-Fleet",
            "name":"Conflicts",
            "value":"sync_gw_node*.service"
        }
    ]
}
`

	tmpl, err := template.New("sgw_fleet").Parse(fleetUnitJsonTemplate)
	if err != nil {
		return "", err
	}

	params := FleetParams{
		CONTAINER_TAG: s.ContainerTag,
	}

	out := &bytes.Buffer{}

	// execute template and write to dest
	err = tmpl.Execute(out, params)
	if err != nil {
		return "", err
	}

	return out.String(), nil

}

func (s SyncGwCluster) addValuesEtcd() error {

	// add values to etcd
	_, err := s.etcdClient.Set(KEY_SYNC_GW_CONFIG, s.ConfigUrl, 0)
	if err != nil {
		return err
	}
	_, err = s.etcdClient.Set(KEY_SYNC_GW_COMMIT, s.CommitOrBranch, 0)
	if err != nil {
		return err
	}

	return nil

}

func (s SyncGwCluster) createBucketIfNeeded() error {

	if s.CreateBucketName == "" {
		return nil
	}

	cb := NewCouchbaseCluster(s.EtcdServers)

	if err := cb.LoadAdminCredsFromEtcd(); err != nil {
		return err
	}

	liveNodeIp, err := cb.FindLiveNode()
	if err != nil {
		return err
	}
	cb.LocalCouchbaseIp = liveNodeIp

	ramQuotaMB := fmt.Sprintf("%v", s.CreateBucketSize)
	replicaNumber := fmt.Sprintf("%v", s.CreateBucketReplicaCount)

	data := url.Values{
		"name":          {s.CreateBucketName},
		"ramQuotaMB":    {ramQuotaMB},
		"authType":      {"none"},
		"replicaNumber": {replicaNumber},
		"proxyPort":     {"11215"},
	}

	return cb.CreateBucket(data)

}
