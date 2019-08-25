package gomodels

import (
	"database/sql"
	"testing"
)

// TestCharField tests the CharField struct methods
func TestCharField(t *testing.T) {
	field := CharField{}

	t.Run("IsPK", func(t *testing.T) {
		field.PrimaryKey = true
		if !field.IsPK() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsUnique", func(t *testing.T) {
		field.Unique = true
		if !field.IsUnique() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		field.Null = true
		if !field.IsNull() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAuto", func(t *testing.T) {
		if field.IsAuto() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsAutoNow", func(t *testing.T) {
		if field.IsAutoNow() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		if field.IsAutoNowAdd() {
			t.Error("expected false, got true")
		}
	})

	t.Run("HasIndex", func(t *testing.T) {
		field.Unique = false
		field.PrimaryKey = false
		field.Index = true
		if !field.HasIndex() {
			t.Error("expected true, got false")
		}
	})

	t.Run("DefaultColumn", func(t *testing.T) {
		if field.DBColumn("foo") != "foo" {
			t.Errorf("expected foo, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("CustomColumn", func(t *testing.T) {
		field.Column = "bar"
		if field.DBColumn("foo") != "bar" {
			t.Errorf("expected bar, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("DataType", func(t *testing.T) {
		field.MaxLength = 100
		if field.DataType("sqlite3") != "VARCHAR(100)" {
			t.Errorf("expected VARCHAR(100), got %s", field.DataType("sqlite3"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("Default", func(t *testing.T) {
		field.Default = "foo"
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(string); !ok || val != "foo" {
			t.Errorf("expected foo, got %s", val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*string); !ok {
			t.Errorf("expected *string, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*sql.NullString); !ok {
			t.Errorf("expected *sql.NullString, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := "foo"
		value := field.Value(recipient)
		if v, ok := value.(string); !ok || v != "foo" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := sql.NullString{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := sql.NullString{String: "foo", Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(string); !ok || v != "foo" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValue", func(t *testing.T) {
		recipient := "foo"
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "foo" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := sql.NullString{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})
}

// TestBooleanField tests the BooleanField struct methods
func TestBooleanField(t *testing.T) {
	field := BooleanField{}

	t.Run("IsPK", func(t *testing.T) {
		if field.IsPK() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsUnique", func(t *testing.T) {
		if field.IsUnique() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		field.Null = true
		if !field.IsNull() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAuto", func(t *testing.T) {
		if field.IsAuto() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsAutoNow", func(t *testing.T) {
		if field.IsAutoNow() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		if field.IsAutoNowAdd() {
			t.Error("expected false, got true")
		}
	})

	t.Run("HasIndex", func(t *testing.T) {
		field.Index = true
		if !field.HasIndex() {
			t.Error("expected true, got false")
		}
	})

	t.Run("DefaultColumn", func(t *testing.T) {
		if field.DBColumn("foo") != "foo" {
			t.Errorf("expected foo, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("CustomColumn", func(t *testing.T) {
		field.Column = "bar"
		if field.DBColumn("foo") != "bar" {
			t.Errorf("expected bar, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("DataType", func(t *testing.T) {
		if field.DataType("sqlite3") != "BOOLEAN" {
			t.Errorf("expected BOOLEAN, got %s", field.DataType("sqlite3"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("DefaultTrue", func(t *testing.T) {
		field.Default = true
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(bool); !ok || !val {
			t.Errorf("expected true, got %t", val)
		}
	})

	t.Run("DefaultFalse", func(t *testing.T) {
		field.Default = false
		field.DefaultFalse = true
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(bool); !ok || val {
			t.Errorf("expected false, got %t", val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*bool); !ok {
			t.Errorf("expected *bool, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*sql.NullBool); !ok {
			t.Errorf("expected *sql.NullBool, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := true
		value := field.Value(recipient)
		if v, ok := value.(bool); !ok || !v {
			t.Errorf("expected true, got %t", v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := sql.NullBool{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %t", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := sql.NullBool{Bool: true, Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(bool); !ok || !v {
			t.Errorf("expected true, got %t", v)
		}
	})

	t.Run("DriverValue", func(t *testing.T) {
		recipient := true
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(bool); !ok || !v {
			t.Errorf("expected true, got %t", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := sql.NullBool{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})
}

// TestIntegerField tests the IntegerField struct methods
func TestIntegerField(t *testing.T) {
	field := IntegerField{}

	t.Run("IsPK", func(t *testing.T) {
		field.PrimaryKey = true
		if !field.IsPK() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsUnique", func(t *testing.T) {
		field.Unique = true
		if !field.IsUnique() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsNull", func(t *testing.T) {
		field.Null = true
		if !field.IsNull() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAuto", func(t *testing.T) {
		field.Auto = true
		if !field.IsAuto() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAutoNow", func(t *testing.T) {
		if field.IsAutoNow() {
			t.Error("expected false, got true")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		if field.IsAutoNowAdd() {
			t.Error("expected false, got true")
		}
	})

	t.Run("HasIndex", func(t *testing.T) {
		field.Unique = false
		field.PrimaryKey = false
		field.Index = true
		if !field.HasIndex() {
			t.Error("expected true, got false")
		}
	})

	t.Run("DefaultColumn", func(t *testing.T) {
		if field.DBColumn("foo") != "foo" {
			t.Errorf("expected foo, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("CustomColumn", func(t *testing.T) {
		field.Column = "bar"
		if field.DBColumn("foo") != "bar" {
			t.Errorf("expected bar, got %s", field.DBColumn("foo"))
		}
	})

	t.Run("DataType", func(t *testing.T) {
		field.Auto = false
		if field.DataType("sqlite3") != "INTEGER" {
			t.Errorf("expected INTEGER, got %s", field.DataType("sqlite3"))
		}
	})

	t.Run("DataTypeAuto", func(t *testing.T) {
		field.Auto = true
		if field.DataType("postgres") != "SERIAL" {
			t.Errorf("expected SERIAL, got %s", field.DataType("postgres"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("Default", func(t *testing.T) {
		field.Default = 42
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(int32); !ok || val != 42 {
			t.Errorf("expected 42, got %d", val)
		}
	})

	t.Run("DefaultZero", func(t *testing.T) {
		field.Default = 0
		field.DefaultZero = true
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(int32); !ok || val != 0 {
			t.Errorf("expected 0, got %d", val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*int32); !ok {
			t.Errorf("expected *int32, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*NullInt32); !ok {
			t.Errorf("expected *gomodels.NullInt32, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := int32(42)
		value := field.Value(recipient)
		if v, ok := value.(int32); !ok || v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := NullInt32{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := NullInt32{Int32: 42, Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(int32); !ok || v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	})

	t.Run("DriverValue", func(t *testing.T) {
		recipient := int32(42)
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(int32); !ok || v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := NullInt32{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("DriverValueNotNull", func(t *testing.T) {
		recipient := NullInt32{Int32: 42, Valid: true}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(int64); !ok || v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	})
}
