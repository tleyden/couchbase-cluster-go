package cbcluster

import (
	"fmt"
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

func ExtractCbVersion(docOptParsed map[string]interface{}) (string, error) {
	return ExtractStringArg(docOptParsed, "--version")
}

func ExtractIntArg(docOptParsed map[string]interface{}, argToExtract string) (int, error) {

	rawVal, found := docOptParsed[argToExtract]
	if !found {
		return -1, fmt.Errorf("Did not find arg: %v", argToExtract)
	}

	intVal, ok := rawVal.(int)
	if !ok {
		return -1, fmt.Errorf("Invalid type for %v", argToExtract)
	}

	return intVal, nil

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

func ExtractNumNodes(docOptParsed map[string]interface{}) (int, error) {

	return ExtractIntArg(docOptParsed, "--num-nodes")

}
