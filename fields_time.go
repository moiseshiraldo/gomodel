package gomodel

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (d *NullTime) Scan(value interface{}) error {
	if t, ok := value.(time.Time); ok {
		d.Time = t
		d.Valid = true
		return nil
	} else if s, ok := value.(string); ok {
		if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
			d.Time = t
			d.Valid = true
			return nil
		}
		if t, err := time.Parse("2006-01-02", s); err == nil {
			d.Time = t
			d.Valid = true
			return nil
		}
		if t, err := time.Parse("15:04:05", s); err == nil {
			d.Time = t
			d.Valid = true
			return nil
		}
	}
	return fmt.Errorf("cannot parse %T into gomodel.NullTime", value)
}

func (d NullTime) Value() (driver.Value, error) {
	if !d.Valid {
		return nil, nil
	}
	return d.Time, nil
}

type TimeChoice struct {
	Value time.Time
	Label string
}

type DateField struct {
	Null       bool         `json:",omitempty"`
	Blank      bool         `json:",omitempty"`
	Choices    []TimeChoice `json:",omitempty"`
	Column     string       `json:",omitempty"`
	Index      bool         `json:",omitempty"`
	Default    time.Time    `json:",omitempty"`
	PrimaryKey bool         `json:",omitempty"`
	Unique     bool         `json:",omitempty"`
	AutoNow    bool         `json:",omitempty"`
	AutoNowAdd bool         `json:",omitempty"`
}

func (f DateField) IsPK() bool {
	return f.PrimaryKey
}

func (f DateField) IsUnique() bool {
	return f.Unique
}

func (f DateField) IsNull() bool {
	return f.Null
}

func (f DateField) IsAuto() bool {
	return false
}

func (f DateField) IsAutoNow() bool {
	return f.AutoNow
}

func (f DateField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

func (f DateField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f DateField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f DateField) DataType(dvr string) string {
	return "DATE"
}

func (f DateField) DefaultVal() (Value, bool) {
	if f.Default.Equal(time.Time{}) {
		return f.Default, false
	}
	return f.Default, true
}

func (f DateField) Recipient() interface{} {
	if f.Null {
		var val NullTime
		return &val
	}
	var val time.Time
	return &val
}

func (f DateField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

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

type TimeField struct {
	Null       bool         `json:",omitempty"`
	Blank      bool         `json:",omitempty"`
	Choices    []TimeChoice `json:",omitempty"`
	Column     string       `json:",omitempty"`
	Index      bool         `json:",omitempty"`
	Default    time.Time    `json:",omitempty"`
	PrimaryKey bool         `json:",omitempty"`
	Unique     bool         `json:",omitempty"`
	AutoNow    bool         `json:",omitempty"`
	AutoNowAdd bool         `json:",omitempty"`
}

func (f TimeField) IsPK() bool {
	return f.PrimaryKey
}

func (f TimeField) IsUnique() bool {
	return f.Unique
}

func (f TimeField) IsNull() bool {
	return f.Null
}

func (f TimeField) IsAuto() bool {
	return false
}

func (f TimeField) IsAutoNow() bool {
	return f.AutoNow
}

func (f TimeField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

func (f TimeField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f TimeField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f TimeField) DataType(dvr string) string {
	return "TIME"
}

func (f TimeField) DefaultVal() (Value, bool) {
	if f.Default.Equal(time.Time{}) {
		return f.Default, false
	}
	return f.Default, true
}

func (f TimeField) Recipient() interface{} {
	if f.Null {
		var val NullTime
		return &val
	}
	var val time.Time
	return &val
}

func (f TimeField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

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

type DateTimeField struct {
	Null       bool         `json:",omitempty"`
	Blank      bool         `json:",omitempty"`
	Choices    []TimeChoice `json:",omitempty"`
	Column     string       `json:",omitempty"`
	Index      bool         `json:",omitempty"`
	Default    time.Time    `json:",omitempty"`
	PrimaryKey bool         `json:",omitempty"`
	Unique     bool         `json:",omitempty"`
	AutoNow    bool         `json:",omitempty"`
	AutoNowAdd bool         `json:",omitempty"`
}

func (f DateTimeField) IsPK() bool {
	return f.PrimaryKey
}

func (f DateTimeField) IsUnique() bool {
	return f.Unique
}

func (f DateTimeField) IsNull() bool {
	return f.Null
}

func (f DateTimeField) IsAuto() bool {
	return false
}

func (f DateTimeField) IsAutoNow() bool {
	return f.AutoNow
}

func (f DateTimeField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}

func (f DateTimeField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f DateTimeField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f DateTimeField) DataType(dvr string) string {
	if dvr == "postgres" {
		return "TIMESTAMP"
	} else {
		return "DATETIME"
	}
}

func (f DateTimeField) DefaultVal() (Value, bool) {
	if f.Default.Equal(time.Time{}) {
		return f.Default, false
	}
	return f.Default, true
}

func (f DateTimeField) Recipient() interface{} {
	if f.Null {
		var val NullTime
		return &val
	}
	var val time.Time
	return &val
}

func (f DateTimeField) Value(rec interface{}) Value {
	if val, ok := rec.(NullTime); ok {
		if !val.Valid {
			return nil
		}
		return val.Time
	}
	return rec
}

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
