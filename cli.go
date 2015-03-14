package cbcluster

import (
	"fmt"
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
		return nil
	}

	rawServerListStr, ok := rawServerList.(string)
	if !ok {
		return nil
	}

	return strings.Split(rawServerListStr, ",")

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

	// We possibly need to tack on "ent" to the end of the version string
	// in the case of enterprise versions, so that we can find the docker
	// image on dockerhub when launched from the fleet script.
	rawEdition, _ := docOptParsed["--edition"]

	if rawEdition != nil {
		switch rawEdition {
		case "":
			// if no edition given, use community edition by not
			// appending anything to version
		case "community":
			// nothing to do in this case, since the docker images don't have
			// community in their name (yet)
		case "enterprise":
			rawVersion = fmt.Sprintf("%v-ent", rawVersion)
		default:
			return "", fmt.Errorf("Invalid value for edition: %v", rawEdition)
		}
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
