package gomodels

import (
	"testing"
	"time"
)

// TestDateField tests the DateField struct methods
func TestDateField(t *testing.T) {
	field := DateField{}

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
		field.AutoNow = true
		if !field.IsAutoNow() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		field.AutoNowAdd = true
		if !field.IsAutoNowAdd() {
			t.Error("expected true, got false")
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
		if field.DataType("sqlite3") != "DATE" {
			t.Errorf("expected DATE, got %s", field.DataType("sqlite3"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("Default", func(t *testing.T) {
		field.Default = time.Now()
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(time.Time); !ok || !val.Equal(field.Default) {
			t.Errorf("expected %s, got %s", field.Default, val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*time.Time); !ok {
			t.Errorf("expected *time.Time, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*NullTime); !ok {
			t.Errorf("expected *gomodels.NullTime, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := time.Now()
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient) {
			t.Errorf("expected %s, got %s", recipient, v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := NullTime{Time: time.Now(), Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient.Time) {
			t.Errorf("expected %s, got %s", recipient.Time, v)
		}
	})

	t.Run("DriverValueTime", func(t *testing.T) {
		recipient := time.Date(2019, 8, 24, 0, 0, 0, 0, time.UTC)
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "2019-08-24" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueString", func(t *testing.T) {
		recipient := "2019-08-24"
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "2019-08-24" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})
}

// TestTimeField tests the TimeField struct methods
func TestTimeField(t *testing.T) {
	field := TimeField{}

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
		field.AutoNow = true
		if !field.IsAutoNow() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		field.AutoNowAdd = true
		if !field.IsAutoNowAdd() {
			t.Error("expected true, got false")
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
		if field.DataType("sqlite3") != "TIME" {
			t.Errorf("expected TIME, got %s", field.DataType("sqlite3"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("Default", func(t *testing.T) {
		field.Default = time.Now()
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(time.Time); !ok || !val.Equal(field.Default) {
			t.Errorf("expected %s, got %s", field.Default, val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*time.Time); !ok {
			t.Errorf("expected *time.Time, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*NullTime); !ok {
			t.Errorf("expected *gomodels.NullTime, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := time.Now()
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient) {
			t.Errorf("expected %s, got %s", recipient, v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := NullTime{Time: time.Now(), Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient.Time) {
			t.Errorf("expected %s, got %s", recipient.Time, v)
		}
	})

	t.Run("DriverValueTime", func(t *testing.T) {
		recipient := time.Date(0, 0, 0, 14, 18, 3, 0, time.UTC)
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "14:18:03" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueString", func(t *testing.T) {
		recipient := "14:18:03"
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "14:18:03" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})
}

// TestDateTimeField tests the DateTimeField struct methods
func TestDateTimeField(t *testing.T) {
	field := DateTimeField{}

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
		field.AutoNow = true
		if !field.IsAutoNow() {
			t.Error("expected true, got false")
		}
	})

	t.Run("IsAutoNowAdd", func(t *testing.T) {
		field.AutoNowAdd = true
		if !field.IsAutoNowAdd() {
			t.Error("expected true, got false")
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
		if field.DataType("sqlite3") != "DATETIME" {
			t.Errorf("expected DATE, got %s", field.DataType("sqlite3"))
		}
		if field.DataType("postgres") != "TIMESTAMP" {
			t.Errorf("expected TIMESTAMP, got %s", field.DataType("postgres"))
		}
	})

	t.Run("NoDefault", func(t *testing.T) {
		if _, ok := field.DefaultVal(); ok {
			t.Error("expected no default value")
		}
	})

	t.Run("Default", func(t *testing.T) {
		field.Default = time.Now()
		val, ok := field.DefaultVal()
		if !ok {
			t.Fatal("expected default value")
		}
		if val, ok := val.(time.Time); !ok || !val.Equal(field.Default) {
			t.Errorf("expected %s, got %s", field.Default, val)
		}
	})

	t.Run("Recipient", func(t *testing.T) {
		field.Null = false
		recipient := field.Recipient()
		if _, ok := recipient.(*time.Time); !ok {
			t.Errorf("expected *time.Time, got %T", recipient)
		}
	})

	t.Run("NullRecipient", func(t *testing.T) {
		field.Null = true
		recipient := field.Recipient()
		if _, ok := recipient.(*NullTime); !ok {
			t.Errorf("expected *gomodels.NullTime, got %T", recipient)
		}
	})

	t.Run("Value", func(t *testing.T) {
		recipient := time.Now()
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient) {
			t.Errorf("expected %s, got %s", recipient, v)
		}
	})

	t.Run("ValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value := field.Value(recipient)
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})

	t.Run("ValueNotNull", func(t *testing.T) {
		recipient := NullTime{Time: time.Now(), Valid: true}
		value := field.Value(recipient)
		if v, ok := value.(time.Time); !ok || !v.Equal(recipient.Time) {
			t.Errorf("expected %s, got %s", recipient.Time, v)
		}
	})

	t.Run("DriverValueTime", func(t *testing.T) {
		recipient := time.Date(2019, 8, 24, 14, 18, 03, 0, time.UTC)
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "2019-08-24 14:18:03" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueString", func(t *testing.T) {
		recipient := "2019-08-24 14:18:03"
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if v, ok := value.(string); !ok || v != "2019-08-24 14:18:03" {
			t.Errorf("expected foo, got %s", v)
		}
	})

	t.Run("DriverValueNull", func(t *testing.T) {
		recipient := NullTime{Valid: false}
		value, err := field.DriverValue(recipient, "sqlite3")
		if err != nil {
			t.Fatal(err)
		}
		if value != nil {
			t.Errorf("expected nil, got %s", value)
		}
	})
}
