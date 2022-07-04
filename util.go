package godm

import (
	"math"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Now return Millisecond current time
func Now() time.Time {
	return time.Unix(0, time.Now().UnixNano()/1e6*1e6)
}

// NewObjectID generates a new ObjectID.
func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

// handles sort symbol: "asc"/"desc" at the end of field
// if "asc"， return sort as 1
// if "desc"， return sort as -1
func ParseSortField(field string) (key string, sort int32) {
	sort = 1
	key = field

	if len(field) != 0 {
		splittedField := strings.Split(field, " ")

		if len(splittedField) == 2 && strings.ToLower(splittedField[1]) == "desc" {
			sort = -1
		}

		key = splittedField[0]
	}

	return key, sort
}

// handles select symbol
// if field has "!" at the beginning, its translated to -1 otherwise to 1
func ParseSelectField(field string) (key string, visible int32) {
	key = field
	visible = 1

	if len(field) != 0 && strings.HasPrefix(field, "!") {
		key = strings.Replace(field, "!", "", -1)

		visible = 0
	}

	return key, visible
}

// CompareVersions compares two version number strings (i.e. positive integers separated by
// periods). Comparisons are done to the lesser precision of the two versions. For example, 3.2 is
// considered equal to 3.2.11, whereas 3.2.0 is considered less than 3.2.11.
//
// Returns a positive int if version1 is greater than version2, a negative int if version1 is less
// than version2, and 0 if version1 is equal to version2.
func CompareVersions(v1 string, v2 string) (int, error) {
	n1 := strings.Split(v1, ".")
	n2 := strings.Split(v2, ".")

	for i := 0; i < int(math.Min(float64(len(n1)), float64(len(n2)))); i++ {
		i1, err := strconv.Atoi(n1[i])

		if err != nil {
			return 0, err
		}

		i2, err := strconv.Atoi(n2[i])

		if err != nil {
			return 0, err
		}

		difference := i1 - i2

		if difference != 0 {
			return difference, nil
		}
	}

	return 0, nil
}
