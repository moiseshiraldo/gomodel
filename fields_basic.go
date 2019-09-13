package gomodel

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

// CharChoice holds a choice option for a CharField.
type CharChoice struct {
	Value string // Value is the choice value.
	Label string // Label is the choice label.
}

// CharField implements the Field interface for small to medium-sized strings.
type CharField struct {
	// PrimaryKey is true if the field is the model primary key.
	PrimaryKey bool `json:",omitempty"`
	// Unique is true if the field value must be unique.
	Unique bool `json:",omitempty"`
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// MaxLength is the max length accepted for field values.
	MaxLength int `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Choices is a list of possible choices for the field.
	Choices []CharChoice `json:",omitempty"`
	// Default is the default value for the field. Blank for no default.
	Default string `json:",omitempty"`
	// DefaultEmpty is true if the empty string is the field default value.
	DefaultEmpty bool `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f CharField) IsPK() bool {
	return f.PrimaryKey
}

// IsUnique implements the IsUnique method of the Field interface.
func (f CharField) IsUnique() bool {
	return f.Unique
}

// IsNull implements the IsNull method of the Field interface.
func (f CharField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f CharField) IsAuto() bool {
	return false
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f CharField) IsAutoNow() bool {
	return false
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f CharField) IsAutoNowAdd() bool {
	return false
}

// HasIndex implements the HasIndex method of the Field interface.
func (f CharField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

// DBColumn implements the DBColumn method of the Field interface.
func (f CharField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f CharField) DataType(driver string) string {
	return fmt.Sprintf("VARCHAR(%d)", f.MaxLength)
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f CharField) DefaultVal() (Value, bool) {
	if f.Default != "" || f.DefaultEmpty {
		return f.Default, true
	}
	return nil, false
}

// Recipient implements the Recipient method of the Field interface.
func (f CharField) Recipient() interface{} {
	if f.Null {
		var val sql.NullString
		return &val
	}
	var val string
	return &val
}

// Value implements the Value method of the Field interface.
func (f CharField) Value(rec interface{}) Value {
	if val, ok := rec.(sql.NullString); ok {
		if !val.Valid {
			return nil
		}
		return val.String
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f CharField) DriverValue(val Value, dvr string) (interface{}, error) {
	if vlr, ok := val.(driver.Valuer); ok {
		return vlr.Value()
	}
	return val, nil
}

// BooleanField implements the Field interface for true/false fields.
type BooleanField struct {
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Default is the default value for the field. Blank for no default.
	Default bool `json:",omitempty"`
	// DefaultFalse is true if false is the field default value.
	DefaultFalse bool `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f BooleanField) IsPK() bool {
	return false
}

// IsUnique implements the IsUnique method of the Field interface.
func (f BooleanField) IsUnique() bool {
	return false
}

// IsNull implements the IsNull method of the Field interface.
func (f BooleanField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f BooleanField) IsAuto() bool {
	return false
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f BooleanField) IsAutoNow() bool {
	return false
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f BooleanField) IsAutoNowAdd() bool {
	return false
}

// HasIndex implements the HasIndex method of the Field interface.
func (f BooleanField) HasIndex() bool {
	return f.Index
}

// DBColumn implements the DBColumn method of the Field interface.
func (f BooleanField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f BooleanField) DataType(dvr string) string {
	return "BOOLEAN"
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f BooleanField) DefaultVal() (Value, bool) {
	if f.Default {
		return true, true
	} else if f.DefaultFalse {
		return false, true
	} else {
		return nil, false
	}
}

// Recipient implements the Recipient method of the Field interface.
func (f BooleanField) Recipient() interface{} {
	if f.Null {
		var val sql.NullBool
		return &val
	}
	var val bool
	return &val
}

// Value implements the Value method of the Field interface.
func (f BooleanField) Value(rec interface{}) Value {
	if val, ok := rec.(sql.NullBool); ok {
		if !val.Valid {
			return nil
		}
		return val.Bool
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f BooleanField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		return vlr.Value()
	}
	return v, nil
}

// NullInt32 represents an int32 that may be null. TODO: remove for Golang 1.13
type NullInt32 struct {
	Int32 int32
	Valid bool // Valid is true if Int32 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullInt32) Scan(value interface{}) error {
	if value == nil {
		n.Int32, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	return setRecipient(&n.Int32, value)
}

// Value implements the driver Valuer interface.
func (n NullInt32) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return int64(n.Int32), nil
}

// IntChoice holds a choice option for an IntegerField.
type IntChoice struct {
	Value int    // Value is the Choice value.
	Label string // Label is che choice label.
}

// IntegerField implements the Field interface for small to medium-sized strings.
type IntegerField struct {
	// PrimaryKey is true if the field is the model primary key.
	PrimaryKey bool `json:",omitempty"`
	// Unique is true if the field value must be unique.
	Unique bool `json:",omitempty"`
	// Null is true if the field can have null values.
	Null bool `json:",omitempty"`
	// Auto is true if the field value will be auto incremented.
	Auto bool `json:",omitempty"`
	// Blank is true if the field is not required. Only used for validation.
	Blank bool `json:",omitempty"`
	// Index is true if the field column should be indexed.
	Index bool `json:",omitempty"`
	// Column is the name of the db column. If blank, it will be the field name.
	Column string `json:",omitempty"`
	// Choices is a list of possible choices for the field.
	Choices []IntChoice `json:",omitempty"`
	// Default is the default value for the field. Blank for no default.
	Default int32 `json:",omitempty"`
	// DefaultZero is true if zero is the field default value.
	DefaultZero bool `json:",omitempty"`
}

// IsPK implements the IsPK method of the Field interface.
func (f IntegerField) IsPK() bool {
	return f.PrimaryKey
}

// IsUnique implements the IsUnique method of the Field interface.
func (f IntegerField) IsUnique() bool {
	return f.Unique
}

// IsNull implements the IsNull method of the Field interface.
func (f IntegerField) IsNull() bool {
	return f.Null
}

// IsAuto implements the IsAuto method of the Field interface.
func (f IntegerField) IsAuto() bool {
	return f.Auto
}

// IsAutoNow implements the IsAutoNow method of the Field interface.
func (f IntegerField) IsAutoNow() bool {
	return false
}

// IsAutoNowAdd implements the IsAutoNowAdd method of the Field interface.
func (f IntegerField) IsAutoNowAdd() bool {
	return false
}

// HasIndex implements the HasIndex method of the Field interface.
func (f IntegerField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

// DBColumn implements the DBColumn method of the Field interface.
func (f IntegerField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

// DataType implements the DataType method of the Field interface.
func (f IntegerField) DataType(dvr string) string {
	if dvr == "postgres" && f.IsAuto() {
		return "SERIAL"
	}
	return "INTEGER"
}

// DefaultVal implements the DefaultVal method of the Field interface.
func (f IntegerField) DefaultVal() (Value, bool) {
	if f.Default != 0 || f.DefaultZero {
		return f.Default, true
	}
	return nil, false
}

// Recipient implements the Recipient method of the Field interface.
func (f IntegerField) Recipient() interface{} {
	if f.Null {
		var val NullInt32
		return &val
	}
	var val int32
	return &val
}

// Value implements the Value method of the Field interface.
func (f IntegerField) Value(rec interface{}) Value {
	if val, ok := rec.(NullInt32); ok {
		if !val.Valid {
			return nil
		}
		return val.Int32
	}
	return rec
}

// DriverValue implements the DriverValue method of the Field interface.
func (f IntegerField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		return vlr.Value()
	}
	return v, nil
}
