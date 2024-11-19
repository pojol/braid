package msg

import (
	"context"
	"testing"

	"github.com/pojol/braid/router"
)

func TestMessage(t *testing.T) {

	NewBuilder(context.TODO()).WithReqHeader(&router.Header{Event: "111"}).Build()

}
