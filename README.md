# GoModel ![GitHub release (latest by date)](https://img.shields.io/github/v/release/moiseshiraldo/gomodel) [![Build Status](https://travis-ci.org/moiseshiraldo/gomodel.svg?branch=master)](https://travis-ci.org/moiseshiraldo/gomodel) [![Test coverage status](https://codecov.io/gh/moiseshiraldo/gomodel/branch/master/graph/badge.svg)](https://codecov.io/gh/moiseshiraldo/gomodel) [![GoDoc](https://godoc.org/github.com/moiseshiraldo/gomodel?status.svg)](https://godoc.org/github.com/jinzhu/gorm)

GoModel is an experimental project aiming to implement the features offered by the Python Django ORM using Go.

Please notice that the project is on early development and so the public API is
likely to change.

1. [**Quick start**](#quick-start)
2. [**Definitions**](#definitions)
   - [Applications](#applications)
   - [Models](#models)
   - [Databases](#databases)
   - [Containers](#containers)
   - [Fields](#fields)
3. [**Schema migrations**](./migration)
4. [**Making queries**](#making-queries)
   - [CRUD](#crud)
   - [Managers](#managers)
   - [QuerySets](#querysets)
   - [Conditioners](#conditioners)
   - [Multiple databases](#multiple-databases)
   - [Transactions](#transactions)
6. [**Testing**](#testing)
7. [**Benchmarks**](./benchmark)


# Quick start

```go
package main

import (
    "fmt"
    _ "github.com/lib/pq"  // Imports database driver.
    "github.com/moiseshiraldo/gomodel"
    "github.com/moiseshiraldo/gomodel/migration"
    "time"
)

// This is how you define models.
var User = gomodel.New(
    "User",
    gomodel.Fields{
        "email":   gomodel.CharField{MaxLength: 100, Index: true},
        "active":  gomodel.BooleanField{DefaultFalse: true},
        "created": gomodel.DateTimeField{AutoNowAdd: true},
    },
    gomodel.Options{},
)

// Models are grouped inside applications.
var app = gomodel.NewApp("main", "/home/project/main/migrations", User.Model)

func setup() {
    // You have to register an application to be able to use its models.
    gomodel.Register(app)
    // And finally open at least a default database connection.
    gomodel.Start(map[string]gomodel.Database{
        "default": {
            Driver:   "postgres",
            Name:     "test",
            User:     "local",
            Password: "local",
        },
    })
}

func checkError(err error) {
    if err != nil {
        panic(err)
    }
}

func main() {
    // Let's create a user.
    user, err := User.Objects.Create(gomodel.Values{"email": "user@test.com"})
    // You'll probably get an error if the users table doesn't exist in the
    // database yet. Check out the migration package for more information!
    checkError(err)

    if _, ok := user.GetIf("forename"); !ok {
        fmt.Println("That field doesn't exist!")
    }
    // But we know this one does and can't be null.
    created := user.Get("created").(time.Time)
    fmt.Printf("This user was created on year %d", created.Year())

    // Do we have any active ones?
    exists, err := User.Objects.Filter(gomodel.Q{"active": true}).Exists()
    checkError(err)
    if !exists {
        // It doesn't seem so, but we can change that.
        user.Set("active", true)
        err := user.Save()
        checkError(err)
    }
    
    // What about now?
    count, err := User.Objects.Filter(gomodel.Q{"active": true}).Count()
    checkError(err)
    fmt.Println("We have %d active users!", count)
    
    // Let's create another one!
    data := struct {
        Email  string
        Active bool
    } {"admin@test.com", true}
    _, err := User.Objects.Create(data)
    checkError(err)
    
    // And print some details about them.
    users, err := User.Objects.All().Load()
    for _, user := range users {
        fmt.Println(
            user.Display("email"), "was created at", user.Display("created"),
        )
    }
    
    // I wonder what happens if we try to get a random one...
    user, err = User.Objects.Get(gomodel.Q{"id": 17})
    if _, ok := err.(*gomodel.ObjectNotFoundError); ok {
        fmt.Println("Keep trying!")
    }
    
    // Enough for today, let's deactivate all the users.
    n, err := User.Objects.Filter(gomodel.Q{"active": true}).Update(
        gomodel.Values{"active": false}
    )
    checkError(err)
    fmt.Printf("%d users have been deactivated\n", n)
    
    // Or even better, kill 'em a... I mean delete them all.
    _, err := User.Objects.All().Delete()
    checkError(err)
}
```

# Definitions

## Applications

An application represents a group of models that share something in common
(a feature, a package...), making it easier to export and reuse them. You can
create application settings using the [NewApp](https://godoc.org/github.com/moiseshiraldo/gomodel/#NewApp)
function:

```go
var app = gomodel.NewApp("main", "/home/project/main/migrations", models...)
```

The first argument is the name of the application, which must be unique. The
second one is the migrations path (check the [migration](./migration) package
for more details), followed by the list of models belonging to the application.

Application settings can be registered with the [Register](https://godoc.org/github.com/moiseshiraldo/gomodel/#Register)
function, that will validate the app details and the models, panicking on
any definition error. A map of registered applications can be obtained calling
[Registry](https://godoc.org/github.com/moiseshiraldo/gomodel/#Registry).

## Models

A model represents a source of data inside an application, usually mapping to a
database table. Models can be created using the [New](https://godoc.org/github.com/moiseshiraldo/gomodel/#New)
function:

```go
var User = gomodel.New(
    "User",
    gomodel.Fields{
        "email": gomodel.CharField{MaxLength: 100, Index: true},
        "active": gomodel.BooleanField{DefaultFalse: true},
        "created": gomodel.DateTimeField{AutoNowAdd: true},
    },
    gomodel.Options{},
)
```

The first argument is the name of the model, which must be unique inside the
application. The second one is the map of [fields](#fields) and the last one the
model [Options](https://godoc.org/github.com/moiseshiraldo/gomodel/#Options).
The function returns a [Dispatcher](https://godoc.org/github.com/moiseshiraldo/gomodel/#Dispatcher)
giving access to the model and the default Objects [manager](#managers).

Please notice that the model must be registered to an application before making
any queries.

## Databases

A [Database](https://godoc.org/github.com/moiseshiraldo/gomodel/#Database)
represents a single organized collection of structured information.

GoModel offers database-abstraction API that lets you create, retrieve, update
and delete objects. The underlying communication is done via the [database/sql](https://golang.org/pkg/database/sql/)
package, so the corresponding [driver](https://github.com/golang/go/wiki/SQLDrivers)
must be imported. The bridge between the API and the the sql package is
constructed implementing the [Engine](https://godoc.org/github.com/moiseshiraldo/gomodel/#Engine)
interface.

At the moment, there are engines available for the `postgres` and the `sqlite3`
drivers, as well as a `mocker` one that can be used for [unit testing](#testing).

Once the [Start](https://godoc.org/github.com/moiseshiraldo/gomodel/#Start)
function has been called, the [Databases](https://godoc.org/github.com/moiseshiraldo/gomodel/#Databases)
function can be used to get a map with all the available databases.

## Containers

A container is just a Go variable where the data for a specific model instance
is stored. It can be a `struct` or any type implementing the [Builder](https://godoc.org/github.com/moiseshiraldo/gomodel/#Builder)
interface. By default, a model will use the [Values](https://godoc.org/github.com/moiseshiraldo/gomodel/#Values)
map to store data. That can be changed passing another container to the model
definition options:

```go
type userCont struct {
    email   string
    active  bool
    created time.Time
}

var User = gomodel.New(
    "User",
    gomodel.Fields{
        "email": gomodel.CharField{MaxLength: 100, Index: true},
        "active": gomodel.BooleanField{DefaultFalse: true},
        "created": gomodel.DateTimeField{AutoNowAdd: true},
    },
    gomodel.Options{
        Container: userCont{},
    },
)
```

A different container can also be set for specific queries:

```go
qs := User.Objects.Filter(gomodel.Q{"active": true}).WithContainer(userCont{})
users, err := qs.Load()
```

## Fields

Click a field name to see the documentation with all the options.

Recipient is the type used to store values on the default map container. Null
Recipient is the type used when the column can be Null. Value is the returned
type when any instance get method is called (`nil` for Null) for any of the
underlying recipients of the field.

| Name                                                                               |  Recipient         | Null Recipient      | Value                         |
|------------------------------------------------------------------------------------|:------------------:|:-------------------:|:-----------------------------:|
| [IntegerField](https://godoc.org/github.com/moiseshiraldo/gomodel/#IntegerField)   | `int32`            | `gomodel.NullInt32` | `int32`                       |
| [CharField](https://godoc.org/github.com/moiseshiraldo/gomodel/#CharField)         | `string`           | `sql.NullString`    | `string`                      |
| [BooleanField](https://godoc.org/github.com/moiseshiraldo/gomodel/#BooleanField)   | `bool`             | `sql.NullBool`      | `bool`                        |
| [DateField](https://godoc.org/github.com/moiseshiraldo/gomodel/#DateField)         | `gomodel.NullTime` | `gomodel.NullTime`  | `time.Time`                   |
| [TimeField](https://godoc.org/github.com/moiseshiraldo/gomodel/#TimeField)         | `gomodel.NullTime` | `gomodel.NullTime`  | `time.Time`                   |
| [DateTimeField](https://godoc.org/github.com/moiseshiraldo/gomodel/#DateTimeField) | `gomodel.NullTime` | `gomodel.NullTime`  | `time.Time`                   |

# Making queries

## CRUD

| Operation    | Single object                               | Multiple objects                                           |
|--------------|---------------------------------------------|------------------------------------------------------------|
| Create       | `user, err := User.Objects.Create(values)`  | Not supported yet                                          |
| Read         | `user, err := User.Objects.Get(conditions)` | `users, err := User.Objects.Filter(conditions).Load()`     |
| Update       | `err := user.Save()`                        | `n, err := User.Objects.Filter(conditions).Update(values)` |
| Delete       | `err := user.Delete()`                      | `n, err := User.Objects.Filter(conditions).Delete()`       |

## Managers

A model [Manager](https://godoc.org/github.com/moiseshiraldo/gomodel/#Manager)
provides access to the database abstraction API that lets you perform CRUD
operations.

By default, a [Dispatcher](https://godoc.org/github.com/moiseshiraldo/gomodel/#Manager)
provides access to the model manager through the `Objects` field. But you can
define a custom model dispatcher with additional managers:
 
```go
type activeManager {
    gomodel.Manager
}

// GetQuerySet overrides the default Manager method to return only active users.
func (m activeManager) GetQuerySet() QuerySet {
	return m.Manager.GetQuerySet().Filter(gomodel.Q{"active": true})
}

// Create overrides the default Manager method to set a created user as active.
func (m activeManager) Create(vals gomodel.Container) (*gomodel.Instance, error) {
    user, err := m.Manager.Create(vals)
    if err := nil {
        return user, err
    }
    user.Set("active", true)
    err = user.Save("active")
    return user, error
}

type customUserDispatcher struct {
    gomodel.Dispatcher
    Active activeManager
}

var userDispatcher = gomodel.New(
    "User",
    gomodel.Fields{
        "email": gomodel.CharField{MaxLength: 100, Index: true},
        "active": gomodel.BooleanField{DefaultFalse: true},
        "created": gomodel.DateTimeField{AutoNowAdd: true},
    },
    gomodel.Options{},
)

var User = customUserDispatcher{
    Dispatcher: userDispatcher,
    Active: activeUsersManager{userDispatcher.Objects},
}
```

And they can be accessed like the default manager:

```go
user, err := User.Active.Create(gomodel.Values{"email": "user@test.com"})
user, err := User.Active.Get("email": "user@test.com")
```

## QuerySets

A [QuerySet](https://godoc.org/github.com/moiseshiraldo/gomodel/#QuerySet) is
an interface that represents a collection of objects on the database and the
methods to interact with them.

The default manager returns a [GenericQuerySet](https://godoc.org/github.com/moiseshiraldo/gomodel/#QuerySet),
but you can define custom querysets with additional methods:

```go
type UserQS struct {
    gomodel.GenericQuerySet
}

func (qs UserQuerySet) Adults() QuerySet {
    return qs.Filter(gomodel.Q{"dob <=": time.Now().AddDate(-18, 0, 0)})
}

type customUserDispatcher struct {
    gomodel.Dispatcher
    Objects Manager
}

var userDispatcher = gomodel.New(
    "User",
    gomodel.Fields{
        "email": gomodel.CharField{MaxLength: 100, Index: true},
        "active": gomodel.BooleanField{DefaultFalse: true},
        "dob": gomodel.DateField{},
    },
    gomodel.Options{},
)

var User = customUserDispatcher{
    Dispatcher: userDispatcher,
    Objects: Manager{userDispatcher.Model, UserQS{}},
}
```

Notice that you will have to cast the queryset to access the custom method:

```go
qs := User.Objects.Filter(gomodel.Q{"active": true}).(UserQS).Adults()
activeAdults, err := qs.Load()
```

## Conditioners

Most of the manager and queryset methods receive a [Conditioner](https://godoc.org/github.com/moiseshiraldo/gomodel/#Conditioner)
as an argument, which is just an interface that represents SQL predicates and
the methods to combine them.

The [Q](https://godoc.org/github.com/moiseshiraldo/gomodel/#Q) type is the
default implementation of the interface. A `Q` is just a map of values where
the key is the column and operator part of the condition, separated by a blank
space. The equal operator can be omitted:

```go
qs := User.Objects.Filter(gomodel.Q{"active": true})
```

At the moment, only the simple comparison operators (`=`, `>`, `<`, `>=`, `<=`)
are supported. You can check if a column is `Null` using the equal operator and
passing the `nil` value.

Complex predicates can be constructed programmatically using the
[And](https://godoc.org/github.com/moiseshiraldo/gomodel/#Q.And),
[AndNot](https://godoc.org/github.com/moiseshiraldo/gomodel/#Q.AndNot),
[Or](https://godoc.org/github.com/moiseshiraldo/gomodel/#Q.Or), and
[OrNot](https://godoc.org/github.com/moiseshiraldo/gomodel/#Q.OrNot) methods:

```go
conditions := gomodel.Q{"active": true}.AndNot(
    gomodel.Q{"pk >=": 100}.Or(gomodel.Q{"email": "user@test.com"}),
)
```

## Multiple databases

You can pass multiple databases to the [Start](https://godoc.org/github.com/moiseshiraldo/gomodel/#Start)
function:

```go
gomodel.Start(map[string]gomodel.Database{
    "default": {
        Driver:   "postgres",
        Name:     "master",
        User:     "local",
        Password: "local",
    },
    "slave": {
        Driver:   "postgres",
        Name:     "slave",
        User:     "local",
        Password: "local",
    }
})
```

For single instances, you can select the target database with the
[SaveOn](https://godoc.org/github.com/moiseshiraldo/gomodel/#Instance.SaveOn) and
[DeleteOn](https://godoc.org/github.com/moiseshiraldo/gomodel/#Instance.DeleteOn)
methods:

```go
err := user.SaveOn("slave")
```

For querysets, you can use the [WithDB](https://godoc.org/github.com/moiseshiraldo/gomodel/#QuerySet.WithDB)
method:

```go
users, err := User.Objects.All().WithDB("slave").Load()
```

## Transactions

You can start a transaction using the `Database` [BeingTx](https://godoc.org/github.com/moiseshiraldo/gomodel/#Database.BeginTx)
method:

```go
db := gomodel.Databases()["default"]
tx, err := db.BeginTx()
```

Which returns a [Transaction](https://godoc.org/github.com/moiseshiraldo/gomodel/#Transaction)
that can be used as a target for instances and querysets:

```go
err := user.SaveOn(tx)
users, err := User.Objects.All().WithTx(tx).Load()
```

And commited or rolled back using the [Commit](https://godoc.org/github.com/moiseshiraldo/gomodel/#Transaction.Commit)
and [Rollback](https://godoc.org/github.com/moiseshiraldo/gomodel/#Transaction.Rollback)
methods.

# Testing

The `mocker` driver can be used to open a mocked database for unit testing:

```go
gomodel.Start(map[string]gomodel.Database{
    "default": {
        Driver:   "mocker",
        Name:     "test",
    },
})
```

The underlying [MockedEngine](https://godoc.org/github.com/moiseshiraldo/gomodel/#MockedEngine)
provides some useful tools for test assertions:

```go
func TestCreate(t *testing.T) {
    db := gomodel.Databases()["default"]
    mockedEngine := db.Engine.(gomodel.MockedEngine)
    _, err := User.Objects.Create(gomodel.Values{"email": "user@test.com"})
    if err != nil {
        t.Fatal(err)
    }
    // Calls returns the number of calls to the given method name.
    if mockedEngine.Calls("InsertRow") != 1 {
        t.Error("expected engine InsertRow method to be called")
    }
    // The Args field contains the arguments for the last call to each method.
    insertValues := mockedEngine.Args.InsertRow.Values
    if _, ok := mockedEngine.Args.InsertRow.Values["email"]; !ok {
        t.Error("email field missing on insert arguments")
    }
}
```

You can also change the return values of the engine methods:

```go
func TestCreateError(t *testing.T) {
    // Reset clears all the method calls, arguments and results.
    mockedEngine.Reset()
    // The Results fields can be used to set custom return values for each method.
    mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
    _, err := User.Objects.Create(Values{"email": "user@test.com"})
    if _, ok := err.(*gomodel.DatabaseError); !ok {
        t.Errorf("expected gomodel.DatabaseError, got %T", err)
    }
})
```
