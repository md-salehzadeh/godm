package godm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/md-salehzadeh/godm/middleware"
	"github.com/md-salehzadeh/godm/operator"
	gOpts "github.com/md-salehzadeh/godm/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Query struct definition
type Query struct {
	filter    bson.D
	sort      bson.D
	project   bson.D
	hint      interface{}
	limit     *int64
	skip      *int64
	batchSize *int64

	ctx        context.Context
	collection *mongo.Collection
	opts       []gOpts.FindOptions
	registry   *bsoncodec.Registry
}

// BatchSize sets the value for the BatchSize field.
// Means the maximum number of documents to be included in each batch returned by the server.
func (q *Query) BatchSize(n int64) QueryI {
	q.batchSize = &n

	return q
}

func makeWhere(filters map[string]any) bson.D {
	filter := bson.D{}

	if len(filters) > 0 {
		for field, value := range filters {
			var _operator string
			var keys []string

			if strings.HasSuffix(field, " <") {
				keys = []string{" <"}

				_operator = operator.Lt
			} else if strings.HasSuffix(field, " <=") {
				keys = []string{" <="}

				_operator = operator.Lte
			} else if strings.HasSuffix(field, " >") {
				keys = []string{" >"}

				_operator = operator.Gt
			} else if strings.HasSuffix(field, " >=") {
				keys = []string{" >="}

				_operator = operator.Gte
			} else if strings.HasSuffix(field, " in") || strings.HasSuffix(field, " IN") {
				keys = []string{" in", " IN"}

				_operator = operator.In
			} else if strings.HasSuffix(field, " not in") || strings.HasSuffix(field, " NOT IN") {
				keys = []string{" not in", " NOT IN"}

				_operator = operator.Nin
			} else if strings.HasSuffix(field, " !=") || strings.HasSuffix(field, " <>") {
				keys = []string{" !=", " <>"}

				_operator = operator.Ne
			} else {
				_operator = operator.Eq
			}

			if len(keys) > 0 {
				for _, key := range keys {
					field = strings.Replace(field, key, "", -1)
				}
			}

			field = strings.Trim(field, " ")

			filter = append(filter, bson.E{field, bson.D{{_operator, value}}})
		}
	}

	return filter
}

func (q *Query) Where(filters map[string]any) QueryI {
	newFilter := makeWhere(filters)

	q.filter = append(q.filter, newFilter...)

	return q
}

func (q *Query) AndWhere(filters map[string]any) QueryI {
	if q.filter == nil {
		return q.Where(filters)
	}

	lastFilter := q.filter

	newFilter := makeWhere(filters)

	q.filter = bson.D{
		{operator.And,
			bson.A{
				lastFilter,
				newFilter,
			},
		},
	}

	return q
}

func (q *Query) OrWhere(filters map[string]any) QueryI {
	if q.filter == nil {
		return q.Where(filters)
	}

	lastFilter := q.filter

	newFilter := makeWhere(filters)

	q.filter = bson.D{
		{operator.Or,
			bson.A{
				lastFilter,
				newFilter,
			},
		},
	}

	return q
}

// Sort is Used to set the sorting rules for the returned results
// Format: "age" or "age asc" means to sort the age field in ascending order, "age desc" means in descending order
// When multiple sort fields are passed in at the same time, they are arranged in the order in which the fields are passed in.
// For example, {"age", "name desc"}, first sort by age in ascending order, then sort by name in descending order
func (q *Query) Sort(fields ...string) QueryI {
	if len(fields) > 0 {
		for _, field := range fields {
			key, sort := ParseSortField(field)

			if key == "" {
				panic("Sort: empty field name")
			}

			q.sort = append(q.sort, bson.E{Key: key, Value: sort})
		}
	}

	return q
}

// Select is used to determine which fields are displayed or not displayed in the returned results
// Format: bson.M{"age": 1} means that only the age field is displayed
// bson.M{"age": 0} means to display other fields except age
// When _id is not displayed and is set to 0, it will be returned to display
func (q *Query) Select(fields ...string) QueryI {
	if len(fields) > 0 {
		for _, field := range fields {
			key, visible := ParseSelectField(field)

			if key == "" {
				panic("Select: empty field name")
			}

			q.project = append(q.project, bson.E{Key: key, Value: visible})
		}
	}

	return q
}

// Skip skip n records
func (q *Query) Skip(n int64) QueryI {
	q.skip = &n

	return q
}

// Hint sets the value for the Hint field.
// This should either be the index name as a string or the index specification
// as a document. The default value is nil, which means that no hint will be sent.
func (q *Query) Hint(hint interface{}) QueryI {
	q.hint = hint

	return q
}

// Limit limits the maximum number of documents found to n
// The default value is 0, and 0  means no limit, and all matching results are returned
// When the limit value is less than 0, the negative limit is similar to the positive limit, but the cursor is closed after returning a single batch result.
// Reference https://docs.mongodb.com/manual/reference/method/cursor.limit/index.html
func (q *Query) Limit(n int64) QueryI {
	q.limit = &n

	return q
}

// One query a record that meets the filter conditions
// If the search fails, an error will be returned
func (q *Query) One(result interface{}) error {
	if len(q.opts) > 0 {
		if err := middleware.Do(q.ctx, q.opts[0].QueryHook, operator.BeforeQuery); err != nil {
			return err
		}
	}

	opt := options.FindOne()

	if q.sort != nil {
		opt.SetSort(q.sort)
	}

	if q.project != nil {
		opt.SetProjection(q.project)
	}

	if q.skip != nil {
		opt.SetSkip(*q.skip)
	}

	if q.hint != nil {
		opt.SetHint(q.hint)
	}

	err := q.collection.FindOne(q.ctx, q.filter, opt).Decode(result)

	if err != nil {
		return err
	}

	if len(q.opts) > 0 {
		if err := middleware.Do(q.ctx, q.opts[0].QueryHook, operator.AfterQuery); err != nil {
			return err
		}
	}

	return nil
}

// All query multiple records that meet the filter conditions
// The static type of result must be a slice pointer
func (q *Query) All(result interface{}) error {
	if len(q.opts) > 0 {
		if err := middleware.Do(q.ctx, q.opts[0].QueryHook, operator.BeforeQuery); err != nil {
			return err
		}
	}

	opt := options.Find()

	if q.sort != nil {
		opt.SetSort(q.sort)
	}

	if q.project != nil {
		opt.SetProjection(q.project)
	}

	if q.limit != nil {
		opt.SetLimit(*q.limit)
	}

	if q.skip != nil {
		opt.SetSkip(*q.skip)
	}

	if q.hint != nil {
		opt.SetHint(q.hint)
	}

	if q.batchSize != nil {
		opt.SetBatchSize(int32(*q.batchSize))
	}

	var err error
	var cursor *mongo.Cursor

	cursor, err = q.collection.Find(q.ctx, q.filter, opt)

	c := Cursor{
		ctx:    q.ctx,
		cursor: cursor,
		err:    err,
	}

	err = c.All(result)

	if err != nil {
		return err
	}

	if len(q.opts) > 0 {
		if err := middleware.Do(q.ctx, q.opts[0].QueryHook, operator.AfterQuery); err != nil {
			return err
		}
	}

	return nil
}

// Count count the number of eligible entries
func (q *Query) Count() (n int64, err error) {
	opt := options.Count()

	if q.limit != nil {
		opt.SetLimit(*q.limit)
	}

	if q.skip != nil {
		opt.SetSkip(*q.skip)
	}

	return q.collection.CountDocuments(q.ctx, q.filter, opt)
}

// Distinct gets the unique value of the specified field in the collection and return it in the form of slice
// result should be passed a pointer to slice
// The function will verify whether the static type of the elements in the result slice is consistent with the data type obtained in mongodb
// reference https://docs.mongodb.com/manual/reference/command/distinct/
func (q *Query) Distinct(key string, result interface{}) error {
	resultVal := reflect.ValueOf(result)

	if resultVal.Kind() != reflect.Ptr {
		return ErrQueryNotSlicePointer
	}

	resultElmVal := resultVal.Elem()

	if resultElmVal.Kind() != reflect.Interface && resultElmVal.Kind() != reflect.Slice {
		return ErrQueryNotSliceType
	}

	opt := options.Distinct()

	res, err := q.collection.Distinct(q.ctx, key, q.filter, opt)

	if err != nil {
		return err
	}

	registry := q.registry

	if registry == nil {
		registry = bson.DefaultRegistry
	}

	valueType, valueBytes, err_ := bson.MarshalValueWithRegistry(registry, res)

	if err_ != nil {
		fmt.Printf("bson.MarshalValue err: %+v\n", err_)

		return err_
	}

	rawValue := bson.RawValue{Type: valueType, Value: valueBytes}

	err = rawValue.Unmarshal(result)

	if err != nil {
		fmt.Printf("rawValue.Unmarshal err: %+v\n", err)

		return ErrQueryResultTypeInconsistent
	}

	return nil
}

// Cursor gets a Cursor object, which can be used to traverse the query result set
// After obtaining the CursorI object, you should actively call the Close interface to close the cursor
func (q *Query) Cursor() CursorI {
	opt := options.Find()

	if q.sort != nil {
		opt.SetSort(q.sort)
	}

	if q.project != nil {
		opt.SetProjection(q.project)
	}

	if q.limit != nil {
		opt.SetLimit(*q.limit)
	}

	if q.skip != nil {
		opt.SetSkip(*q.skip)
	}

	if q.batchSize != nil {
		opt.SetBatchSize(int32(*q.batchSize))
	}

	var err error
	var cur *mongo.Cursor

	cur, err = q.collection.Find(q.ctx, q.filter, opt)

	return &Cursor{
		ctx:    q.ctx,
		cursor: cur,
		err:    err,
	}
}

// Apply runs the findAndModify command, which allows updating, replacing
// or removing a document matching a query and atomically returning either the old
// version (the default) or the new version of the document (when ReturnNew is true)
//
// The Sort and Select query methods affect the result of Apply. In case
// multiple documents match the query, Sort enables selecting which document to
// act upon by ordering it first. Select enables retrieving only a selection
// of fields of the new or old document.
//
// When Change.Replace is true, it means replace at most one document in the collection
// and the update parameter must be a document and cannot contain any update operators;
// if no objects are found and Change.Upsert is false, it will returns ErrNoDocuments.
// When Change.Remove is true, it means delete at most one document in the collection
// and returns the document as it appeared before deletion; if no objects are found,
// it will returns ErrNoDocuments.
// When both Change.Replace and Change.Remove are false，it means update at most one document
// in the collection and the update parameter must be a document containing update operators;
// if no objects are found and Change.Upsert is false, it will returns ErrNoDocuments.
//
// reference: https://docs.mongodb.com/manual/reference/command/findAndModify/
func (q *Query) Apply(change Change, result interface{}) error {
	var err error

	if change.Remove {
		err = q.findOneAndDelete(change, result)
	} else if change.Replace {
		err = q.findOneAndReplace(change, result)
	} else {
		err = q.findOneAndUpdate(change, result)
	}

	return err
}

// findOneAndDelete
// reference: https://docs.mongodb.com/manual/reference/method/db.collection.findOneAndDelete/
func (q *Query) findOneAndDelete(change Change, result interface{}) error {
	opts := options.FindOneAndDelete()

	if q.sort != nil {
		opts.SetSort(q.sort)
	}

	if q.project != nil {
		opts.SetProjection(q.project)
	}

	return q.collection.FindOneAndDelete(q.ctx, q.filter, opts).Decode(result)
}

// findOneAndReplace
// reference: https://docs.mongodb.com/manual/reference/method/db.collection.findOneAndReplace/
func (q *Query) findOneAndReplace(change Change, result interface{}) error {
	opts := options.FindOneAndReplace()

	if q.sort != nil {
		opts.SetSort(q.sort)
	}

	if q.project != nil {
		opts.SetProjection(q.project)
	}

	if change.Upsert {
		opts.SetUpsert(change.Upsert)
	}

	if change.ReturnNew {
		opts.SetReturnDocument(options.After)
	}

	err := q.collection.FindOneAndReplace(q.ctx, q.filter, change.Update, opts).Decode(result)

	if change.Upsert && !change.ReturnNew && err == mongo.ErrNoDocuments {
		return nil
	}

	return err
}

// findOneAndUpdate
// reference: https://docs.mongodb.com/manual/reference/method/db.collection.findOneAndUpdate/
func (q *Query) findOneAndUpdate(change Change, result interface{}) error {
	opts := options.FindOneAndUpdate()

	if q.sort != nil {
		opts.SetSort(q.sort)
	}

	if q.project != nil {
		opts.SetProjection(q.project)
	}

	if change.Upsert {
		opts.SetUpsert(change.Upsert)
	}

	if change.ReturnNew {
		opts.SetReturnDocument(options.After)
	}

	err := q.collection.FindOneAndUpdate(q.ctx, q.filter, change.Update, opts).Decode(result)

	if change.Upsert && !change.ReturnNew && err == mongo.ErrNoDocuments {
		return nil
	}

	return err
}
