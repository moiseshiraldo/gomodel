package gomodel

import (
	"testing"
)

// TestMockedEngine tests the MockedEngine methods
func TestMockedEngine(t *testing.T) {
	engine := MockedEngine{
		calls:   make(map[string]int),
		Args:    &MockedEngineArgs{},
		Results: &MockedEngineResults{},
		tx:      true,
	}
	model := &Model{}

	t.Run("Calls", func(t *testing.T) {
		matrix := map[string]error{
			"CreateTable": engine.CreateTable(model, true),
			"CommitTx":    engine.CommitTx(),
			"RollbackTx":  engine.RollbackTx(),
			"RenameTable": engine.RenameTable(model, model),
			"DropTable":   engine.DropTable(model),
			"AddIndex":    engine.AddIndex(model, "test_idx", "email"),
			"DropIndex":   engine.DropIndex(model, "test_idx"),
			"AddColumns":  engine.AddColumns(model, Fields{}),
			"DropColumns": engine.DropColumns(model),
		}
		for method, err := range matrix {
			if err != nil {
				t.Fatal(err)
			}
			if engine.Calls(method) != 1 {
				t.Fatalf("expected method %s to be called", method)
			}
		}
	})

	t.Run("TxSupport", func(t *testing.T) {
		engine.Results.TxSupport = true
		if !engine.TxSupport() {
			t.Fatal("expected true, got false")
		}
	})
}
