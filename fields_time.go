package gomodels

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Date struct {
	Time  time.Time
	Valid bool
}

func (d *Date) Scan(value interface{}) error {
	if t, ok := value.(time.Time); ok {
		d.Time = t
		d.Valid = true
	} else if s, ok := value.(string); ok {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			d.Time = t
			d.Valid = true
		}
	}
	return nil
}

func (d Date) Value() (driver.Value, error) {
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

func (f DateField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f DateField) IsPk() bool {
	return f.PrimaryKey
}

func (f DateField) IsAuto() bool {
	return false
}

func (f DateField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f DateField) DefaultVal() (Value, bool) {
	if f.Default.Equal(time.Time{}) {
		return f.Default, false
	}
	return f.Default, true
}

func (f DateField) Recipient() interface{} {
	var val Date
	return &val
}

func (f DateField) SqlDatatype(dvr string) string {
	dt := fmt.Sprintf(
		"DATE %s", sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique),
	)
	if !f.Default.Equal(time.Time{}) {
		dt += fmt.Sprintf(" DEFAULT '%s'", f.Default.Format("2006-01-02"))
	}
	return dt
}

func (f DateField) Value(rec interface{}) Value {
	if vlr, ok := rec.(driver.Valuer); ok {
		if val, err := vlr.Value(); err == nil {
			if t, ok := val.(time.Time); ok {
				return t
			}
		}
	}
	return rec
}

func (f DateField) DriverValue(v Value, dvr string) (interface{}, error) {
	if vlr, ok := v.(driver.Valuer); ok {
		if val, err := vlr.Value(); err == nil {
			v = val
		}
	}
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02"), nil
	} else if s, ok := v.(string); ok {
		return s, nil
	}
	return v, fmt.Errorf("invalid value")
}

func (f DateField) IsAutoNow() bool {
	return f.AutoNow
}

func (f DateField) IsAutoNowAdd() bool {
	return f.AutoNowAdd
}
