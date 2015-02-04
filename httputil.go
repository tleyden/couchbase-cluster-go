package cbcluster

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type middlewareFunc func(req *http.Request)

func getJsonDataMiddleware(endpointUrl string, into interface{}, middleware middlewareFunc) error {

	client := &http.Client{}

	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return err
	}

	middleware(req)

	// req.SetBasicAuth(c.AdminUsername, c.AdminPassword)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Failed to GET %v.  Status code: %v", endpointUrl, resp.StatusCode)
	}

	d := json.NewDecoder(resp.Body)
	return d.Decode(into)

}

func getJsonData(endpointUrl string, into interface{}) error {
	return getJsonDataMiddleware(endpointUrl, into, func(req *http.Request) {})
}
