package cbcluster

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
	"github.com/tleyden/fakehttp"
)

func TestGenerateNodeFleetUnitJson(t *testing.T) {
	c := CouchbaseFleet{}
	unitJson, err := c.generateNodeFleetUnitJson()

	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)

}

func TestGenerateSidekickFleetUnitJson(t *testing.T) {
	c := CouchbaseFleet{}
	unitJson, err := c.generateSidekickFleetUnitJson("%i")

	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)
}

func TestFindAllUnits(t *testing.T) {

	FLEET_API_ENDPOINT = "http://localhost:5977"

	mockFleetApi := fakehttp.NewHTTPServerWithPort(5977)
	mockFleetApi.Start()

	mockFleetApi.Response(200, jsonHeaders(), "{}")

	c := CouchbaseFleet{}
	allUnits, err := c.findAllFleetUnits()
	assert.True(t, err == nil)

	assert.True(t, len(allUnits) == 0)

}

func jsonHeaders() map[string]string {
	return map[string]string{"Content-Type": "application/json"}
}
