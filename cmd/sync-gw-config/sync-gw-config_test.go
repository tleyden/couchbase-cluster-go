package main

import (
	"testing"

	"github.com/couchbaselabs/go.assert"
)

func TestRequiresRewrite(t *testing.T) {

	testString1 := "Some stufff {{ .COUCHBASE_SERVER_IP }} blah"
	testString2 := "Some stufff blah"
	assert.True(t, requiresRewrite(testString1))
	assert.False(t, requiresRewrite(testString2))

}
