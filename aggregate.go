package godm

import (
	"context"

	opts "github.com/md-salehzadeh/godm/options"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Pipeline define the pipeline for aggregate
type Pipeline []bson.D

// Aggregate is a handle to a aggregate
type Aggregate struct {
	ctx        context.Context
	pipeline   interface{}
	collection *mongo.Collection
	options    []opts.AggregateOptions
}

// All iterates the cursor from aggregate and decodes each document into results.
func (a *Aggregate) All(results interface{}) error {
	opts := options.Aggregate()
	if len(a.options) > 0 {
		opts = a.options[0].AggregateOptions
	}
	c, err := a.collection.Aggregate(a.ctx, a.pipeline, opts)
	if err != nil {
		return err
	}
	return c.All(a.ctx, results)
}

// One iterates the cursor from aggregate and decodes current document into result.
func (a *Aggregate) One(result interface{}) error {
	opts := options.Aggregate()
	if len(a.options) > 0 {
		opts = a.options[0].AggregateOptions
	}
	c, err := a.collection.Aggregate(a.ctx, a.pipeline, opts)
	if err != nil {
		return err
	}
	cr := Cursor{
		ctx:    a.ctx,
		cursor: c,
		err:    err,
	}
	defer cr.Close()
	if !cr.Next(result) {
		return ErrNoSuchDocuments
	}
	return err
}

// Iter return the cursor after aggregate
func (a *Aggregate) Iter() CursorI {
	opts := options.Aggregate()
	if len(a.options) > 0 {
		opts = a.options[0].AggregateOptions
	}
	c, err := a.collection.Aggregate(a.ctx, a.pipeline, opts)
	return &Cursor{
		ctx:    a.ctx,
		cursor: c,
		err:    err,
	}
}
