package gomodels

import (
	"encoding/json"
	"testing"
)

// TestFields tests Fields marshal/unmarshall methods
func TestFields(t *testing.T) {

	t.Run("UnmarshalInvalidJSON", func(t *testing.T) {
		fields := Fields{}
		err := fields.UnmarshalJSON([]byte("-"))
		if _, ok := err.(*json.SyntaxError); !ok {
			t.Errorf("expected json.SyntaxError, got %T", err)
		}
	})

	t.Run("UnmarshalUnknownField", func(t *testing.T) {
		fields := Fields{}
		data := []byte(`{"foo": {"InvalidField": {}}}`)
		err := fields.UnmarshalJSON(data)
		if err == nil || err.Error() != "invalid field type: InvalidField" {
			t.Errorf("expected invalid field error, got %s", err)
		}
	})

	t.Run("UnmarshalInvalidField", func(t *testing.T) {
		fields := Fields{}
		data := []byte(`{"foo": {"IntegerField": {"PrimaryKey": []}}}`)
		err := fields.UnmarshalJSON(data)
		if _, ok := err.(*json.UnmarshalTypeError); !ok {
			t.Errorf("expected json.UnmarshalTypeError, got %T", err)
		}
	})

	t.Run("UnmarshalValidField", func(t *testing.T) {
		fields := Fields{}
		data := []byte(`{"foo": {"CharField": {"MaxLength": 100}}}`)
		err := fields.UnmarshalJSON(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(fields) != 1 {
			t.Fatal("expected map to contain one field")
		}
		if _, ok := fields["foo"]; !ok {
			t.Fatal("expected map to contain foo field")
		}
		if _, ok := fields["foo"].(CharField); !ok {
			t.Fatalf("expected CharField, got %T", fields["foo"])
		}
		field := fields["foo"].(CharField)
		if field.MaxLength != 100 {
			t.Errorf("expected MaxLength to be 100, got %d", field.MaxLength)
		}
	})

	t.Run("Marshal", func(t *testing.T) {
		fields := Fields{"foo": BooleanField{Default: true}}
		data, err := fields.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		expected := `{"foo":{"BooleanField":{"Default":true}}}`
		if string(data) != expected {
			t.Fatalf("expected %s, got %s", expected, string(data))
		}
	})
}
