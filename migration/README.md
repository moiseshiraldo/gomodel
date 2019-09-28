# Migrations

The migration package provides the tools to detect and manage the changes made
to application models, store them in version control and apply them to the
database schema.

1. [**Quick start**](#quick-start)
2. [**Managing changes**](#managing-changes)
   - [Migration files](#migration-files)
   - [Automatic detection](#automatic-detection)
   - [Applying migrations](#apply-migrations)  
4. [**Supported operations**](#supported-operations)
   - [CreatedModel](#createmodel)
   - [DeleteModel](#deletemodel)
   - [AddFields](#addfields)
   - [RemoveFields](#removefields)
   - [AddIndex](#addindex)
   - [RemoveIndex](#removeindex)

# Quick start

Example usage:

```go
package main

import (
    "fmt"
    _ "github.com/gwenn/gosqlite"  // Imports SQLite driver.
    "github.com/moiseshiraldo/gomodel"
    "github.com/moiseshiraldo/gomodel/migration"
    "os"
)

var User = gomodel.New(
    "User",
    gomodel.Fields{
        "email":   gomodel.CharField{MaxLength: 100, Index: true},
        "active":  gomodel.BooleanField{DefaultFalse: true},
        "created": gomodel.DateTimeField{AutoNowAdd: true},
    },
    gomodel.Options{},
)

var app = gomodel.NewApp("main", "", User.Model)

func setup() {
    gomodel.Register(app)
    gomodel.Start(map[string]gomodel.Database{
        "default": {
            Driver: "sqlite3",
            Name:   ":memory:",
        },
    })
}

func main() {
    setup()
    // Detects changes and writes migration files.
    if _, err := migration.Make("main", migration.MakeOptions{}); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    // Runs all pending migrations.
    if err := migration.Run(migration.RunOptions{}); err != nil {
        fmt.Println("something went wrong:", err)
        os.Exit(1)
    }
    fmt.Println("Changes applied to database schema!")
}
```

Check the [MakeOptions](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#MakeOptions)
and [RunOptions](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#RunOptions)
documentation for more details.

If the database schema is not going to change, or you don't need to keep track
of the changes (e.g. in-memory SQLite database), you can detect and apply all
the changes directly using the [MakeAndRun](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#MakeAndRun)
function:

```go
func main() {
    setup()
    if err := migration.MakeAndRun("default"); err != nil {
        fmt.Println("something went wrong:", err)
    }
}
```

# Managing changes

When your project contains a considerable number of models that change over
time, it's probably a good idea to keep track of the database schema changes
along with the code ones.

## Migration files

GoModel will read and write the schema changes to the directory specified when
you create and register an application:

```go
var app = gomodel.NewApp("main", "/home/dev/project/main/migrations", User.Model)
``` 

If you specify a relative path, the full one will be constructed from
`$GOPATH/src`. Each application must have a unique migrations path, that only
contains JSON files following the naming convention `{number}_{name}.json`,
where `number` is a four digit number representing the order of the node in the
graph of changes and `name` can be any arbitrary name (e.g. `0001_initial.json`).

You can create a new empty migration using the [Make](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#Make)
function:

```go
_, err := migration.Make("main", migration.MakeOptions{Empty: true})
if err != nil {
    fmt.Println("something went wrong:", err)
}
```

The migration file will look something like this:

```json
{
  "App": "main",
  "Dependencies": [
    [
      "main",
      "0001_initial"
    ]
  ],
  "Operations": []
}
``` 

The `App` attribute is the name of the application. `Dependencies` is the list
of migrations (application and full name) that should be applied before the
described one. And `Operations` is the list of changes to be applied to the
database schema (see [supported operations](#supported-operations)).

## Automatic detection

The [Make](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#Make)
function can be used to automatically detect any changes on the application
models and write them to migration files:

```go
if _, err := migration.Make("main", migration.MakeOptions{}); err != nil {
    fmt.Println(err)
}
```

The function will load and process the existing migration files, compare the
resulted state with the model definitions and write any changes to new
migration files.

## Applying migrations

The [Run](https://godoc.org/github.com/moiseshiraldo/gomodel/migration#Run)
function can be used to apply changes from migration files to the database
schema. For example:

```go
options := migration.RunOptions{
    App: "main",
    Node: "0002",
    Database: "default",
}
if _, err := migration.Run(options); err != nil {
    fmt.Println(err)
}
```

The code above would apply the changes up to and including the second migration.
If any migration greater than the second one was already applied, it would be
reverted.

# Supported operations 

## CreateModel

```json
{
  "CreateModel": {
    "Name": "User",
    "Table": "users",
    "Fields": {
      "id": {
        "IntegerField": {
          "PrimaryKey": true,
          "Auto": true
        }
      },
      "email": {
        "CharField": {
          "MaxLength": 100,
          "Index": true
        }
      }
    }
  }
}
```

## DeleteModel

```json
{
  "DeleteModel": {
    "Name": "User",
  }
}
```

## AddFields

```json
{
  "AddFields": {
    "Model": "User",
    "Fields": {
      "active": {
        "BooleanField": {
          "DefaultFalse": true
        }
      },
      "created": {
        "DateTimeField": {
          "AutoNowAdd": true
        }
      }
    }
  }
}
```

## RemoveFields

```json
{
  "RemoveFields": {
    "Model": "User",
    "Fields": [
      "active",
      "created"
    ]
  }
}
```

## AddIndex

```json
{
  "AddIndex": {
    "Model": "User",
    "Name": "users_user_email_auto_idx",
    "Fields": [
      "email"
    ]
  }
}
```

## RemoveIndex

```json
{
  "RemoveIndex": {
    "Model": "User",
    "Name": "users_user_email_auto_idx"
  }
}
```
