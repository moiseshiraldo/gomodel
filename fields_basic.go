package gomodels

import (
	"encoding/json"
	"fmt"
)

type CharChoice struct {
	Value string
	Label string
}

type CharField struct {
	Null       bool         `json:",omitempty"`
	Blank      bool         `json:",omitempty"`
	Choices    []CharChoice `json:",omitempty"`
	Column     string       `json:",omitempty"`
	Index      bool         `json:",omitempty"`
	Default    string       `json:",omitempty"`
	PrimaryKey bool         `json:",omitempty"`
	Unique     bool         `json:",omitempty"`
	MaxLength  int          `json:",omitempty"`
}

func (f CharField) IsPk() bool {
	return f.PrimaryKey
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

func (f CharField) CreateSQL() string {
	query := fmt.Sprintf("varchar(%d)", f.MaxLength)
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	return query
}

type BooleanField struct {
	Null    bool   `json:",omitempty"`
	Blank   bool   `json:",omitempty"`
	Column  string `json:",omitempty"`
	Index   bool   `json:",omitempty"`
	Default bool   `json:",omitempty"`
}

func (f BooleanField) IsPk() bool {
	return false
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

func (f BooleanField) CreateSQL() string {
	query := "bool"
	if f.Null {
		query += " NULL"
	} else {
		query += " NOT NULL"
	}
	return query
}

type IntChoice struct {
	Value int
	Label string
}

type IntegerField struct {
	Null       bool        `json:",omitempty"`
	Blank      bool        `json:",omitempty"`
	Choices    []IntChoice `json:",omitempty"`
	Column     string      `json:",omitempty"`
	Index      bool        `json:",omitempty"`
	Default    int         `json:",omitempty"`
	PrimaryKey bool        `json:",omitempty"`
	Unique     bool        `json:",omitempty"`
}

func (f IntegerField) IsPk() bool {
	return f.PrimaryKey
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

func (f IntegerField) CreateSQL() string {
	query := "integer"
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	return query
}

type AutoField IntegerField

func (f AutoField) IsPk() bool {
	return f.PrimaryKey
}

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

func (f AutoField) CreateSQL() string {
	query := "integer"
	query += sqlColumnOptions(f.Null, f.PrimaryKey, f.Unique)
	query += " AUTOINCREMENT"
	return query
}
