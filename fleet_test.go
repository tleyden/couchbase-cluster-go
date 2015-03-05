package cbcluster

import (
	"log"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestGenerateNodeFleetUnitJson2(t *testing.T) {
	c := CouchbaseFleet{}
	unitJson, err := c.generateNodeFleetUnitJson2()
	assert.True(t, err == nil)
	log.Printf("unitJson: %v", unitJson)

	c.GenerateUnits("/tmp")

}
