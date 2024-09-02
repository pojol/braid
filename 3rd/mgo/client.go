package mgo

import (
	"braid/lib/tracer"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type lmgoclient struct {
	cli *mongo.Client
}

// StartSession 启动session
func (c *lmgoclient) StartSession(ctx context.Context, dbName string) (tracer.ISpan, mongo.Session, error) {

	/*
		span, terr := tracing(ctx, dbName, "session", nil, nil)
	*/
	session, serr := c.cli.StartSession()

	return nil, session, serr
}

func (c *lmgoclient) EndSession(ctx context.Context, span tracer.ISpan, session mongo.Session) {
	session.EndSession(ctx)
	if span != nil {
		span.End(ctx)
	}
}
