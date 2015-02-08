package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/coreos/go-etcd/etcd"
)

// This makes it easy to run a command after possibly first grabbing the latest version
// via "go get".  It attempts to solve the problem of having to rebuild docker images
// after tiny code changes, which makes the code-test loop too long.

const (
	KEY_ENABLE_CODE_REFRESH = "/couchbase.com/enable-code-refresh"
)

func main() {

	// connect to etcd and see if we even need to update the code
	requiresUpdate := checkUpdateRequired()

	if requiresUpdate {
		log.Printf("Going to try to update to latest code")
		updateAndRebuild()
	}

	// invoke appropriate binary with command line args
	if err := invokeTarget(); err != nil {
		log.Fatalf("Error invoking target: %v", err)
	}

}

func checkUpdateRequired() bool {

	// find etcd server list from args
	etcdServers := findEtcdServersFromArgs(os.Args)

	// create etcd client
	etcdClient := etcd.NewClient(etcdServers)
	etcdClient.SetConsistency(etcd.STRONG_CONSISTENCY)

	// get value of /couchbase.com/enable-code-refresh
	_, err := etcdClient.Get(KEY_ENABLE_CODE_REFRESH, false, false)
	if err != nil {
		// if we got an error, assume key not there
		return false
	}

	return true

}

func findEtcdServersFromArgs(args []string) []string {

	// TODO
	/*
		for _, arg := range args {



		}
	*/
	return []string{}

}

func updateAndRebuild() {

	// go get -u -v github.com/tleyden/couchbase-cluster-go/...
	goGetArgs := []string{
		"get",
		"-u",
		"-v",
		"github.com/tleyden/couchbase-cluster-go/...",
	}
	cmdGoGet := exec.Command("go", goGetArgs...)
	cmdGoGet.Stdout = os.Stdout
	cmdGoGet.Stderr = os.Stderr

	if err := cmdGoGet.Run(); err != nil {
		log.Printf("Error trying to go get: %v.  Ignoring it", err)
		return
	}

	cmdGodepInstall := exec.Command(
		"godep",
		"go",
		"install",
		"github.com/tleyden/couchbase-cluster-go/...",
	)
	cmdGodepInstall.Dir = fmt.Sprintf(
		"%v/src/github.com/tleyden/couchbase-cluster-go",
		os.Getenv("GOPATH"),
	)

	cmdGodepInstall.Stdout = os.Stdout
	cmdGodepInstall.Stderr = os.Stderr

	if err := cmdGodepInstall.Run(); err != nil {
		log.Printf("Error trying to godep install: %v.  Ignoring it", err)
		return
	}

}

func invokeTarget() error {

	// get args with this binary stripped off
	args := os.Args[1:]

	if len(args) == 0 {
		return fmt.Errorf("No target given in args")
	}

	target := args[0]

	remainingArgs := args[1:]

	cmd := exec.Command(target, remainingArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
