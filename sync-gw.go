package cbcluster

import (
	"fmt"
	"log"
	"net/url"

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
	s.ConfigUrl = configUrl

	commitOrBranch, _ := ExtractStringArg(arguments, "--sync-gw-commit")
	if commitOrBranch != "" {
		s.CommitOrBranch = commitOrBranch
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
	if s.CreateBucketReplicaCount == 0 {
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

	return nil
}

func (s SyncGwCluster) kickOffFleetUnits() error {

}

func (s SyncGwCluster) addValuesEtcd() error {

	// add values to etcd
	_, err := c.etcdClient.Set(KEY_SYNC_GW_CONFIG, c.ConfigUrl, 0)
	if err != nil {
		return err
	}
	_, err := c.etcdClient.Set(KEY_SYNC_GW_COMMIT, c.CommitOrBranch, 0)
	if err != nil {
		return err
	}

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
	StupidPortHack(cb)

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
