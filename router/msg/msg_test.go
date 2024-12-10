package msg

import (
	"context"
	"testing"

	"github.com/pojol/braid/router"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {

	NewBuilder(context.TODO()).WithReqHeader(&router.Header{Event: "111"}).Build()

}

type testObj struct {
	Name        string
	Gold        int
	Probability float64
	Lst         []string
}

func TestObjectSerialize(t *testing.T) {
	b := NewBuilder(context.TODO())

	b.WithReqCustomObject(&testObj{
		Name:        "test1",
		Gold:        10,
		Probability: 1.11,
		Lst:         []string{"a", "b", "c"},
	})
	assert.Equal(t, nil, b.wrapper.Err)

	obj := &testObj{}
	err := b.GetReqCustomObject(obj)
	assert.Equal(t, nil, err)

	assert.Equal(t, obj.Gold, 10)
	assert.Equal(t, obj.Name, "test1")
	assert.Equal(t, obj.Probability, 1.11)
	assert.Equal(t, obj.Lst, []string{"a", "b", "c"})
}
