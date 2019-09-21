package gomodel

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// NullTime represents a time.Time that may be null.
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the Scanner interface.
func (d *NullTime) Scan(value interface{}) error {
	if value == nil {
		d.Time, d.Valid = time.Time{}, false
		return nil
	} else if t, ok := value.(time.Time); ok {
		d.Time = t
		d.Valid = true
		return nil
	} else if s, ok := value.(string); ok {
		value = s
	}
	d.Valid = true
	return setRecipient(&d.Time, value)
}

// Value implements the driver Valuer interface.
func (d NullTime) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time, nil
}

// timeFieldMarshaler is used to deal with time.Time zero value for JSON.
type timeFieldMarshaler struct {
	PrimaryKey bool     `json:",omitempty"`
	Unique     bool     `json:",omitempty"`
	Null       bool     `json:",omitempty"`
	AutoNow    bool     `json:",omitempty"`
	AutoNowAdd bool     `json:",omitempty"`
	Blank      bool     `json:",omitempty"`
	Index      bool     `json:",omitempty"`
	Column     string   `json:",omitempty"`
	Choices    []Choice `json:",omitempty"`
	Default    string   `json:",omitempty"`
}

// DateField implements the Field interface for dates.
type DateField struct {
	// PrimaryKey is true if the field is the model primary key.
	PrimaryKey bool `json:",omitempty"`
	// Unique is true if the field value must be unique.
	Unique bool `json:",omitempty"`
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// AutoNow is true if the value is auto generated on row updates.
	AutoNow bool `json:",omitempty"`
	// AutoNowAdd is true if the value is auto generated on row inserts.
	AutoNowAdd bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Choices is a list of possible choices for the field.
	Choices []Choice `json:",omitempty"`
	// Default is the default value for the field. Blank for no default
	Default time.Time `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f DateField) IsPK() bool {
	return f.PrimaryKey
}

// IsUnique implements the IsUnique method of the Field interface.
func (f DateField) IsUnique() bool {
	return f.Unique
}

// IsNull implements the IsNull method of the Field interface.
func (f DateField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f DateField) IsAuto() bool {
	return false
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f DateField) IsAutoNow() bool {
	return f.AutoNow
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f DateField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

// HasIndex implements the HasIndex method of the Field interface.
func (f DateField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

// DBColumn implements the DBColumn method of the Field interface.
func (f DateField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f DateField) DataType(dvr string) string {
	return "DATE"
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f DateField) DefaultVal() (Value, bool) {
	if f.Default.IsZero() {
		return f.Default, false
	}
	return f.Default, true
}

// Recipient implements the Recipient method of the Field interface.
func (f DateField) Recipient() interface{} {
	var val NullTime
	return &val
}

// Value implements the Value method of the Field interface.
func (f DateField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f DateField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		if val, err := vlr.Value(); err == nil {
			v = val
		}
	}
	if v == nil {
		return v, nil
	} else if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02"), nil
	} else if s, ok := v.(string); ok {
		return s, nil
	}
	return v, fmt.Errorf("invalid value")
}

// DisplayValue implements the DisplayValue method of the Field interface.
func (f DateField) DisplayValue(val Value) string {
	val = f.Value(val)
	for _, choice := range f.Choices {
		if reflect.DeepEqual(val, choice.Value) {
			return choice.Label
		}
	}
	if t, ok := val.(time.Time); ok {
		return t.Format("2006-01-02")
	}
	return fmt.Sprintf("%v", val)
}

// MarshalJSON implements the json.Marshaler interface.
func (f DateField) MarshalJSON() ([]byte, error) {
	result := timeFieldMarshaler{
		PrimaryKey: f.PrimaryKey,
		Unique:     f.Unique,
		Null:       f.Null,
		AutoNow:    f.AutoNow,
		AutoNowAdd: f.AutoNowAdd,
		Blank:      f.Blank,
		Index:      f.Index,
		Column:     f.Column,
		Choices:    f.Choices,
	}
	if !f.Default.IsZero() {
		result.Default = f.Default.Format("2006-01-02")
	}
	return json.Marshal(result)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *DateField) UnmarshalJSON(data []byte) error {
	r := timeFieldMarshaler{}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if r.Default != "" {
		t, err := time.Parse("2006-01-02", r.Default)
		if err != nil {
			return err
		}
		f.Default = t
	}
	f.PrimaryKey = r.PrimaryKey
	f.Unique = r.Unique
	f.Null = r.Null
	f.AutoNow = r.AutoNow
	f.AutoNowAdd = r.AutoNowAdd
	f.Blank = r.Blank
	f.Index = r.Index
	f.Column = r.Column
	f.Choices = r.Choices
	return nil
}

// TimeField implements the Field interface for time values.
type TimeField struct {
	// PrimaryKey is true if the field is the model primary key.
	PrimaryKey bool `json:",omitempty"`
	// Unique is true if the field value must be unique.
	Unique bool `json:",omitempty"`
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// AutoNow is true if the value is auto generated on row updates.
	AutoNow bool `json:",omitempty"`
	// AutoNowAdd is true if the value is auto generated on row inserts.
	AutoNowAdd bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Choices is a list of possible choices for the field.
	Choices []Choice `json:",omitempty"`
	// Default is the default value for the field. Blank for no default
	Default time.Time `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f TimeField) IsPK() bool {
	return f.PrimaryKey
}

// IsUnique implements the IsUnique method of the Field interface.
func (f TimeField) IsUnique() bool {
	return f.Unique
}

// IsNull implements the IsNull method of the Field interface.
func (f TimeField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f TimeField) IsAuto() bool {
	return false
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f TimeField) IsAutoNow() bool {
	return f.AutoNow
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f TimeField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

// HasIndex implements the HasIndex method of the Field interface.
func (f TimeField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

// DBColumn implements the DBColumn method of the Field interface.
func (f TimeField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f TimeField) DataType(dvr string) string {
	return "TIME"
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f TimeField) DefaultVal() (Value, bool) {
	if f.Default.IsZero() {
		return f.Default, false
	}
	return f.Default, true
}

// Recipient implements the Recipient method of the Field interface.
func (f TimeField) Recipient() interface{} {
	var val NullTime
	return &val
}

// Value implements the Value method of the Field interface.
func (f TimeField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f TimeField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		if val, err := vlr.Value(); err == nil {
			v = val
		}
	}
	if v == nil {
		return v, nil
	} else if t, ok := v.(time.Time); ok {
		return t.Format("15:04:05"), nil
	} else if s, ok := v.(string); ok {
		return s, nil
	}
	return v, fmt.Errorf("invalid value")
}

// DisplayValue implements the DisplayValue method of the Field interface.
func (f TimeField) DisplayValue(val Value) string {
	val = f.Value(val)
	for _, choice := range f.Choices {
		if reflect.DeepEqual(val, choice.Value) {
			return choice.Label
		}
	}
	if t, ok := val.(time.Time); ok {
		return t.Format("15:04:05")
	}
	return fmt.Sprintf("%v", val)
}

// MarshalJSON implements the json.Marshaler interface.
func (f TimeField) MarshalJSON() ([]byte, error) {
	result := timeFieldMarshaler{
		PrimaryKey: f.PrimaryKey,
		Unique:     f.Unique,
		Null:       f.Null,
		AutoNow:    f.AutoNow,
		AutoNowAdd: f.AutoNowAdd,
		Blank:      f.Blank,
		Index:      f.Index,
		Column:     f.Column,
		Choices:    f.Choices,
	}
	if !f.Default.IsZero() {
		result.Default = f.Default.Format("15:04:05.999999999Z07:00")
	}
	return json.Marshal(result)
}

func (f *TimeField) UnmarshalJSON(data []byte) error {
	r := timeFieldMarshaler{}
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	if r.Default != "" {
		t, err := time.Parse("15:04:05.999999999Z07:00", r.Default)
		if err != nil {
			return err
		}
		f.Default = t
	}
	f.PrimaryKey = r.PrimaryKey
	f.Unique = r.Unique
	f.Null = r.Null
	f.AutoNow = r.AutoNow
	f.AutoNowAdd = r.AutoNowAdd
	f.Blank = r.Blank
	f.Index = r.Index
	f.Column = r.Column
	f.Choices = r.Choices
	return nil
}

// DateTimeField implements the Field interface for datetime values.
type DateTimeField struct {
	// PrimaryKey is true if the field is the model primary key.
	PrimaryKey bool `json:",omitempty"`
	// Unique is true if the field value must be unique.
	Unique bool `json:",omitempty"`
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// AutoNow is true if the value is auto generated on row updates.
	AutoNow bool `json:",omitempty"`
	// AutoNowAdd is true if the value is auto generated on row inserts.
	AutoNowAdd bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Choices is a list of possible choices for the field.
	Choices []Choice `json:",omitempty"`
	// Default is the default value for the field. Blank for no default
	Default time.Time `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f DateTimeField) IsPK() bool {
	return f.PrimaryKey
}

// IsUnique implements the IsUnique method of the Field interface.
func (f DateTimeField) IsUnique() bool {
	return f.Unique
}

// IsNull implements the IsNull method of the Field interface.
func (f DateTimeField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f DateTimeField) IsAuto() bool {
	return false
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f DateTimeField) IsAutoNow() bool {
	return f.AutoNow
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f DateTimeField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

// HasIndex implements the HasIndex method of the Field interface.
func (f DateTimeField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

// DBColumn implements the DBColumn method of the Field interface.
func (f DateTimeField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f DateTimeField) DataType(dvr string) string {
	if dvr == "postgres" {
		return "TIMESTAMP"
	}
	return "DATETIME"
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f DateTimeField) DefaultVal() (Value, bool) {
	if f.Default.IsZero() {
		return f.Default, false
	}
	return f.Default, true
}

// Recipient implements the Recipient method of the Field interface.
func (f DateTimeField) Recipient() interface{} {
	var val NullTime
	return &val
}

// Value implements the Value method of the Field interface.
func (f DateTimeField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f DateTimeField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		if val, err := vlr.Value(); err == nil {
			v = val
		}
	}
	if v == nil {
		return v, nil
	} else if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05"), nil
	} else if s, ok := v.(string); ok {
		return s, nil
	}
	return v, fmt.Errorf("invalid value")
}

// DisplayValue implements the DisplayValue method of the Field interface.
func (f DateTimeField) DisplayValue(val Value) string {
	val = f.Value(val)
	for _, choice := range f.Choices {
		if reflect.DeepEqual(val, choice.Value) {
			return choice.Label
		}
	}
	if t, ok := val.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("%v", val)
}

// MarshalJSON implements the json.Marshaler interface.
func (f DateTimeField) MarshalJSON() ([]byte, error) {
	result := timeFieldMarshaler{
		PrimaryKey: f.PrimaryKey,
		Unique:     f.Unique,
		Null:       f.Null,
		AutoNow:    f.AutoNow,
		AutoNowAdd: f.AutoNowAdd,
		Blank:      f.Blank,
		Index:      f.Index,
		Column:     f.Column,
		Choices:    f.Choices,
	}
	if !f.Default.IsZero() {
		result.Default = f.Default.Format(time.RFC3339Nano)
	}
	return json.Marshal(result)
}
