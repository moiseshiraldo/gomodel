package gomodel

import (
	"fmt"
	"testing"
)

type mockedQuerySet struct {
	GenericQuerySet
	calls map[string]int
}

func (qs mockedQuerySet) Wrap(gen GenericQuerySet) QuerySet {
	qs.GenericQuerySet = gen
	return qs
}

func (qs mockedQuerySet) Filter(cond Conditioner) QuerySet {
	qs.calls["Filter"] += 1
	return qs
}

func (qs mockedQuerySet) Exclude(cond Conditioner) QuerySet {
	qs.calls["Exclude"] += 1
	return qs
}

func (qs mockedQuerySet) Get(cond Conditioner) (*Instance, error) {
	qs.calls["Get"] += 1
	return &Instance{container: Values{"qs": qs}}, nil
}

func (qs mockedQuerySet) WithContainer(container Container) QuerySet {
	qs.calls["WithContainer"] += 1
	return qs
}

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
	manager := Manager{
		Model: model, QuerySet: mockedQuerySet{calls: map[string]int{}},
	}
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
		mocked, ok := qs.(mockedQuerySet)
		if !ok {
			t.Fatalf("expected mockedQuerySet, got %T", qs)
		}
		if mocked.GenericQuerySet.model != model {
			t.Error("expected queryset to be linked to model")
		}
	})

	t.Run("Filter", func(t *testing.T) {
		qs := manager.Filter(Q{"active": true})
		mocked, ok := qs.(mockedQuerySet)
		if !ok {
			t.Fatalf("expected mockedQuerySet, got %T", qs)
		}
		if mocked.calls["Filter"] != 1 {
			t.Error("expected  queryset Filter method to be called")
		}
	})

	t.Run("Exclude", func(t *testing.T) {
		qs := manager.Exclude(Q{"email": "user@test.com"})
		mocked, ok := qs.(mockedQuerySet)
		if !ok {
			t.Fatalf("expected mockedQuerySet, got %T", qs)
		}
		if mocked.calls["Exclude"] != 1 {
			t.Error("expected queryset Exclude method to be called")
		}
	})

	t.Run("Get", func(t *testing.T) {
		instance, _ := manager.Get(Q{"pk": 42})
		mocked, ok := instance.container.(Values)["qs"].(mockedQuerySet)
		if !ok {
			t.Fatal("expected mockedQuerySet")
		}
		if mocked.calls["Get"] != 1 {
			t.Error("expected queryset Exclude method to be called")
		}
	})

	t.Run("WithContainer", func(t *testing.T) {
		qs := manager.WithContainer(Values{})
		mocked, ok := qs.(mockedQuerySet)
		if !ok {
			t.Fatalf("expected mockedQuerySet, got %T", qs)
		}
		if mocked.calls["WithContainer"] != 1 {
			t.Error("expected queryset WithContainer method to be called")
		}
	})

}
