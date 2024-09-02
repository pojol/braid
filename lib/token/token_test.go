package token

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {

	entityid := "111"
	token, err := Create(entityid)
	assert.Equal(t, err, nil)

	eid, err := Parse(token)
	assert.Equal(t, err, nil)
	assert.Equal(t, eid, entityid)

}
