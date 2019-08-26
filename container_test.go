package gomodels

import (
	"testing"
	"time"
)

// TestValues tests the Values type methods
func TestValues(t *testing.T) {
	values := Values{
		"email":  "user@test.com",
		"active": true,
	}

	t.Run("New", func(t *testing.T) {
		builder := values.New()
		newValues, ok := builder.(Values)
		if !ok {
			t.Fatalf("expected Values, got %T", builder)
		}
		if len(newValues) > 0 {
			t.Error("expected new builder to be empty")
		}
	})

	t.Run("GetUnknown", func(t *testing.T) {
		if _, ok := values.Get("username"); ok {
			t.Error("expected no value to be returned")
		}
	})

	t.Run("Get", func(t *testing.T) {
		val, ok := values.Get("email")
		if !ok {
			t.Fatal("expected value to be returned")
		}
		if s, ok := val.(string); !ok || s != "user@test.com" {
			t.Errorf("expected user@test.com, got %s", s)
		}
	})

	t.Run("SetInvalid", func(t *testing.T) {
		if err := values.Set("active", 42, BooleanField{}); err == nil {
			t.Error("expected cannot convert value error")
		}
	})

	t.Run("SetInvalidNull", func(t *testing.T) {
		if err := values.Set("loginAttempts", nil, IntegerField{}); err == nil {
			t.Error("expected cannot convert value error")
		}
	})

	t.Run("SetNull", func(t *testing.T) {
		field := IntegerField{Null: true}
		if err := values.Set("loginAttempts", nil, field); err != nil {
			t.Fatal(err)
		}
		if _, ok := values["loginAttempts"].(NullInt32); !ok {
			t.Errorf(
				"expected gomodels.NullInt32, got %T", values["loginAttempts"],
			)
		}
	})

	t.Run("SetValue", func(t *testing.T) {
		field := DateTimeField{Null: true}
		if err := values.Set("created", time.Now(), field); err != nil {
			t.Fatal(err)
		}
		if _, ok := values["created"].(NullTime); !ok {
			t.Errorf("expected gomodels.NullTime, got %T", values["created"])
		}
	})
}
