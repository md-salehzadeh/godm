package options

import "go.mongodb.org/mongo-driver/mongo/options"

type IndexModel struct {
	Key []string // Index key fields; prefix name with dash (-) for descending order
	*options.IndexOptions
}
