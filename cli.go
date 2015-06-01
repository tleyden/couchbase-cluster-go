package cbcluster

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func IsCommandEnabled(arguments map[string]interface{}, commandKey string) bool {
	val, ok := arguments[commandKey]
	if !ok {
		return false
	}
	boolVal, ok := val.(bool)
	if !ok {
		return false
	}
	return boolVal
}

// convert from comma separated list to a string slice
func ExtractEtcdServerList(docOptParsed map[string]interface{}) []string {

	rawServerList, found := docOptParsed["--etcd-servers"]
	if !found {
		log.Printf("ExtractEtcdServerList calling EtcdServerListFromEnv")
		return EtcdServerListFromEnv(docOptParsed)
	}

	rawServerListStr, ok := rawServerList.(string)
	if !ok {
		log.Printf("ExtractEtcdServerList calling EtcdServerListFromEnv")
		return EtcdServerListFromEnv(docOptParsed)
	}

	log.Printf("ExtractEtcdServerList returning: %v", rawServerListStr)
	return strings.Split(rawServerListStr, ",")

}

func EtcdServerListFromEnv(docOptParsed map[string]interface{}) []string {

	serviceName, found := docOptParsed["--k8s-service-name"]
	if !found {
		log.Printf("EtcdServerListFromEnv did not find --k8s-service-name")
		return nil
	}

	hostNameEnvVar := fmt.Sprintf("%v_SERVICE_HOST", serviceName)
	hostName := os.Getenv(hostNameEnvVar)

	portEnvVar := fmt.Sprintf("%v_SERVICE_PORT", serviceName)
	port := os.Getenv(portEnvVar)

	log.Printf("Etcd hostname: %v port: %v", hostName, port)

	if len(hostName) == 0 {
		log.Printf("EtcdServerListFromEnv did not find ENV variable: %v", hostNameEnvVar)
		return nil
	}
	if len(port) == 0 {
		log.Printf("EtcdServerListFromEnv did not find ENV variable: %v", portEnvVar)
		return nil
	}

	return []string{fmt.Sprintf("%v:%v", hostName, port)}

}

func ExtractUserPass(docOptParsed map[string]interface{}) (string, error) {
	return ExtractStringArg(docOptParsed, "--userpass")
}

func ExtractDockerTagOrLatest(docOptParsed map[string]interface{}) string {
	dockerTag, err := ExtractStringArg(docOptParsed, "--docker-tag")
	if err != nil || dockerTag == "" {
		return "latest"
	}
	return dockerTag

}

func ExtractCbVersion(docOptParsed map[string]interface{}) (string, error) {

	rawVersion, err := ExtractStringArg(docOptParsed, "--version")
	if err != nil {
		return rawVersion, err
	}

	// if they passed in "latest", ignore the edition and just use the
	// the "latest" docker tag, which will pull latest enterprise edition.
	if rawVersion == "latest" {
		return rawVersion, nil
	}

	rawEdition, _ := docOptParsed["--edition"]
	if rawEdition == nil || rawEdition == "" {
		rawEdition = "community"
	}

	switch rawEdition {
	case "community":
		rawVersion = fmt.Sprintf("community-%v", rawVersion)
	case "enterprise":
		rawVersion = fmt.Sprintf("enterprise-%v", rawVersion)
	default:
		return "", fmt.Errorf("Invalid value for edition: %v", rawEdition)
	}

	return rawVersion, nil

}

func ExtractIntArg(docOptParsed map[string]interface{}, argToExtract string) (int, error) {

	rawVal, found := docOptParsed[argToExtract]
	if !found {
		return -1, fmt.Errorf("Did not find arg: %v", argToExtract)
	}

	stringVal, ok := rawVal.(string)
	if !ok {
		return -1, fmt.Errorf("Invalid type for %v", argToExtract)
	}

	intVal, err := strconv.ParseInt(stringVal, 10, 64)
	if err != nil {
		return -1, err
	}

	return int(intVal), nil

}

func ExtractStringArg(docOptParsed map[string]interface{}, argToExtract string) (string, error) {

	rawVal, found := docOptParsed[argToExtract]
	if !found {
		return "", fmt.Errorf("Did not find arg: %v", argToExtract)
	}

	stringVal, ok := rawVal.(string)
	if !ok {
		return "", fmt.Errorf("Invalid type for %v", argToExtract)
	}

	return stringVal, nil

}

func ExtractBoolArg(docOptParsed map[string]interface{}, argToExtract string) bool {

	rawVal, found := docOptParsed[argToExtract]
	if !found {
		return false
	}

	boolVal, ok := rawVal.(bool)
	if !ok {
		return false
	}

	return boolVal

}

func ExtractSkipCheckCleanState(docOptParsed map[string]interface{}) bool {

	return ExtractBoolArg(docOptParsed, "--skip-clean-slate-check")

}

func ExtractNumNodes(docOptParsed map[string]interface{}) (int, error) {

	return ExtractIntArg(docOptParsed, "--num-nodes")

}
