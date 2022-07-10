package godm

import (
	"fmt"
	"reflect"
	"strings"

	gOpts "github.com/md-salehzadeh/godm/options"
)

type Model struct {
	connection *Connection
	collection *Collection
	document   interface{}
}

func (c *Connection) RegisterModel(document interface{}, collName string) {
	if document == nil {
		panic("document can not be nil")
	}

	reflectType := reflect.TypeOf(document)

	typeName := strings.ToLower(reflectType.Elem().Name())

	if _, ok := c.modelRegistry[typeName]; !ok {
		collection := c.Database(c.Config.Database).Collection(collName)

		model := &Model{
			connection: c,
			collection: collection,
			document:   document,
		}

		c.modelRegistry[typeName] = model
		c.typeRegistry[typeName] = reflectType.Elem()
	} else {
		fmt.Printf("Tried to register model '%v' twice\n", typeName)
	}
}

func (c *Connection) Model(name string) *Model {
	_name := strings.ToLower(name)

	if _, ok := c.modelRegistry[_name]; ok {
		return c.modelRegistry[_name]
	}

	panic(fmt.Sprintf("DB: Model '%v' is not registered", name))
}

func (m *Model) Find(opts ...gOpts.FindOptions) QueryI {
	query := m.collection.Find(opts...)

	query.setDocument(m.document)

	return query
}
