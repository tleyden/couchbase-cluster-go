package cbcluster

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
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
