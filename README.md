# Godm

`Godm` is a `Go` `driver` for `MongoDB` . It is based on [MongoDB official driver](https://github.com/mongodb/mongo-go-driver), but easier to use.

- `Godm` allows users to use the new features of `MongoDB` in a more elegant way.

## Requirements

-`Go 1.18` and above.

-`MongoDB 2.6` and above.

## Features
- CRUD to documents, with all official supported options
- Sort、limit、count、select、distinct
- Transactions
- Hooks
- Automatically default and custom fields
- Predefine operator keys
- Aggregate、indexes operation、cursor
- Validation tags
- Plugin

## Installation

- Use `go mod` to automatically install dependencies by `import github.com/md-salehzadeh/godm`

Or 

- Use `go get github.com/md-salehzadeh/godm`

## Usage

- Start

    `import` and create a new connection
    ```go
    import (
        "context"
      
        "github.com/md-salehzadeh/godm"
    )
    
    ctx := context.Background()
    client, err := godm.Connect(ctx, &godm.Config{Uri: "mongodb://localhost:27017"})
    db := client.Database("class")
    coll := db.Collection("user")
    ```
    If your connection points to a fixed database and collection, recommend using the following way to initialize the connection.
    All operations can be based on `cli`:
    
    ```go
    cli, err := godm.Open(ctx, &godm.Config{Uri: "mongodb://localhost:27017", Database: "class", Coll: "user"})
    ```
    
    ***The following examples will be based on `cli`, if you use the first way for initialization, replace `cli` with `client`、`db` or `coll`***
    
    Make sure to defer a call to Disconnect after instantiating your client:
    
    ```go
    defer func() {
    if err = cli.Close(ctx); err != nil {
            panic(err)
        }
    }()
    ```

- Create index

    Before doing the operation, we first initialize some data:
    
    ```go
    type UserInfo struct {
        Name   string `bson:"name"`
        Age    uint16 `bson:"age"`
        Weight uint32 `bson:"weight"`
    }
    
    var userInfo = UserInfo{
        Name: "xm",
        Age: 7,
        Weight: 40,
    }
    ```
    
    Create index
    
    ```go
    cli.CreateOneIndex(context.Background(), options.IndexModel{Key: []string{"name"}})
    cli.CreateIndexes(context.Background(), []options.IndexModel{{Key: []string{"id2", "id3"}}})
    ```

- Insert a document

    ```go
    // insert one document
    result, err := cli.InsertOne(ctx, userInfo)
    ```

- Find a document

    ```go
    // find one document
      one := UserInfo{}
      err = cli.Find(ctx, bson.M{"name": userInfo.Name}).One(&one)
    ```

- Delete documents
    
    ```go
    err = cli.Remove(ctx, bson.M{"age": 7})
    ```

- Insert multiple data

    ```go
    // multiple insert
    var userInfos = []UserInfo{
        UserInfo{Name: "a1", Age: 6, Weight: 20},
        UserInfo{Name: "b2", Age: 6, Weight: 25},
        UserInfo{Name: "c3", Age: 6, Weight: 30},
        UserInfo{Name: "d4", Age: 6, Weight: 35},
        UserInfo{Name: "a1", Age: 7, Weight: 40},
        UserInfo{Name: "a1", Age: 8, Weight: 45},
    }
    result, err = cli.Collection.InsertMany(ctx, userInfos)
    ```

- Search all, sort and limit
    ```go
    // find all, sort and limit
    batch := []UserInfo{}
    cli.Find(ctx, bson.M{"age": 6}).Sort("weight").Limit(7).All(&batch)
    ```
- Count

    ````go
    count, err := cli.Find(ctx, bson.M{"age": 6}).Count()
    ````

- Update

    ````go
    // UpdateOne one
    err := cli.UpdateOne(ctx, bson.M{"name": "d4"}, bson.M{"$set": bson.M{"age": 7}})
    
    // UpdateAll
    result, err := cli.UpdateAll(ctx, bson.M{"age": 6}, bson.M{"$set": bson.M{"age": 10}})
    ````

- Select

    ````go
    err := cli.Find(ctx, bson.M{"age": 10}).Select(bson.M{"age": 1}).One(&one)
    ````

- Aggregate

    ```go
    matchStage := bson.D{{"$match", []bson.E{{"weight", bson.D{{"$gt", 30}}}}}}
    groupStage := bson.D{{"$group", bson.D{{"_id", "$name"}, {"total", bson.D{{"$sum", "$age"}}}}}}
    var showsWithInfo []bson.M
    err = cli.Aggregate(context.Background(), Pipeline{matchStage, groupStage}).All(&showsWithInfo)
    ```

- Support All mongoDB Options when create connection

    ````go
    poolMonitor := &event.PoolMonitor{
        Event: func(evt *event.PoolEvent) {
            switch evt.Type {
            case event.GetSucceeded:
                fmt.Println("GetSucceeded")
            case event.ConnectionReturned:
                fmt.Println("ConnectionReturned")
            }
        },
    }
    opt := options.Client().SetPoolMonitor(poolMonitor)  // more options use the chain options.
    cli, err := Open(ctx, &Config{Uri: URI, Database: DATABASE, Coll: COLL}, opt) 
    
    
    ````

- Transactions

    The super simple and powerful transaction, with features like `timeout`、`retry`:
    ````go
    callback := func(sessCtx context.Context) (interface{}, error) {
        // Important: make sure the sessCtx used in every operation in the whole transaction
        if _, err := cli.InsertOne(sessCtx, bson.D{{"abc", int32(1)}}); err != nil {
            return nil, err
        }
        if _, err := cli.InsertOne(sessCtx, bson.D{{"xyz", int32(999)}}); err != nil {
            return nil, err
        }
        return nil, nil
    }
    result, err = cli.DoTransaction(ctx, callback)
    ````
    [More about transaction](https://github.com/md-salehzadeh/godm/wiki/Transactions)

- Predefine operator keys

    ````go
    // aggregate
    matchStage := bson.D{{operator.Match, []bson.E{{"weight", bson.D{{operator.Gt, 30}}}}}}
    groupStage := bson.D{{operator.Group, bson.D{{"_id", "$name"}, {"total", bson.D{{operator.Sum, "$age"}}}}}}
    var showsWithInfo []bson.M
    err = cli.Aggregate(context.Background(), Pipeline{matchStage, groupStage}).All(&showsWithInfo)
    ````

- Hooks

    Godm flexible hooks:

    ````go
    type User struct {
        Name         string    `bson:"name"`
        Age          int       `bson:"age"`
    }
    func (u *User) BeforeInsert(ctx context.Context) error {
        fmt.Println("before insert called")
        return nil
    }
    func (u *User) AfterInsert(ctx context.Context) error {
        fmt.Println("after insert called")
        return nil
    }
    
    u := &User{Name: "Alice", Age: 7}
    _, err := cli.InsertOne(context.Background(), u)
    ````
    [More about hooks](https://github.com/md-salehzadeh/godm/wiki/Hooks)

- Automatically fields

    Godm support two ways to make specific fields automatically update in specific API
   
    - Default fields
    
    Inject `field.DefaultField` in document struct, Godm will update `createAt`、`updateAt` and `_id` in update and insert operation.
    
    ````go
    type User struct {
      field.DefaultField `bson:",inline"`
    
      Name string `bson:"name"`
      Age  int    `bson:"age"`
    }
  
  	u := &User{Name: "Lucas", Age: 7}
  	_, err := cli.InsertOne(context.Background(), u)
    // Fields with tag createAt、updateAt and _id will be generated automatically 
    ```` 

    - Custom fields
    
    Define the custom fields, Godm will update them in update and insert operation.

    ```go
    type User struct {
        Name string `bson:"name"`
        Age  int    `bson:"age"`
    
        MyId         string    `bson:"myId"`
        CreateTimeAt time.Time `bson:"createTimeAt"`
        UpdateTimeAt int64     `bson:"updateTimeAt"`
    }
    // Define the custom fields
    func (u *User) CustomFields() field.CustomFieldsBuilder {
        return field.NewCustom().SetCreateAt("CreateTimeAt").SetUpdateAt("UpdateTimeAt").SetId("MyId")
    }
  
    u := &User{Name: "Lucas", Age: 7}
    _, err := cli.InsertOne(context.Background(), u)
    // CreateTimeAt、UpdateTimeAt and MyId will be generated automatically 
  
    // suppose Id and ui is ready
  	err = cli.ReplaceOne(context.Background(), bson.M{"_id": Id}, &ui)
    // UpdateTimeAt will update
    ```

    Check [examples here](https://github.com/md-salehzadeh/godm/blob/master/field_test.go)

    [More about automatically fields](https://github.com/md-salehzadeh/godm/wiki/Automatically-update-fields)

- Validation tags

    Godm Validation tags is Based on [go-playground/validator](https://github.com/go-playground/validator).
    
    So Godm support [all validations on structs in go-playground/validator](https://github.com/go-playground/validator#usage-and-documentation), such as:
    
    ```go
    type User struct {
        FirstName string            `bson:"fname"`
        LastName  string            `bson:"lname"`
        Age       uint8             `bson:"age" validate:"gte=0,lte=130" `    // Age must in [0,130]
        Email     string            `bson:"e-mail" validate:"required,email"` //  Email can't be empty string, and must has email format
        CreateAt  time.Time         `bson:"createAt" validate:"lte"`          // CreateAt must lte than current time
        Relations map[string]string `bson:"relations" validate:"max=2"`       // Relations can't has more than 2 elements
    }
    ```
    
    Godm tags only supported in following API：
    ` InsertOne、InsertyMany、Upsert、UpsertId、ReplaceOne `

- Plugin
    
    - Implement following method:
    
    ```go
    func Do(ctx context.Context, doc interface{}, opType operator.OpType, opts ...interface{}) error{
      // do anything
    }
    ```
    
    - Call Register() in package middleware, register the method `Do`
    
      Godm will call `Do` before and after the [operation](operator/operate_type.go)
      
    ```go
    middleware.Register(Do)
    ```
    [Example](middleware/middleware_test.go)
    
    The `hook`、`automatically fields` and `validation tags` in Godm run on **plugin**.
    
## `Godm` vs `go.mongodb.org/mongo-driver`

Below we give an example of multi-file search、sort and limit to illustrate the improvement compare to `go.mongodb.org/mongo-driver`.
How do we do in`go.mongodb.org/mongo-driver`:

```go
// go.mongodb.org/mongo-driver
// find all, sort and limit
findOptions := options.Find()
findOptions.SetLimit(7) // set limit
var sorts D
sorts = append(sorts, E{Key: "weight", Value: 1})
findOptions.SetSort(sorts) // set sort

batch := []UserInfo{}
cur, err := coll.Find(ctx, bson.M{"age": 6}, findOptions)
cur.All(ctx, &batch)
```

How do we do in `godm`:

```go
// godm
// find all, sort and limit
batch := []UserInfo{}
cli.Find(ctx, bson.M{"age": 6}).Sort("weight").Limit(7).All(&batch)
```