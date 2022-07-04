package options

import "go.mongodb.org/mongo-driver/mongo/options"

type UpsertOptions struct {
	UpsertHook interface{}
	*options.ReplaceOptions
}
