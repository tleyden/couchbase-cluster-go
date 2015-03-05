package cbcluster

import (
	"log"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestGenerateNodeFleetUnitJson(t *testing.T) {
	c := CouchbaseFleet{}
	unitJson, err := c.generateNodeFleetUnitJson()
	assert.True(t, err == nil)
	log.Printf("unitJson: %v", unitJson)

}

func TestGenerateSidekickFleetUnitJson(t *testing.T) {
	c := CouchbaseFleet{}
	unitJson, err := c.generateSidekickFleetUnitJson("%i")
	assert.True(t, err == nil)
	log.Printf("sidekick unitJson: %v", unitJson)

	c.GenerateUnits("/tmp")

}
