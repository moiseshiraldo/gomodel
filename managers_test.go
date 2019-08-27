package gomodels

import (
	"fmt"
	"testing"
)

// TestManager tests the Manager struct methods
func TestManager(t *testing.T) {
	// Model setup
	model := &Model{
		name: "User",
		pk:   "id",
		fields: Fields{
			"id":      IntegerField{Auto: true},
			"email":   CharField{MaxLength: 100},
			"active":  BooleanField{DefaultFalse: true},
			"created": DateTimeField{AutoNowAdd: true},
		},
		meta: Options{Container: Values{}},
	}
	manager := Manager{model, GenericQuerySet{}}
	// DB setup
	engine, _ := enginesRegistry["mocker"].Start(Database{})
	mockedEngine := engine.(MockedEngine)
	dbRegistry["default"] = Database{id: "default", Engine: engine}
	defer func() { dbRegistry = map[string]Database{} }()

	t.Run("CreateInvalidContainer", func(t *testing.T) {
		_, err := manager.Create("invalid")
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("CreateInvalidValue", func(t *testing.T) {
		_, err := manager.Create(Values{"active": 42})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("CreateInsertError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		_, err := manager.Create(Values{"email": "user@test.com"})
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Error("expected engine InsertRow method to be called")
		}
	})

	t.Run("Create", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Id = 23
		instance, err := manager.Create(Values{"email": "user@test.com"})
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Error("expected engine InsertRow method to be called")
		}
		insertValues := mockedEngine.Args.InsertRow.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"active", "created", "email"} {
			if _, ok := insertValues[name]; !ok {
				t.Errorf("missing %s value on InsertRow arguments", name)
			}
			if _, ok := instanceValues[name]; !ok {
				t.Errorf("instance is missing %s value", name)
			}
		}
		if id, ok := instanceValues["id"]; !ok || id != int32(23) {
			t.Errorf("expected id to be 23, got %s", id)
		}
	})

	t.Run("All", func(t *testing.T) {
		qs := manager.All()
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		if gqs.model != model {
			t.Error("expected queryset to be linked to model")
		}
	})

	t.Run("Filter", func(t *testing.T) {
		qs := manager.Filter(Q{"active": true})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		filter := gqs.cond.Predicate()
		if _, ok := filter["active"]; !ok {
			t.Error("filter is missing active condition")
		}
	})

	t.Run("Exclude", func(t *testing.T) {
		qs := manager.Exclude(Q{"email": "user@test.com"})
		gqs, ok := qs.(GenericQuerySet)
		if !ok {
			t.Fatalf("expected GenericQuerySet, got %T", qs)
		}
		cond, _, isNot := gqs.cond.Next()
		if !isNot {
			t.Error("expected filter to be NOT")
		}
		filter := cond.Predicate()
		if _, ok := filter["email"]; !ok {
			t.Error("filter is missing email condition")
		}
	})

}
