package options

import "go.mongodb.org/mongo-driver/mongo/options"

type RemoveOptions struct {
	RemoveHook interface{}
	*options.DeleteOptions
}
