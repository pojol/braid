package router

import "testing"

func TestMessage(t *testing.T) {

	NewMsg().WithReqHeader(&Header{Event: "111"}).Build()

}
