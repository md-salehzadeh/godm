package godm

import "context"

// CollectionI
//type CollectionI interface {
//	Find(filter interface{}) QueryI
//	InsertOne(doc interface{}) (*mongo.InsertOneResult, error)
//	InsertMany(docs ...interface{}) (*mongo.InsertManyResult, error)
//	Upsert(filter interface{}, replacement interface{}) (*mongo.UpdateResult, error)
//	UpdateOne(filter interface{}, update interface{}) error
//	UpdateAll(filter interface{}, update interface{}) (*mongo.UpdateResult, error)
//	DeleteOne(filter interface{}) error
//	RemoveAll(selector interface{}) (*mongo.DeleteResult, error)
//	EnsureIndex(indexes []string, isUnique bool)
//	EnsureIndexes(uniques []string, indexes []string)
//}

// Change holds fields for running a findAndModify command via the Query.Apply method.
type Change struct {
	Update    interface{} // update/replace document
	Replace   bool        // Whether to replace the document rather than updating
	Remove    bool        // Whether to remove the document found rather than updating
	Upsert    bool        // Whether to insert in case the document isn't found, take effect when Remove is false
	ReturnNew bool        // Should the modified document be returned rather than the old one, take effect when Remove is false
}

// CursorI Cursor interface
type CursorI interface {
	Next(result interface{}) bool
	Close() error
	Err() error
	All(results interface{}) error
	//ID() int64
}

// QueryI Query interface
type QueryI interface {
	setDocument(document interface{})
	Where(filters map[string]any) QueryI
	AndWhere(filters map[string]any) QueryI
	OrWhere(filters map[string]any) QueryI
	Sort(fields ...string) QueryI
	Select(fields ...string) QueryI
	Skip(n int64) QueryI
	BatchSize(n int64) QueryI
	Limit(n int64) QueryI
	One(result interface{}) error
	All(ctx context.Context, result_ ...interface{}) (interface{}, error)
	Count() (n int64, err error)
	Distinct(key string, result interface{}) error
	Cursor() CursorI
	Apply(change Change, result interface{}) error
	Hint(hint interface{}) QueryI
}

// AggregateI define the interface of aggregate
type AggregateI interface {
	All(results interface{}) error
	One(result interface{}) error
	Iter() CursorI
}
