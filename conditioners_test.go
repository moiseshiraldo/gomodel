package gomodel

import (
	"testing"
)

// TestQ test the Q struct methods
func TestQ(t *testing.T) {
	filter := Q{"email": "user@test.com"}

	t.Run("Predicate", func(t *testing.T) {
		cond := filter.Conditions()
		if val, ok := cond["email"].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", val)
		}
	})

	t.Run("Next", func(t *testing.T) {
		cond, _, _ := filter.Next()
		if cond != nil {
			t.Errorf("expected nil, got %v", cond)
		}
	})

	t.Run("And", func(t *testing.T) {
		filter := filter.And(Q{"active": true})
		cond := filter.Conditions()
		if val, ok := cond["email"].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", val)
		}
		next, isOr, isNot := filter.Next()
		if next == nil {
			t.Fatal("expected Next() to return a Conditioner")
		}
		if isOr || isNot {
			t.Error("expected next conditioner to be AND")
		}
		cond = next.Conditions()
		if val, ok := cond["active"].(bool); !ok || !val {
			t.Errorf("expected true, got %t", val)
		}
	})

	t.Run("AndNot", func(t *testing.T) {
		filter := filter.AndNot(Q{"active": true})
		cond := filter.Conditions()
		if val, ok := cond["email"].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", val)
		}
		next, isOr, isNot := filter.Next()
		if next == nil {
			t.Fatal("expected Next() to return a Conditioner")
		}
		if isOr || !isNot {
			t.Error("expected next conditioner to be AND NOT")
		}
		cond = next.Conditions()
		if val, ok := cond["active"].(bool); !ok || !val {
			t.Errorf("expected true, got %t", val)
		}
	})

	t.Run("Or", func(t *testing.T) {
		filter := filter.Or(Q{"active": true})
		cond := filter.Conditions()
		if val, ok := cond["email"].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", val)
		}
		next, isOr, isNot := filter.Next()
		if next == nil {
			t.Fatal("expected Next() to return a Conditioner")
		}
		if !isOr || isNot {
			t.Error("expected next conditioner to be OR")
		}
		cond = next.Conditions()
		if val, ok := cond["active"].(bool); !ok || !val {
			t.Errorf("expected true, got %t", val)
		}
	})

	t.Run("OrNot", func(t *testing.T) {
		filter := filter.OrNot(Q{"active": true})
		cond := filter.Conditions()
		if val, ok := cond["email"].(string); !ok || val != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", val)
		}
		next, isOr, isNot := filter.Next()
		if next == nil {
			t.Fatal("expected Next() to return a Conditioner")
		}
		if !isOr || !isNot {
			t.Error("expected next conditioner to be OR NOT")
		}
		cond = next.Conditions()
		if val, ok := cond["active"].(bool); !ok || !val {
			t.Errorf("expected true, got %t", val)
		}
	})
}
