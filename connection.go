package godm

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/md-salehzadeh/godm/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	opts "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config for initial mongodb instance
type Config struct {
	// URI example: [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// URI Reference: https://docs.mongodb.com/manual/reference/connection-string/
	Uri      string `json:"uri"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Coll     string `json:"coll"`
	// ConnectTimeoutMS specifies a timeout that is used for creating connections to the server.
	//	If set to 0, no timeout will be used.
	//	The default is 30 seconds.
	ConnectTimeoutMS *int64 `json:"connectTimeoutMS"`
	// MaxPoolSize specifies that maximum number of connections allowed in the driver's connection pool to each server.
	// If this is 0, it will be set to math.MaxInt64,
	// The default is 100.
	MaxPoolSize *uint64 `json:"maxPoolSize"`
	// MinPoolSize specifies the minimum number of connections allowed in the driver's connection pool to each server. If
	// this is non-zero, each server's pool will be maintained in the background to ensure that the size does not fall below
	// the minimum. This can also be set through the "minPoolSize" URI option (e.g. "minPoolSize=100"). The default is 0.
	MinPoolSize *uint64 `json:"minPoolSize"`
	// SocketTimeoutMS specifies how long the driver will wait for a socket read or write to return before returning a
	// network error. If this is 0 meaning no timeout is used and socket operations can block indefinitely.
	// The default is 300,000 ms.
	SocketTimeoutMS *int64 `json:"socketTimeoutMS"`
	// ReadPreference determines which servers are considered suitable for read operations.
	// default is PrimaryMode
	ReadPreference *ReadPref `json:"readPreference"`
	// can be used to provide authentication options when configuring a Client.
	Auth *Credential `json:"auth"`
}

// Credential can be used to provide authentication options when configuring a Client.
//
// AuthMechanism: the mechanism to use for authentication. Supported values include "SCRAM-SHA-256", "SCRAM-SHA-1",
// "MONGODB-CR", "PLAIN", "GSSAPI", "MONGODB-X509", and "MONGODB-AWS". This can also be set through the "authMechanism"
// URI option. (e.g. "authMechanism=PLAIN"). For more information, see
// https://docs.mongodb.com/manual/core/authentication-mechanisms/.
// AuthSource: the name of the database to use for authentication. This defaults to "$external" for MONGODB-X509,
// GSSAPI, and PLAIN and "admin" for all other mechanisms. This can also be set through the "authSource" URI option
// (e.g. "authSource=otherDb").
//
// Username: the username for authentication. This can also be set through the URI as a username:password pair before
// the first @ character. For example, a URI for user "user", password "pwd", and host "localhost:27017" would be
// "mongodb://user:pwd@localhost:27017". This is optional for X509 authentication and will be extracted from the
// client certificate if not specified.
//
// Password: the password for authentication. This must not be specified for X509 and is optional for GSSAPI
// authentication.
//
// PasswordSet: For GSSAPI, this must be true if a password is specified, even if the password is the empty string, and
// false if no password is specified, indicating that the password should be taken from the context of the running
// process. For other mechanisms, this field is ignored.
type Credential struct {
	AuthMechanism string `json:"authMechanism"`
	AuthSource    string `json:"authSource"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	PasswordSet   bool   `json:"passwordSet"`
}

// ReadPref determines which servers are considered suitable for read operations.
type ReadPref struct {
	// MaxStaleness is the maximum amount of time to allow a server to be considered eligible for selection.
	// Supported from version 3.4.
	MaxStalenessMS int64 `json:"maxStalenessMS"`
	// indicates the user's preference on reads.
	// PrimaryMode as default
	Mode readpref.Mode `json:"mode"`
}

// Client creates client to mongo
type Connection struct {
	Client *mongo.Client
	Config Config

	registry      *bsoncodec.Registry
	modelRegistry map[string]*Model
	typeRegistry  map[string]reflect.Type
}

// Connect creates Godm MongoDB Connection
func Connect(ctx context.Context, conf *Config, _opts ...options.ClientOptions) (*Connection, error) {
	options, err := newConnectOpts(conf, _opts...)

	if err != nil {
		return nil, err
	}

	client, err := client(ctx, options)

	if err != nil {
		return nil, err
	}

	connection := &Connection{
		Client:        client,
		Config:        *conf,
		registry:      options.Registry,
		modelRegistry: make(map[string]*Model),
		typeRegistry:  make(map[string]reflect.Type),
	}

	return connection, nil
}

// creates connection to MongoDB
func client(ctx context.Context, opts *opts.ClientOptions) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, opts)

	if err != nil {
		return nil, err
	}

	// half of default connect timeout
	pCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

	defer cancel()

	if err = client.Ping(pCtx, readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}

// newConnectOpts creates client options from conf
// Godm will follow this way official mongodb driver do：
// - the configuration in uri takes precedence over the configuration in the setter
// - Check the validity of the configuration in the uri, while the configuration in the setter is basically not checked
func newConnectOpts(conf *Config, _opts ...options.ClientOptions) (*opts.ClientOptions, error) {
	options := opts.Client()

	for _, apply := range _opts {
		options = opts.MergeClientOptions(apply.ClientOptions)
	}

	if conf.ConnectTimeoutMS != nil {
		timeoutDur := time.Duration(*conf.ConnectTimeoutMS) * time.Millisecond

		options.SetConnectTimeout(timeoutDur)
	}

	if conf.SocketTimeoutMS != nil {
		timeoutDur := time.Duration(*conf.SocketTimeoutMS) * time.Millisecond

		options.SetSocketTimeout(timeoutDur)
	} else {
		options.SetSocketTimeout(300 * time.Second)
	}

	if conf.MaxPoolSize != nil {
		options.SetMaxPoolSize(*conf.MaxPoolSize)
	}

	if conf.MinPoolSize != nil {
		options.SetMinPoolSize(*conf.MinPoolSize)
	}

	if conf.ReadPreference != nil {
		readPreference, err := newReadPref(*conf.ReadPreference)

		if err != nil {
			return nil, err
		}

		options.SetReadPreference(readPreference)
	}

	if conf.Auth != nil {
		auth, err := newAuth(*conf.Auth)

		if err != nil {
			return nil, err
		}

		options.SetAuth(auth)
	}

	uri := conf.Uri

	if uri == "" {
		uri = fmt.Sprintf("mongodb://%s:%s", conf.Host, conf.Port)
	}

	options.ApplyURI(uri)

	return options, nil
}

// creates options.Credential from conf.Auth
func newAuth(auth Credential) (credential opts.Credential, err error) {
	if auth.AuthMechanism != "" {
		credential.AuthMechanism = auth.AuthMechanism
	}

	if auth.AuthSource != "" {
		credential.AuthSource = auth.AuthSource
	}

	if auth.Username != "" {
		// Validate and process the username.
		if strings.Contains(auth.Username, "/") {
			err = ErrNotSupportedUsername

			return
		}

		credential.Username, err = url.QueryUnescape(auth.Username)

		if err != nil {
			err = ErrNotSupportedUsername

			return
		}
	}

	credential.PasswordSet = auth.PasswordSet

	if auth.Password != "" {
		if strings.Contains(auth.Password, ":") {
			err = ErrNotSupportedPassword

			return
		}

		if strings.Contains(auth.Password, "/") {
			err = ErrNotSupportedPassword

			return
		}

		credential.Password, err = url.QueryUnescape(auth.Password)

		if err != nil {
			err = ErrNotSupportedPassword

			return
		}

		credential.Password = auth.Password
	}

	return
}

// creates readpref.ReadPref from config
func newReadPref(pref ReadPref) (*readpref.ReadPref, error) {
	readPrefOpts := make([]readpref.Option, 0, 1)

	if pref.MaxStalenessMS != 0 {
		readPrefOpts = append(readPrefOpts, readpref.WithMaxStaleness(time.Duration(pref.MaxStalenessMS)*time.Millisecond))
	}

	mode := readpref.PrimaryMode

	if pref.Mode != 0 {
		mode = pref.Mode
	}

	readPreference, err := readpref.New(mode, readPrefOpts...)

	return readPreference, err
}

// closes sockets to the topology referenced by this Client.
func (c *Connection) Close(ctx context.Context) error {
	err := c.Client.Disconnect(ctx)

	return err
}

// confirms connection is alive
func (c *Connection) Ping(timeout int64) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	defer cancel()

	if err = c.Client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	return nil
}

// creates connection to database
func (c *Connection) Database(name string, options ...*options.DatabaseOptions) *Database {
	opts := opts.Database()

	if len(options) > 0 {
		if options[0].DatabaseOptions != nil {
			opts = options[0].DatabaseOptions
		}
	}

	return &Database{database: c.Client.Database(name, opts), registry: c.registry}
}

// creates one session on client
// Watch out, close session after operation done
func (c *Connection) Session(_opts ...*options.SessionOptions) (*Session, error) {
	sessionOpts := opts.Session()

	if len(_opts) > 0 && _opts[0].SessionOptions != nil {
		sessionOpts = _opts[0].SessionOptions
	}

	s, err := c.Client.StartSession(sessionOpts)

	return &Session{session: s}, err
}

// DoTransaction do whole transaction in one function
// precondition：
// - version of mongoDB server >= v4.0
// - Topology of mongoDB server is not Single
// At the same time, please pay attention to the following
// - make sure all operations in callback use the sessCtx as context parameter
// - if operations in callback takes more than(include equal) 120s, the operations will not take effect,
// - if operation in callback return godm.ErrTransactionRetry,
//   the whole transaction will retry, so this transaction must be idempotent
// - if operations in callback return godm.ErrTransactionNotSupported,
// - If the ctx parameter already has a Session attached to it, it will be replaced by this session.
func (c *Connection) DoTransaction(ctx context.Context, callback func(sessCtx context.Context) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
	if !c.transactionAllowed() {
		return nil, ErrTransactionNotSupported
	}

	s, err := c.Session()

	if err != nil {
		return nil, err
	}

	defer s.EndSession(ctx)

	return s.StartTransaction(ctx, callback, opts...)
}

// gets the version of mongoDB server, like 4.4.0
func (c *Connection) ServerVersion() string {
	var buildInfo bson.Raw

	err := c.Client.Database("admin").RunCommand(context.Background(), bson.D{{"buildInfo", 1}}).Decode(&buildInfo)

	if err != nil {
		fmt.Println("run command err", err)

		return ""
	}

	v, err := buildInfo.LookupErr("version")

	if err != nil {
		fmt.Println("look up err", err)

		return ""
	}

	return v.StringValue()
}

// transactionAllowed check if transaction is allowed
func (c *Connection) transactionAllowed() bool {
	vr, err := CompareVersions("4.0", c.ServerVersion())

	if err != nil {
		return false
	}

	if vr > 0 {
		fmt.Println("transaction is not supported because mongo server version is below 4.0")
		return false
	}

	// TODO dont know why need to do `cli, err := Open(ctx, &c.conf)` in topology() to get topo,
	// Before figure it out, we only use this function in UT
	//topo, err := c.topology()
	//if topo == description.Single {
	//	fmt.Println("transaction is not supported because mongo server topology is single")
	//	return false
	//}

	return true
}
