package cbcluster

import (
	"log"
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestGenerateSyncGwNodeFleetUnitJson(t *testing.T) {
	s := SyncGwCluster{}
	unitJson, err := s.generateFleetUnitJson()
	log.Printf("unitJson: %v", unitJson)
	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)

}

func TestGenerateSyncGwSidekickFleetUnitJson(t *testing.T) {
	s := SyncGwCluster{}
	unitJson, err := s.generateFleetSidekickUnitJson(1)
	log.Printf("unitJson: %v", unitJson)
	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)
}
