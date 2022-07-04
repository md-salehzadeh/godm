package field

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DefaultFieldHook defines the interface to change default fields by hook
type DefaultFieldHook interface {
	DefaultUpdateAt()
	DefaultCreateAt()
	DefaultId()
}

// DefaultField defines the default fields to handle when operation happens
// import the DefaultField in document struct to make it working
type DefaultField struct {
	Id       primitive.ObjectID `bson:"_id"`
	CreateAt time.Time          `bson:"createAt"`
	UpdateAt time.Time          `bson:"updateAt"`
}

// DefaultUpdateAt changes the default updateAt field
func (df *DefaultField) DefaultUpdateAt() {
	df.UpdateAt = time.Now().Local()
}

// DefaultCreateAt changes the default createAt field
func (df *DefaultField) DefaultCreateAt() {
	if df.CreateAt.IsZero() {
		df.CreateAt = time.Now().Local()
	}
}

// DefaultCreateAt changes the default _id field
func (df *DefaultField) DefaultId() {
	if df.Id.IsZero() {
		df.Id = primitive.NewObjectID()
	}
}
