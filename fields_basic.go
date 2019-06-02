package gomodels

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type CharChoice struct {
	Value string
	Label string
}

type CharField struct {
	Null         bool         `json:",omitempty"`
	Blank        bool         `json:",omitempty"`
	Choices      []CharChoice `json:",omitempty"`
	Column       string       `json:",omitempty"`
	Index        bool         `json:",omitempty"`
	Default      string       `json:",omitempty"`
	DefaultEmpty bool         `json:",omitempty"`
	PrimaryKey   bool         `json:",omitempty"`
	Unique       bool         `json:",omitempty"`
	MaxLength    int          `json:",omitempty"`
}

func (f CharField) FromJSON(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}

func (f CharField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f CharField) IsPk() bool {
	return f.PrimaryKey
}

func (f CharField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f CharField) DefaultVal() (bool, Value) {
	if f.Default != "" || f.DefaultEmpty {
		return true, f.Default
	} else {
		return false, nil
	}
}

func (f CharField) Recipient() interface{} {
	if f.Null {
		var val sql.NullString
		return &val
	}
	var val string
	return &val
}

func (f CharField) CreateSQL() string {
	query := fmt.Sprintf("varchar(%d)", f.MaxLength)
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	if f.Default != "" || f.DefaultEmpty {
		query += fmt.Sprintf(" DEFAULT '%s'", f.Default)
	}
	return query
}

type BooleanField struct {
	Null         bool   `json:",omitempty"`
	Blank        bool   `json:",omitempty"`
	Column       string `json:",omitempty"`
	Index        bool   `json:",omitempty"`
	Default      bool   `json:",omitempty"`
	DefaultFalse bool   `json:",omitempty"`
}

func (f BooleanField) FromJSON(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}

func (f BooleanField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f BooleanField) IsPk() bool {
	return false
}

func (f BooleanField) HasIndex() bool {
	return f.Index
}

func (f BooleanField) DefaultVal() (bool, Value) {
	if f.Default {
		return true, true
	} else if f.DefaultFalse {
		return true, false
	} else {
		return false, nil
	}
}

func (f BooleanField) Recipient() interface{} {
	if f.Null {
		var val sql.NullBool
		return &val
	}
	var val bool
	return &val
}

func (f BooleanField) CreateSQL() string {
	query := "bool"
	if f.Null {
		query += " NULL"
	} else {
		query += " NOT NULL"
	}
	if f.Default {
		query += " DEFAULT true"
	} else if f.DefaultFalse {
		query += " DEFAULT false"
	}
	return query
}

type IntChoice struct {
	Value int
	Label string
}

type IntegerField struct {
	Null        bool        `json:",omitempty"`
	Blank       bool        `json:",omitempty"`
	Choices     []IntChoice `json:",omitempty"`
	Column      string      `json:",omitempty"`
	Index       bool        `json:",omitempty"`
	Default     int         `json:",omitempty"`
	DefaultZero bool        `json:",omitempty"`
	PrimaryKey  bool        `json:",omitempty"`
	Unique      bool        `json:",omitempty"`
}

func (f IntegerField) FromJSON(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}

func (f IntegerField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f IntegerField) IsPk() bool {
	return f.PrimaryKey
}

func (f IntegerField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f IntegerField) DefaultVal() (bool, Value) {
	if f.Default != 0 || f.DefaultZero {
		return true, f.Default
	} else {
		return false, nil
	}
}

func (f IntegerField) Recipient() interface{} {
	if f.Null {
		var val sql.NullInt64
		return &val
	}
	var val int64
	return &val
}

func (f IntegerField) CreateSQL() string {
	query := "integer"
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	if f.Default != 0 || f.DefaultZero {
		query += fmt.Sprintf(" DEFAULT %d", f.Default)
	}
	return query
}

type AutoField IntegerField

func (f AutoField) FromJSON(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}

func (f AutoField) DBColumn(name string) string {
	if f.Column != "" {
		return f.Column
	}
	return name
}

func (f AutoField) IsPk() bool {
	return f.PrimaryKey
}

func (f AutoField) HasIndex() bool {
	return f.Index && !(f.PrimaryKey || f.Unique)
}

func (f AutoField) DefaultVal() (bool, Value) {
	return false, nil
}

func (f AutoField) Recipient() interface{} {
	var val int64
	return &val
}

func (f AutoField) CreateSQL() string {
	query := "integer"
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	query += " AUTOINCREMENT"
	return query
}
