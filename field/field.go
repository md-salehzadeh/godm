package field

import (
	"context"
	"reflect"
	"time"

	"github.com/md-salehzadeh/godm/operator"
)

var nilTime time.Time

// filedHandler defines the relations between field type and handler
var fieldHandler = map[operator.OpType]func(doc interface{}) error{
	operator.BeforeInsert:  beforeInsert,
	operator.BeforeUpdate:  beforeUpdate,
	operator.BeforeReplace: beforeUpdate,
	operator.BeforeUpsert:  beforeUpsert,
}

//func init() {
//	middleware.Register(Do)
//}

// Do call the specific method to handle field based on fType
// Don't use opts here
func Do(ctx context.Context, doc interface{}, opType operator.OpType, opts ...interface{}) error {
	to := reflect.TypeOf(doc)
	if to == nil {
		return nil
	}
	switch reflect.TypeOf(doc).Kind() {
	case reflect.Slice:
		return sliceHandle(doc, opType)
	case reflect.Ptr:
		v := reflect.ValueOf(doc).Elem()
		switch v.Kind() {
		case reflect.Slice:
			return sliceHandle(v.Interface(), opType)
		default:
			return do(doc, opType)
		}
	}
	//fmt.Println("not support type")
	return nil
}

// sliceHandle handles the slice docs
func sliceHandle(docs interface{}, opType operator.OpType) error {
	// []interface{}{UserType{}...}
	if h, ok := docs.([]interface{}); ok {
		for _, v := range h {
			if err := do(v, opType); err != nil {
				return err
			}
		}
		return nil
	}
	// []UserType{}
	s := reflect.ValueOf(docs)
	for i := 0; i < s.Len(); i++ {
		if err := do(s.Index(i).Interface(), opType); err != nil {
			return err
		}
	}
	return nil
}

// beforeInsert handles field before insert
// If value of field createAt is valid in doc, upsert doesn't change it
// If value of field id is valid in doc, upsert doesn't change it
// Change the value of field updateAt anyway
func beforeInsert(doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultId()
		ih.DefaultCreateAt()
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomId(doc)
		fields.(*CustomFields).CustomCreateTime(doc)
		fields.(*CustomFields).CustomUpdateTime(doc)
	}
	return nil
}

// beforeUpdate handles field before update
func beforeUpdate(doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomUpdateTime(doc)
	}
	return nil
}

// beforeUpsert handles field before upsert
// If value of field createAt is valid in doc, upsert doesn't change it
// If value of field id is valid in doc, upsert doesn't change it
// Change the value of field updateAt anyway
func beforeUpsert(doc interface{}) error {
	if ih, ok := doc.(DefaultFieldHook); ok {
		ih.DefaultId()
		ih.DefaultCreateAt()
		ih.DefaultUpdateAt()
	}
	if ih, ok := doc.(CustomFieldsHook); ok {
		fields := ih.CustomFields()
		fields.(*CustomFields).CustomId(doc)
		fields.(*CustomFields).CustomCreateTime(doc)
		fields.(*CustomFields).CustomUpdateTime(doc)
	}
	return nil
}

// do check if opType is supported and call fieldHandler
func do(doc interface{}, opType operator.OpType) error {
	if f, ok := fieldHandler[opType]; !ok {
		return nil
	} else {
		return f(doc)
	}
}
