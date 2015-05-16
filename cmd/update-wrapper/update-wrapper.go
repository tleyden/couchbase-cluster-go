package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tleyden/go-etcd/etcd"
)

// This makes it easy to run a command after possibly first grabbing the latest version
// via "go get".  It attempts to solve the problem of having to rebuild docker images
// after tiny code changes, which makes the code-test loop too long.

const (
	KEY_ENABLE_CODE_REFRESH = "/couchbase.com/enable-code-refresh"
	SKIP_ETCD_CHECK         = "--skip-etcd-check"
)

func main() {

	// by default:
	//   - check etcd to see if there's a value for the enable-code-refresh key
	//   - if so, update to latest code
	// if first arg is --skip-etcd-check:
	//   - it will skip the check to etcd, and update to the latest code

	// get args with this binary stripped off
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalf("No target given in args")
		return
	}

	requiresUpdate := false

	switch args[0] {
	case SKIP_ETCD_CHECK:
		requiresUpdate = true
	default:
		// connect to etcd and see if we even need to update the code
		requiresUpdate = checkUpdateRequired()
	}

	if requiresUpdate {
		log.Printf("Update-Wrapper: updating to latest code")
		updateAndRebuild()
	} else {
		log.Printf("Update-Wrapper: skipping update")
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

	for i, arg := range args {
		if strings.Contains(arg, "--etcd-servers") {
			if strings.Contains(arg, "=") {
				// tokenize on = ..
				tokens := strings.Split(arg, "=")
				commaSepList := tokens[1] // srv1,srv2..
				return strings.Split(commaSepList, ",")
			} else {
				// assume the next arg is list of etcd servers
				if len(args) >= i+1 {
					nextArg := args[i+1]
					// split on "," and return array
					return strings.Split(nextArg, ",")
				}

			}

		}

	}

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

}

// This figures out:
//   - The target command being wrapped by this update-wrapper
//   - The arguments to pass to that target command
func getTargetAndRemainingArgs() (target string, remainingArgs []string, err error) {

	// get args with this binary stripped off
	args := os.Args[1:]

	if len(args) == 0 {
		return "", []string{}, fmt.Errorf("No target given in args")
	}

	// at this point, args[0] will either be:
	//   - the target command
	//   - a --skip-etcd-check flag

	target = args[0]
	if target == SKIP_ETCD_CHECK {
		// in this case, we we to strip this arg off as well, since we
		// don't care about the --skip-etcd-check flag
		args = os.Args[1:]
		target = args[0]
	}

	remainingArgs = args[1:]

	return target, remainingArgs, nil

}

func invokeTarget() error {

	target, remainingArgs, err := getTargetAndRemainingArgs()
	if err != nil {
		return err
	}

	cmd := exec.Command(target, remainingArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
