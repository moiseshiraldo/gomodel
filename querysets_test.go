package gomodels

import (
	"fmt"
	"testing"
)

// rowsMocker mocks the results returned by the database engine
type rowsMocker struct {
	number int
}

func (r *rowsMocker) Close() error {
	return nil
}

func (r rowsMocker) Err() error {
	return nil
}

func (r *rowsMocker) Next() bool {
	r.number -= 1
	return r.number > -1
}

func (r *rowsMocker) Scan(dest ...interface{}) error {
	id := dest[0].(*int32)
	email := dest[1].(*string)
	*id = int32(r.number + 1)
	*email = "user@test.com"
	return nil
}

// TestGenericQuerySet tests the GenericQuerySet struct methods
func TestGenericQuerySet(t *testing.T) {
	// Model setup
	model := &Model{
		name: "User",
		pk:   "id",
		fields: Fields{
			"id":      IntegerField{Auto: true},
			"email":   CharField{MaxLength: 100},
			"active":  BooleanField{DefaultFalse: true},
			"updated": DateTimeField{AutoNow: true},
		},
		meta: Options{Container: Values{}},
	}
	// DB setup
	engine, _ := enginesRegistry["mocker"].Start(Database{})
	mockedEngine := engine.(MockedEngine)
	dbRegistry["default"] = Database{id: "default", Engine: engine}
	defer func() { dbRegistry = map[string]Database{} }()

	t.Run("New", func(t *testing.T) {
		qs := GenericQuerySet{}.New(model, GenericQuerySet{})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		if gqs.model != model {
			t.Error("expected queryset to be linked to model")
		}
		if gqs.database != "default" {
			t.Error("expected queryset to be linked to default database")
		}
		if len(gqs.fields) != len(model.fields) {
			t.Errorf(
				"expected qs fields length to be %d, got %d",
				len(model.fields), len(gqs.fields),
			)
		}
	})

	t.Run("Wrap", func(t *testing.T) {
		qs := GenericQuerySet{model: model}
		wqs, ok := GenericQuerySet{}.Wrap(qs).(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs.Wrap(qs))
		}
		if wqs.model != model {
			t.Error("expected queryset to be linked to model")
		}
	})

	t.Run("Model", func(t *testing.T) {
		qs := GenericQuerySet{model: model}
		if qs.Model() != model {
			t.Error("expected model to be returned")
		}
	})

	t.Run("WithContainer", func(t *testing.T) {
		qs := GenericQuerySet{base: GenericQuerySet{}}.WithContainer(Values{})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		if _, ok := gqs.container.(Values); !ok {
			t.Errorf("expected container to be Values, got %T", gqs.container)
		}
	})

	t.Run("Filter", func(t *testing.T) {
		qs := GenericQuerySet{base: GenericQuerySet{}}.Filter(Q{"active": true})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		if _, ok := gqs.cond.Conditions()["active"]; !ok {
			t.Error("filter is missing active condition")
		}
	})

	t.Run("Exclude", func(t *testing.T) {
		qs := GenericQuerySet{base: GenericQuerySet{}}.Exclude(Q{"id": 10})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		filter, _, isNot := gqs.cond.Next()
		if _, ok := filter.Conditions()["id"]; !ok {
			t.Error("filter is missing id condition")
		}
		if !isNot {
			t.Error("expected filter to be NOT")
		}
	})

	t.Run("Only", func(t *testing.T) {
		qs := GenericQuerySet{base: GenericQuerySet{}}.Only("id", "email")
		gqs := qs.(GenericQuerySet)
		if len(gqs.fields) != 2 {
			t.Fatalf("expected qs fields len to be 2, got %d", len(gqs.fields))
		}
		if gqs.fields[0] != "id" || gqs.fields[1] != "email" {
			t.Errorf(
				"expected fields (id, email), got: (%s, %s)",
				gqs.fields[0], gqs.fields[1],
			)
		}
	})

	t.Run("QueryInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Query()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("Query", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.SelectQuery.Query = Query{Stmt: "test query"}
		qs := GenericQuerySet{model: model, database: "default"}
		query, err := qs.Query()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("SelectQuery") != 1 {
			t.Fatal("expected engine SelectQuery method to be called")
		}
		if query.Stmt != "test query" {
			t.Errorf("expected 'test query', got %s", query.Stmt)
		}
	})

	t.Run("LoadInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Load()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("LoadInvalidRecipients", func(t *testing.T) {
		mockedEngine.Reset()
		container := struct {
			id    int
			email string
		}{}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: container,
			fields:    []string{"id", "email", "active"},
		}
		_, err := qs.Load()
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("LoadDatabaseError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Err = fmt.Errorf("db error")
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email", "active"},
		}
		_, err := qs.Load()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("GetRows") != 1 {
			t.Error("expected engine GetRows method to be called")
		}
	})

	t.Run("Load", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{2}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		result, err := qs.Load()
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result))
		}
		values, ok := result[0].container.(Values)
		if !ok {
			t.Fatalf("expected Values, got %T", result[0].container)
		}
		if id, ok := values["id"].(int32); !ok || id != int32(2) {
			t.Errorf("expected id to be 2, got %d", id)
		}
		if e, ok := values["email"].(string); !ok || e != "user@test.com" {
			t.Errorf("expected email to be user@test.com, got %s", e)
		}
	})

	t.Run("SliceInvalid", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		_, err := qs.Slice(10, 4)
		if _, ok := err.(*QuerySetError); !ok {
			t.Errorf("expected QuerySetError, got %T", err)
		}
	})

	t.Run("Slice", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{5}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		_, err := qs.Slice(2, 4)
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("GetRows") != 1 {
			t.Fatal("expected engine GetRows method to be called")
		}
		args := mockedEngine.Args.GetRows
		if args.Start != 2 || args.End != 4 {
			t.Errorf(
				"expected GetRows args (2, 4), got (%d, %d)",
				args.Start, args.End,
			)
		}
	})

	t.Run("GetInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("GetInvalidContainer", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("GetInvalidRecipients", func(t *testing.T) {
		mockedEngine.Reset()
		container := struct {
			id    int
			email string
		}{}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: container,
			fields:    []string{"id", "email", "active"},
		}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("GetDatabaseError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Err = fmt.Errorf("db error")
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email", "active"},
		}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("GetRows") != 1 {
			t.Error("expected engine GetRows method to be called")
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{0}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*ObjectNotFoundError); !ok {
			t.Errorf("expected ObjectNotFoundError, got %T", err)
		}
	})

	t.Run("GetMultiple", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{2}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		_, err := qs.Get(Q{"id": 23})
		if _, ok := err.(*MultipleObjectsError); !ok {
			t.Errorf("expected MultipleObjectsError, got %T", err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{1}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: Values{},
			fields:    []string{"id", "email"},
		}
		instance, err := qs.Get(Q{"id": 1})
		if err != nil {
			t.Fatal(err)
		}
		values, ok := instance.container.(Values)
		if !ok {
			t.Fatalf("expected Values, got %T", instance.container)
		}
		if id, ok := values["id"].(int32); !ok || id != int32(1) {
			t.Errorf("expected id to be 1, got %d", id)
		}
		if e, ok := values["email"].(string); !ok || e != "user@test.com" {
			t.Errorf("expected email to be user@test.com, got %s", e)
		}
	})

	t.Run("GetStruct", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.GetRows.Rows = &rowsMocker{1}
		type userContainer struct {
			Id    int32
			Email string
		}
		qs := GenericQuerySet{
			model:     model,
			database:  "default",
			container: userContainer{},
			fields:    []string{"id", "email"},
		}
		instance, err := qs.Get(Q{"id": 1})
		if err != nil {
			t.Fatal(err)
		}
		user, ok := instance.container.(*userContainer)
		if !ok {
			t.Fatalf("expected userContainer, got %T", instance.container)
		}
		if user.Id != 1 {
			t.Errorf("expected id to be 1, got %d", user.Id)
		}
		if user.Email != "user@test.com" {
			t.Errorf("expected email to be user@test.com, got %s", user.Email)
		}
	})

	t.Run("ExistsInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Exists()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("ExistsDatabaseError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.Exists.Err = fmt.Errorf("db error")
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Exists()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("Exists") != 1 {
			t.Error("expected engine Exists method to be called")
		}
	})

	t.Run("Exists", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.Exists.Result = true
		qs := GenericQuerySet{model: model, database: "default"}
		result, err := qs.Exists()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("Exists") != 1 {
			t.Fatal("expected engine Exists method to be called")
		}
		if !result {
			t.Error("expected true, got false")
		}
	})

	t.Run("CountInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Count()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("CountDatabaseError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CountRows.Err = fmt.Errorf("db error")
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Count()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("CountRows") != 1 {
			t.Error("expected engine CountRows method to be called")
		}
	})

	t.Run("Count", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.CountRows.Number = 42
		qs := GenericQuerySet{model: model, database: "default"}
		result, err := qs.Count()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("CountRows") != 1 {
			t.Fatal("expected engine CountRows method to be called")
		}
		if result != 42 {
			t.Errorf("expected 42, got %d", result)
		}
	})

	t.Run("UpdateInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Update(Values{})
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("UpdateInvalidValues", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Update("invalid")
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("UpdateInvalidValues", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Update(Values{"active": true})
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("UpdateRows") != 1 {
			t.Fatal("expected engine UpdateRows method to be called")
		}
		dbValues := mockedEngine.Args.UpdateRows.Values
		if _, ok := dbValues["updated"]; !ok {
			t.Error("expected UpdateRows values to contain updated field")
		}
	})

	t.Run("DeleteInvalidDB", func(t *testing.T) {
		mockedEngine.Reset()
		qs := GenericQuerySet{model: model, database: "slave"}
		_, err := qs.Delete()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("DeleteDatabaseError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DeleteRows.Err = fmt.Errorf("db error")
		qs := GenericQuerySet{model: model, database: "default"}
		_, err := qs.Delete()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("DeleteRows") != 1 {
			t.Error("expected engine DeleteRows method to be called")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.DeleteRows.Number = 4
		qs := GenericQuerySet{model: model, database: "default"}
		result, err := qs.Delete()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("DeleteRows") != 1 {
			t.Fatal("expected engine CountRows method to be called")
		}
		if result != 4 {
			t.Errorf("expected 4, got %d", result)
		}
	})
}
