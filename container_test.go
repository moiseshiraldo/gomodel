package gomodels

import (
	"database/sql"
	"testing"
	"time"
)

type mockedField struct {
	CharField
	recipient interface{}
}

func (f mockedField) Recipient() interface{} {
	return f.recipient
}

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

// TestRecipients tests value conversions when setting recipients
func TestRecipients(t *testing.T) {
	field := mockedField{}
	values := Values{}

	t.Run("NilRecipient", func(t *testing.T) {
		field.recipient = nil
		if err := values.Set("test", "value", field); err == nil {
			t.Fatal("expected nil pointer error")
		}
	})

	t.Run("StringValue", func(t *testing.T) {
		var s string
		var b []byte
		var r sql.RawBytes
		recipients := []interface{}{&s, &b, &r}
		for _, recipient := range recipients {
			field.recipient = recipient
			if err := values.Set("test", "value", field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("ByteSliceValue", func(t *testing.T) {
		var s string
		var i interface{}
		var b []byte
		var r sql.RawBytes
		recipients := []interface{}{&s, &i, &b, &r}
		for _, recipient := range recipients {
			field.recipient = recipient
			if err := values.Set("test", []byte("value"), field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("TimeValue", func(t *testing.T) {
		var tm time.Time
		var s string
		var b []byte
		var r sql.RawBytes
		recipients := []interface{}{&tm, &s, &b, &r}
		for _, recipient := range recipients {
			field.recipient = recipient
			if err := values.Set("test", time.Now(), field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("NilValue", func(t *testing.T) {
		var i interface{}
		var b []byte
		var r sql.RawBytes
		recipients := []interface{}{&i, &b, &r}
		for _, recipient := range recipients {
			field.recipient = recipient
			if err := values.Set("test", nil, field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("StringRecipient", func(t *testing.T) {
		var s string
		field.recipient = &s
		source := []interface{}{
			int(7), int8(7), int16(7), int32(7), int64(7), uint(7), uint8(7),
			uint16(7), uint32(7), uint64(7), float32(7), float64(7), true,
		}
		for _, val := range source {
			if err := values.Set("test", val, field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("ByteSliceRecipient", func(t *testing.T) {
		var b []byte
		field.recipient = &b
		source := []interface{}{
			int(7), int8(7), int16(7), int32(7), int64(7), uint(7), uint8(7),
			uint16(7), uint32(7), uint64(7), float32(7), float64(7), true, "s",
		}
		for _, val := range source {
			if err := values.Set("test", val, field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("RawSqlBytesRecipient", func(t *testing.T) {
		var r sql.RawBytes
		field.recipient = &r
		source := []interface{}{
			int(7), int8(7), int16(7), int32(7), int64(7), uint(7), uint8(7),
			uint16(7), uint32(7), uint64(7), float32(7), float64(7), true, "s",
		}
		for _, val := range source {
			if err := values.Set("test", val, field); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("BoolRecipient", func(t *testing.T) {
		var b bool
		field.recipient = &b
		if err := values.Set("test", 1, field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("InterfaceRecipient", func(t *testing.T) {
		var i interface{}
		field.recipient = &i
		if err := values.Set("test", 1, field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ReflectValue", func(t *testing.T) {
		var i IntegerField
		field.recipient = &i
		if err := values.Set("test", IntegerField{}, field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("NilConversionError", func(t *testing.T) {
		var i IntegerField
		var u uint
		var f float32
		var s string
		recipients := []interface{}{&i, &u, &f, &s}
		for _, recipient := range recipients {
			field.recipient = recipient
			if err := values.Set("test", nil, field); err == nil {
				t.Fatal("expected unsupported conversion error")
			}
		}
	})

	t.Run("PointerRecipientNilValue", func(t *testing.T) {
		var p *string
		field.recipient = &p
		if err := values.Set("test", nil, field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerRecipient", func(t *testing.T) {
		var p *string
		field.recipient = &p
		if err := values.Set("test", "value", field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerIntRecipient", func(t *testing.T) {
		var p *int
		field.recipient = &p
		if err := values.Set("test", "42", field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerIntRecipientInvalidValue", func(t *testing.T) {
		var p *int
		field.recipient = &p
		if err := values.Set("test", "42s", field); err == nil {
			t.Fatal("expected conversion error")
		}
	})

	t.Run("PointerUIntRecipient", func(t *testing.T) {
		var p *uint
		field.recipient = &p
		if err := values.Set("test", "42", field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerUIntRecipientInvalidValue", func(t *testing.T) {
		var p *uint
		field.recipient = &p
		if err := values.Set("test", "42s", field); err == nil {
			t.Fatal("expected conversion error")
		}
	})

	t.Run("PointerFloatRecipient", func(t *testing.T) {
		var f *float64
		field.recipient = &f
		if err := values.Set("test", "42.0", field); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("PointerFloatRecipientInvalidValue", func(t *testing.T) {
		var f *float64
		field.recipient = &f
		if err := values.Set("test", "42s", field); err == nil {
			t.Fatal("expected conversion error")
		}
	})

}
