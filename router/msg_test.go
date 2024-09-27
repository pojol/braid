package router

import (
	"context"
	"testing"
)

func TestMessage(t *testing.T) {

	NewMsgWrap(context.TODO()).WithReqHeader(&Header{Event: "111"}).Build()

}
