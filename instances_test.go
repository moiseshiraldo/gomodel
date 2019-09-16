package gomodel

import (
	"fmt"
	"testing"
	"time"
)

// TestInstance tests the Instance struct methods
func TestInstance(t *testing.T) {
	model := &Model{
		name: "User",
		fields: Fields{
			"id":     IntegerField{Auto: true},
			"email":  CharField{MaxLength: 100},
			"active": BooleanField{},
			"dob":    DateField{Null: true},
		},
	}
	dob := NullTime{
		Time:  time.Date(1942, 11, 27, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
	instance := Instance{model, Values{"email": "user@test.com", "dob": dob}}

	t.Run("Model", func(t *testing.T) {
		if instance.Model() != model {
			t.Error("wrong model returned")
		}
	})

	t.Run("Container", func(t *testing.T) {
		container := instance.Container()
		values, ok := container.(Values)
		if !ok {
			t.Fatalf("expected Values, got %T", container)
		}
		if _, ok := values["email"]; !ok {
			t.Error("container is missing email field")
		}
	})

	t.Run("GetIfUnknownField", func(t *testing.T) {
		if _, ok := instance.GetIf("foo"); ok {
			t.Error("expected no value to be returned")
		}
	})

	t.Run("GetIfNoValue", func(t *testing.T) {
		if _, ok := instance.GetIf("active"); ok {
			t.Error("expected no value to be returned")
		}
	})

	t.Run("GetIf", func(t *testing.T) {
		val, ok := instance.GetIf("email")
		if !ok {
			t.Fatal("expected value to be returned")
		}
		if s, ok := val.(string); !ok || s != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", s)
		}
	})

	t.Run("GetIfNullable", func(t *testing.T) {
		val, ok := instance.GetIf("dob")
		if !ok {
			t.Fatal("expected value to be returned")
		}
		if d, ok := val.(time.Time); !ok || !d.Equal(dob.Time) {
			t.Errorf("expected %s, got %s", dob.Time, d)
		}
	})

	t.Run("Get", func(t *testing.T) {
		if val := instance.Get("foo"); val != nil {
			t.Errorf("expected nil, got %s", val)
		}
	})

	t.Run("SetUnknownField", func(t *testing.T) {
		err := instance.Set("foo", 42)
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("SetInvalidValue", func(t *testing.T) {
		err := instance.Set("active", 42)
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("SetUnknownStructField", func(t *testing.T) {
		instance.container = struct{ active bool }{false}
		err := instance.Set("email", "new@test.com")
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("SetStructFieldInvalidValue", func(t *testing.T) {
		instance.container = struct{ active bool }{false}
		err := instance.Set("active", 42)
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("Set", func(t *testing.T) {
		instance.container = Values{"email": "user@test.com", "dob": dob}
		if err := instance.Set("email", "new@test.com"); err != nil {
			t.Fatal(err)
		}
		values := instance.container.(Values)
		if s, ok := values["email"].(string); !ok || s != "new@test.com" {
			t.Errorf("expected new@test.com, got %s", s)
		}
	})

	t.Run("SetInvalidValues", func(t *testing.T) {
		err := instance.SetValues(Values{"active": 42})
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("SetValues", func(t *testing.T) {
		instance.container = Values{"email": "user@test.com", "active": false}
		newValues := Values{"email": "new@test.com", "active": true}
		if err := instance.SetValues(newValues); err != nil {
			t.Fatal(err)
		}
		values := instance.container.(Values)
		if s, ok := values["email"].(string); !ok || s != "new@test.com" {
			t.Errorf("expected new@test.com, got %s", s)
		}
		if a, ok := values["active"].(bool); !ok || !a {
			t.Errorf("expected true, got %t", a)
		}
	})

	t.Run("SetValues", func(t *testing.T) {
		instance.container = Values{"email": "user@test.com", "active": false}
		newValues := struct {
			Email  string
			Active bool
			Dob    string
		}{"new@test.com", true, "1972-04-12"}
		if err := instance.SetValues(newValues); err != nil {
			t.Fatal(err)
		}
		values := instance.container.(Values)
		if s, ok := values["email"].(string); !ok || s != "new@test.com" {
			t.Errorf("expected new@test.com, got %s", s)
		}
		if a, ok := values["active"].(bool); !ok || !a {
			t.Errorf("expected true, got %t", a)
		}
	})
}

// TestInstanceSave tests the Instance save method
func TestInstanceSave(t *testing.T) {
	// Model setup
	model := &Model{
		name: "User",
		pk:   "id",
		fields: Fields{
			"id":      IntegerField{Auto: true, PrimaryKey: true},
			"email":   CharField{MaxLength: 100},
			"active":  BooleanField{Default: true},
			"created": DateTimeField{AutoNowAdd: true},
			"updated": DateTimeField{AutoNow: true},
		},
	}
	instance := Instance{model, Values{}}
	// DB Setup
	engine, _ := enginesRegistry["mocker"].Start(Database{})
	mockedEngine := engine.(MockedEngine)
	dbRegistry["default"] = Database{id: "default", Engine: engine}
	defer func() { dbRegistry = map[string]Database{} }()

	t.Run("UnknownField", func(t *testing.T) {
		err := instance.Save("foo", "bar")
		if _, ok := err.(*ContainerError); !ok {
			t.Errorf("expected ContainerError, got %T", err)
		}
	})

	t.Run("InsertError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Err = fmt.Errorf("db error")
		err := instance.Save()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Error("expected engine InsertRow method to be called")
		}
	})

	t.Run("Insert", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Id = 23
		instance.container = Values{"email": "user@test.com"}
		err := instance.Save()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Error("expected engine InsertRow method to be called")
		}
		insertValues := mockedEngine.Args.InsertRow.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"active", "created", "updated", "email"} {
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

	t.Run("InsertOnInvalidTarget", func(t *testing.T) {
		mockedEngine.Reset()
		err := instance.SaveOn(mockedEngine)
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("InsertOn", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.InsertRow.Id = 23
		tx := &Transaction{Engine: mockedEngine, DB: Database{id: "default"}}
		instance.container = Values{"email": "user@test.com"}
		err := instance.SaveOn(tx)
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Error("expected engine InsertRow method to be called")
		}
		insertValues := mockedEngine.Args.InsertRow.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"active", "created", "updated", "email"} {
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

	t.Run("UpdateError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.UpdateRows.Err = fmt.Errorf("db error")
		instance.container = Values{"id": 23, "email": "user@test.com"}
		err := instance.Save()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
		if mockedEngine.Calls("UpdateRows") != 1 {
			t.Error("expected engine UpdateRows method to be called")
		}
	})

	t.Run("Update", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.UpdateRows.Number = 1
		instance.container = Values{"id": 23, "email": "user@test.com"}
		err := instance.Save()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("UpdateRows") != 1 {
			t.Fatal("expected engine UpdateRows method to be called")
		}
		updateValues := mockedEngine.Args.UpdateRows.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"updated", "email"} {
			if _, ok := updateValues[name]; !ok {
				t.Errorf("missing %s value on UpdateRows arguments", name)
			}
			if _, ok := instanceValues[name]; !ok {
				t.Errorf("instance is missing %s value", name)
			}
		}
	})

	t.Run("UpdateOnInvalidTarget", func(t *testing.T) {
		mockedEngine.Reset()
		instance.container = Values{"id": 23, "email": "user@test.com"}
		err := instance.SaveOn(mockedEngine)
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("UpdateOn", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.UpdateRows.Number = 1
		instance.container = Values{"id": 23, "email": "user@test.com"}
		err := instance.SaveOn("default")
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("UpdateRows") != 1 {
			t.Fatal("expected engine UpdateRows method to be called")
		}
		updateValues := mockedEngine.Args.UpdateRows.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"updated", "email"} {
			if _, ok := updateValues[name]; !ok {
				t.Errorf("missing %s value on UpdateRows arguments", name)
			}
			if _, ok := instanceValues[name]; !ok {
				t.Errorf("instance is missing %s value", name)
			}
		}
	})

	t.Run("UpdateNotExists", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.UpdateRows.Number = 0
		mockedEngine.Results.InsertRow.Id = 23
		instance.container = Values{"id": 23, "email": "user@test.com"}
		err := instance.Save()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("InsertRow") != 1 {
			t.Fatal("expected engine Insert method to be called")
		}
		insertValues := mockedEngine.Args.InsertRow.Values
		instanceValues := instance.container.(Values)
		for _, name := range []string{"active", "created", "updated", "email"} {
			if _, ok := insertValues[name]; !ok {
				t.Errorf("missing %s value on InsertRow arguments", name)
			}
			if _, ok := instanceValues[name]; !ok {
				t.Errorf("instance is missing %s value", name)
			}
		}
	})
}
