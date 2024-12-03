package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/lib/log"
	"github.com/pojol/braid/tests/mock"
)

var factory *mock.MockActorFactory
var loader core.IActorLoader

func TestMain(m *testing.M) {
	slog, _ := log.NewServerLogger("test")
	log.SetSLog(slog)

	defer log.Sync()

	factory = mock.BuildActorFactory()
	loader = mock.BuildDefaultActorLoader(factory)

	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()
	redis.BuildClientWithOption(redis.WithAddr(fmt.Sprintf("redis://%s", mr.Addr())))

	os.Exit(m.Run())
}
