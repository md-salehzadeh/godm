package options

import "go.mongodb.org/mongo-driver/mongo/options"

type InsertOneOptions struct {
	InsertHook interface{}
	*options.InsertOneOptions
}
type InsertManyOptions struct {
	InsertHook interface{}
	*options.InsertManyOptions
}
