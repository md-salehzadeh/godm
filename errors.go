package godm

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrQueryNotSlicePointer return if result argument is not a pointer to a slice
	ErrQueryNotSlicePointer = errors.New("result argument must be a pointer to a slice")
	// ErrQueryNotSliceType return if result argument is not slice address
	ErrQueryNotSliceType = errors.New("result argument must be a slice address")
	// ErrQueryResultTypeInconsistent return if result type is not equal mongodb value type
	ErrQueryResultTypeInconsistent = errors.New("result type is not equal mongodb value type")
	// ErrQueryResultValCanNotChange return if the value of result can not be changed
	ErrQueryResultValCanNotChange = errors.New("the value of result can not be changed")
	// ErrNoSuchDocuments return if no document found
	ErrNoSuchDocuments = mongo.ErrNoDocuments
	// ErrTransactionRetry return if transaction need to retry
	ErrTransactionRetry = errors.New("retry transaction")
	// ErrTransactionNotSupported return if transaction not supported
	ErrTransactionNotSupported = errors.New("transaction not supported")
	// ErrNotSupportedUsername return if username is invalid
	ErrNotSupportedUsername = errors.New("username not supported")
	// ErrNotSupportedPassword return if password is invalid
	ErrNotSupportedPassword = errors.New("password not supported")
	// ErrNotValidSliceToInsert return if insert argument is not valid slice
	ErrNotValidSliceToInsert = errors.New("must be valid slice to insert")
	// ErrReplacementContainUpdateOperators return if replacement document contain update operators
	ErrReplacementContainUpdateOperators = errors.New("replacement document cannot contain keys beginning with '$'")
)

// IsErrNoDocuments check if err is no documents, both mongo-go-driver error and godm custom error
// Deprecated, simply call if err == ErrNoSuchDocuments or if err == mongo.ErrNoDocuments
func IsErrNoDocuments(err error) bool {
	return err == ErrNoSuchDocuments
}

// IsDup check if err is mongo E11000 (duplicate err)。
func IsDup(err error) bool {
	return err != nil && strings.Contains(err.Error(), "E11000")
}
