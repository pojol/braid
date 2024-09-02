package mgo

import (
	"braid/lib/tracer"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LCollection struct {
	trc  tracer.ITracer
	coll *mongo.Collection
}

func (c *LCollection) tracing(ctx context.Context, collection, cmd string, filter, document interface{}) (tracer.ISpan, error) {

	if c.trc != nil {

		span, err := c.trc.GetSpan("")
		if err == nil {
			span.Begin(ctx)

			span.SetTag("collection", collection)
			span.SetTag("cmd", cmd)

			if filter != nil {
				span.SetTag("filter", filter)
			}
			if document != nil {
				span.SetTag("document", document)
			}

			return span, nil
		}

		return nil, err
	}

	return nil, errors.New("can't get global tracing ptr")

}

func (c *LCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "find", filter, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.Find(ctx, filter, opts...)
}

func (c *LCollection) Count(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "count", filter, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.CountDocuments(ctx, filter, opts...)
}

func (c *LCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {

	span, err := c.tracing(ctx, c.coll.Name(), "findone", filter, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.FindOne(ctx, filter, opts...)
}

func (c *LCollection) UpdateOne(ctx context.Context, filter interface{}, document interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "updateone", filter, document)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.UpdateOne(ctx, filter, document, opts...)
}

func (c *LCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "insertone", nil, document)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.InsertOne(ctx, document, opts...)
}

func (c *LCollection) InsertMany(ctx context.Context, document []interface{}) (*mongo.InsertManyResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "insertmany", nil, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.InsertMany(ctx, document)
}

func (c *LCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "deleteone", filter, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.DeleteOne(ctx, filter, opts...)
}

func (c *LCollection) SetTTL(ctx context.Context, indexmodel mongo.IndexModel, opts ...*options.CreateIndexesOptions) (string, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "setttl", indexmodel, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.Indexes().CreateOne(ctx, indexmodel, opts...)
}

func (c *LCollection) CreateIndexes(ctx context.Context, indexmodels []mongo.IndexModel, opts ...*options.CreateIndexesOptions) ([]string, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "setttl", indexmodels, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.Indexes().CreateMany(ctx, indexmodels, opts...)
}

func (c *LCollection) IndexList(ctx context.Context, opts ...*options.ListIndexesOptions) (*mongo.Cursor, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "setttl", "", nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.Indexes().List(ctx, opts...)
}

func (c *LCollection) DeleteOneIndex(ctx context.Context, indexname string, opts ...*options.DropIndexesOptions) (bson.Raw, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "deleteIndex", "", nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.Indexes().DropOne(ctx, indexname, opts...)
}

func (c *LCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "deletemany", filter, nil)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.DeleteMany(ctx, filter, opts...)

}

func (c *LCollection) ReplaceOne(ctx context.Context, filter interface{}, document interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {

	span, err := c.tracing(ctx, c.coll.Name(), "replaceone", filter, document)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.ReplaceOne(ctx, filter, document, opts...)
}

func (c *LCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {

	span, err := c.tracing(ctx, c.coll.Name(), "replaceone", filter, update)
	if err == nil {
		defer span.End(ctx)
	}

	return c.coll.FindOneAndUpdate(ctx, filter, update, opts...)
}

func (c *LCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return c.coll.Aggregate(ctx, pipeline, opts...)
}

func (c *LCollection) CountDocuments(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	return c.coll.CountDocuments(ctx, filter, opts...)
}

func (c *LCollection) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	span, err := c.tracing(ctx, c.coll.Name(), "BulkWrite", nil, nil)
	if err == nil {
		defer span.End(ctx)
	}
	return c.coll.BulkWrite(ctx, models, opts...)
}
