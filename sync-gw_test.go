package cbcluster

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestGenerateSyncGwNodeFleetUnitJson(t *testing.T) {
	s := SyncGwCluster{}
	unitJson, err := s.generateFleetUnitJson()
	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)

}

func TestGenerateSyncGwSidekickFleetUnitJson(t *testing.T) {
	s := SyncGwCluster{}
	unitJson, err := s.generateFleetSidekickUnitJson(1)
	assert.True(t, err == nil)
	assert.True(t, len(unitJson) > 0)
}
