package gomodel

import (
	"encoding/json"
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
			t.Errorf("expected *gomodel.NullTime, got %T", recipient)
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
		recipient := NullTime{
			Time:  time.Date(2019, 8, 24, 0, 0, 0, 0, time.UTC),
			Valid: true,
		}
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

	t.Run("DriverValueInvalid", func(t *testing.T) {
		recipient := 42
		if _, err := field.DriverValue(recipient, "sqlite3"); err == nil {
			t.Error("expected invalid value error")
		}
	})

	t.Run("DisplayValue", func(t *testing.T) {
		recipient := time.Date(2019, 8, 24, 0, 0, 0, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "2019-08-24" {
			t.Errorf("expected 2019-08-24, got %s", value)
		}
	})

	t.Run("DisplayValueNoTime", func(t *testing.T) {
		recipient := 30
		value := field.DisplayValue(recipient)
		if value != "30" {
			t.Errorf("expected 30, got %s", value)
		}
	})

	t.Run("DisplayValueChoice", func(t *testing.T) {
		field.Choices = []Choice{
			{time.Date(2019, 8, 24, 0, 0, 0, 0, time.UTC), "Date One"},
			{time.Date(2019, 10, 17, 0, 0, 0, 0, time.UTC), "Date Two"},
		}
		recipient := time.Date(2019, 10, 17, 0, 0, 0, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "Date Two" {
			t.Errorf("expected Date Two, got %s", value)
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		f := DateField{Default: time.Date(2019, 8, 24, 0, 0, 0, 0, time.UTC)}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Default":"2019-08-24"}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("MarshalDefaultZero", func(t *testing.T) {
		f := DateField{Null: true}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Null":true}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("UnmarshalField", func(t *testing.T) {
		f := DateField{}
		data := []byte(`{"Default":"2019-08-24"}`)
		if err := json.Unmarshal(data, &f); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("UnmarshalInvalidTime", func(t *testing.T) {
		f := DateField{}
		data := []byte(`{"Default":"invalid"}`)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatal("expected time parsing error")
		}
	})

	t.Run("UnmarshalInvalidData", func(t *testing.T) {
		f := DateField{}
		data := []byte(`{"PrimaryKey": "invalid"}`)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatal("expected parsing error")
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
			t.Errorf("expected *gomodel.NullTime, got %T", recipient)
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

	t.Run("DriverValueInvalid", func(t *testing.T) {
		recipient := 42
		if _, err := field.DriverValue(recipient, "sqlite3"); err == nil {
			t.Error("expected invalid value error")
		}
	})

	t.Run("DisplayValue", func(t *testing.T) {
		recipient := time.Date(0, 0, 0, 14, 18, 3, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "14:18:03" {
			t.Errorf("expected 14:18:03, got %s", value)
		}
	})

	t.Run("DisplayValueNoTime", func(t *testing.T) {
		recipient := 30
		value := field.DisplayValue(recipient)
		if value != "30" {
			t.Errorf("expected 30, got %s", value)
		}
	})

	t.Run("DisplayValueChoice", func(t *testing.T) {
		field.Choices = []Choice{
			{time.Date(0, 0, 0, 14, 18, 3, 0, time.UTC), "Time One"},
			{time.Date(0, 0, 0, 18, 05, 51, 0, time.UTC), "Time Two"},
		}
		recipient := time.Date(0, 0, 0, 18, 05, 51, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "Time Two" {
			t.Errorf("expected Time Two, got %s", value)
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		f := TimeField{Default: time.Date(0, 0, 0, 14, 18, 3, 0, time.UTC)}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Default":"14:18:03Z"}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("MarshalDefaultZero", func(t *testing.T) {
		f := TimeField{Null: true}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Null":true}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("UnmarshalField", func(t *testing.T) {
		f := TimeField{}
		data := []byte(`{"Default":"14:18:03Z"}`)
		err := json.Unmarshal(data, &f)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("UnmarshalInvalidTime", func(t *testing.T) {
		f := TimeField{}
		data := []byte(`{"Default":"invalid"}`)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatal("expected time parsing error")
		}
	})

	t.Run("UnmarshalInvalidData", func(t *testing.T) {
		f := TimeField{}
		data := []byte(`{"PrimaryKey": "invalid"}`)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatal("expected parsing error")
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
			t.Errorf("expected *gomodel.NullTime, got %T", recipient)
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

	t.Run("DriverValueInvalid", func(t *testing.T) {
		recipient := 42
		if _, err := field.DriverValue(recipient, "sqlite3"); err == nil {
			t.Error("expected invalid value error")
		}
	})

	t.Run("DisplayValue", func(t *testing.T) {
		recipient := time.Date(2019, 8, 24, 14, 18, 03, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "2019-08-24 14:18:03" {
			t.Errorf("expected 2019-08-24 14:18:03, got %s", value)
		}
	})

	t.Run("DisplayValueNoTime", func(t *testing.T) {
		recipient := 30
		value := field.DisplayValue(recipient)
		if value != "30" {
			t.Errorf("expected 30, got %s", value)
		}
	})

	t.Run("DisplayValueChoice", func(t *testing.T) {
		field.Choices = []Choice{
			{time.Date(2019, 8, 24, 14, 18, 03, 0, time.UTC), "Datetime One"},
			{time.Date(2019, 11, 23, 10, 54, 12, 0, time.UTC), "Datetime Two"},
		}
		recipient := time.Date(2019, 11, 23, 10, 54, 12, 0, time.UTC)
		value := field.DisplayValue(recipient)
		if value != "Datetime Two" {
			t.Errorf("expected Datetime Two, got %s", value)
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		f := DateTimeField{
			Default: time.Date(2019, 8, 24, 14, 18, 03, 0, time.UTC),
		}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Default":"2019-08-24T14:18:03Z"}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("MarshalDefaultZero", func(t *testing.T) {
		f := DateTimeField{Null: true}
		data, err := f.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"Null":true}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})

	t.Run("UnmarshalField", func(t *testing.T) {
		f := DateTimeField{}
		data := []byte(`{"Default":"2019-08-24T14:18:03Z"}`)
		err := json.Unmarshal(data, &f)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("UnmarshalInvalidTime", func(t *testing.T) {
		f := DateTimeField{}
		data := []byte(`{"Default":"invalid"}`)
		if err := json.Unmarshal(data, &f); err == nil {
			t.Fatal("expected time parsing error")
		}
	})
}
