package cbcluster

import (
	"fmt"
	"log"
	"net/url"

	"github.com/coreos/go-etcd/etcd"
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

	createBucketReplicaCount, _ := ExtractIntArg(arguments, "--create-bucket-replicas")
	s.CreateBucketReplicaCount = createBucketReplicaCount

	s.ContainerTag = ExtractDockerTagOrLatest(arguments)

	return nil
}

func (s SyncGwCluster) LaunchSyncGateway() error {

	// create bucket (if user asked for this)
	if err := s.createBucketIfNeeded(); err != nil {
		return err
	}

	// # add values to etcd
	// etcdctl set /couchbase.com/sync-gateway/config "$configFileOrURL"
	// etcdctl set /couchbase.com/sync-gateway/commit "$commit"

	// kick off fleet units

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
