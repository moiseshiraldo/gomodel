package gomodel

import (
	"fmt"
	"testing"
)

// TestStart tests the Start function
func TestStart(t *testing.T) {
	defer func() { dbRegistry = map[string]Database{} }()

	t.Run("UnknownDriver", func(t *testing.T) {
		dbRegistry = map[string]Database{}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Start function to panic")
			}
		}()
		Start(map[string]Database{
			"default": {Driver: "qwerty"},
		})
	})

	t.Run("MissingDefault", func(t *testing.T) {
		dbRegistry = map[string]Database{}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected Start function to panic")
			}
		}()
		Start(map[string]Database{
			"slave": {Driver: "mocker"},
		})
	})

	t.Run("Success", func(t *testing.T) {
		dbRegistry = map[string]Database{}
		err := Start(map[string]Database{
			"default": {Driver: "mocker"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := dbRegistry["default"]; !ok {
			t.Fatal("db registry is missing default db")
		}
		mockedEngine := dbRegistry["default"].Engine.(MockedEngine)
		if mockedEngine.Calls("Start") != 1 {
			t.Error("expected engine Start method to be called")
		}
	})
}

// TestStop tests the Stop function
func TestStop(t *testing.T) {
	defer func() { dbRegistry = map[string]Database{} }()

	t.Run("Error", func(t *testing.T) {
		engine, _ := enginesRegistry["mocker"].Start(Database{})
		mockedEngine := engine.(MockedEngine)
		mockedEngine.Results.Stop = fmt.Errorf("db error")
		dbRegistry = map[string]Database{
			"default": {id: "default", Engine: mockedEngine},
		}
		err := Stop()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		engine, _ := enginesRegistry["mocker"].Start(Database{})
		mockedEngine := engine.(MockedEngine)
		dbRegistry = map[string]Database{
			"default": {id: "default", Engine: mockedEngine},
		}
		err := Stop()
		if err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("Stop") != 1 {
			t.Error("expected engine Stop method to be called")
		}
	})

}

// TestDatabases tests the Databases function
func TestDatabases(t *testing.T) {
	dbRegistry = map[string]Database{"default": {}}
	defer func() { dbRegistry = map[string]Database{} }()

	result := Databases()
	if _, ok := result["default"]; !ok {
		t.Fatal("result is missing the default db")
	}
	result["slave"] = Database{}
	if _, ok := dbRegistry["slave"]; ok {
		t.Error("original db registry was modified")
	}

}

// TestDatabase tests the Database structs methods
func TestDatabase(t *testing.T) {
	engine, _ := enginesRegistry["mocker"].Start(Database{})
	mockedEngine := engine.(MockedEngine)
	db := Database{id: "slave", Engine: mockedEngine}

	t.Run("Id", func(t *testing.T) {
		if db.Id() != "slave" {
			t.Errorf("expected slave, got %s", db.Id())
		}
	})

	t.Run("Connection", func(t *testing.T) {
		db.Conn()
		if mockedEngine.Calls("DB") != 1 {
			t.Error("expected engine DB method to be called")
		}
	})

	t.Run("TxError", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		_, err := db.BeginTx()
		if _, ok := err.(*DatabaseError); !ok {
			t.Errorf("expected DatabaseError, got %T", err)
		}
	})

	t.Run("BeginTx", func(t *testing.T) {
		mockedEngine.Reset()
		if _, err := db.BeginTx(); err != nil {
			t.Fatal(err)
		}
		if mockedEngine.Calls("BeginTx") != 1 {
			t.Error("expected engine BeginTx method to be called")
		}
	})
}

// TestTransaction tests the Transaction structs methods
func TestTransaction(t *testing.T) {
	engine, _ := enginesRegistry["mocker"].Start(Database{})
	mockedEngine := engine.(MockedEngine)
	tx := Transaction{DB: Database{}, Engine: engine}

	t.Run("Connection", func(t *testing.T) {
		tx.Conn()
		if mockedEngine.Calls("Tx") != 1 {
			t.Error("expected engine Tx method to be called")
		}
	})

	t.Run("Commit", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		tx.Commit()
		if mockedEngine.Calls("CommitTx") != 1 {
			t.Error("expected engine CommitTx method to be called")
		}
	})

	t.Run("BeginTx", func(t *testing.T) {
		mockedEngine.Reset()
		mockedEngine.Results.BeginTx = fmt.Errorf("db error")
		tx.Rollback()
		if mockedEngine.Calls("RollbackTx") != 1 {
			t.Error("expected engine RollbackTx method to be called")
		}
	})
}
