package router

import "testing"

func TestMessage(t *testing.T) {

	NewMsgWrap().WithReqHeader(&Header{Event: "111"}).Build()

}
