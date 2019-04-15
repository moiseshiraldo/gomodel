package gomodels

import "encoding/json"

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

func (f CharField) FromJson(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
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

func (f BooleanField) FromJson(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
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

func (f IntegerField) FromJson(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}

type AutoField IntegerField

func (f AutoField) IsPk() bool {
	return f.PrimaryKey
}

func (f AutoField) FromJson(raw []byte) (Field, error) {
	err := json.Unmarshal(raw, &f)
	return f, err
}
