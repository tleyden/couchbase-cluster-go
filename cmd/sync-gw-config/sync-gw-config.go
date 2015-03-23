package main

import (
	"io/ioutil"
	"log"
	"regexp"

	"github.com/docopt/docopt-go"
	"github.com/tleyden/couchbase-cluster-go"
)

func main() {

	usage := `Sync-Gw-Config.

Usage:
  sync-gw-config rewrite --destination=<config-dest> [--etcd-servers=<server-list>]
  sync-gw-config -h | --help

Options:
  -h --help     Show this screen.
  --etcd-servers=<server-list>  Comma separated list of etcd servers, or omit to connect to etcd running on localhost
  --destination=<config-dest> The path where the updated config should be written
`

	arguments, err := docopt.Parse(usage, nil, true, "Sync-Gw-Config", false)
	log.Printf("args: %v.  err: %v", arguments, err)

	if cbcluster.IsCommandEnabled(arguments, "rewrite") {
		if err := rewriteConfig(arguments); err != nil {
			log.Fatalf("Failed: %v", err)
		}
		return
	}

	log.Printf("Nothing to do!")

}

// does this config need to be rerwritten?  if it doesn't have
// any placeholder variables, then the answer is no.
func requiresRewrite(syncGwConfig string) bool {
	re := regexp.MustCompile(`{{.*}}`)
	placeholder := re.FindString(syncGwConfig)
	return placeholder != ""
}

func rewriteConfig(arguments map[string]interface{}) error {

	etcdServers := cbcluster.ExtractEtcdServerList(arguments)
	dest, err := cbcluster.ExtractStringArg(arguments, "--destination")
	if err != nil {
		return err
	}

	syncGwCluster := cbcluster.NewSyncGwCluster(etcdServers)

	// get the sync gw config from etcd (cbcluster.KEY_SYNC_GW_CONFIG)
	syncGwConfig, err := syncGwCluster.FetchSyncGwConfig()

	if !requiresRewrite(syncGwConfig) {
		log.Printf("No placeholder variables in config, no rewrite required")
		return nil
	}

	// get a couchbase live node
	couchbaseCluster := cbcluster.NewCouchbaseCluster(etcdServers)
	liveNodeIp, err := couchbaseCluster.FindLiveNode()

	log.Printf("LiveNodeIp: %v", liveNodeIp)

	if err != nil {
		return err
	}

	// run the sync gw config through go templating engine
	updatedConfig, err := syncGwCluster.UpdateConfig(liveNodeIp, syncGwConfig)
	if err != nil {
		return err
	}

	log.Printf("Rewritten sync gw config file: %v", string(updatedConfig))

	// write the new config to the dest file
	if err := ioutil.WriteFile(dest, updatedConfig, 0644); err != nil {
		return err
	}

	return nil

}
