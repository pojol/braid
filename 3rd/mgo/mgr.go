package mgo

import (
	"braid/lib/tracer"
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mgr struct {
	sync.RWMutex

	collections map[string]*lmgoclient
	trc         tracer.ITracer
}

var (
	// Mgr is a global instance of mgr
	mgrPtr *mgr
	once   sync.Once
)

func Build(opts ...Option) error {

	once.Do(func() {
		p := &Parm{}
		for _, opt := range opts {
			opt(p)
		}

		mgrPtr = &mgr{
			collections: make(map[string]*lmgoclient),
			trc:         p.tracer,
		}

		for _, c := range p.conns {

			optCli := options.Client()
			optCli.ApplyURI(c.Addr)
			optCli.SetConnectTimeout(p.connTimeout)
			optCli.SetMaxPoolSize(p.poolSize)

			cli, err := mongo.Connect(context.TODO(), optCli)
			if err != nil {
				panic(fmt.Errorf("lmgo connect %s err: %s", c.Name, err.Error()))
			}

			err = cli.Ping(context.Background(), nil)
			if err != nil {
				panic(fmt.Errorf("lmgo ping %s err: %s", c.Name, err.Error()))
			}

			mgrPtr.collections[c.Name] = &lmgoclient{
				cli: cli,
			}
		}
	})

	return nil
}

func Collection(db string, collection string) *LCollection {
	mgrPtr.RLock()
	defer mgrPtr.RUnlock()

	cli, ok := mgrPtr.collections[db]
	if !ok {
		panic(fmt.Errorf("lmgo get collection err: %s", "db not found"))
	}

	return &LCollection{
		coll: cli.cli.Database(db).Collection(collection),
	}
}

// Destroy 销毁 mongodb 驱动
func (m *mgr) Destroy() {

	for _, v := range m.collections {
		v.cli.Disconnect(context.TODO())
	}

}
