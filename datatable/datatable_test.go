package datatable

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/godog/gherkin"
)

func TestFromGherkin(t *testing.T) {
	fields, rows := testData()
	table := append([][]string{fields}, rows...)

	dt, err := FromGherkin(buildTable(table))
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	if !reflect.DeepEqual(dt.Fields(), fields) {
		t.Fatalf("expected fields %#v, got %#v", fields, dt.Fields())
	}

	if !reflect.DeepEqual(dt.RowValues(), rows) {
		t.Fatalf("expected fields %#v, got %#v", rows, dt.RowValues())
	}
}

func TestFromMalformedGherkin(t *testing.T) {
	_, err := FromGherkin(buildTable([][]string{{"foo"}}))
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestNewWithOptions(t *testing.T) {
	cases := []struct {
		name        string
		options     *Options
		fields      []string
		rows        [][]string
		expectError bool
	}{
		{
			name:   "no options, correct row length",
			fields: []string{"name", "value"},
			rows:   [][]string{{"foo", "bar"}},
		},
		{
			name:        "no options, incorrect row length",
			fields:      []string{"name", "value"},
			rows:        [][]string{{"foo"}},
			expectError: true,
		},
		{
			name:    "required fields present",
			options: &Options{RequiredFields: []string{"value", "name"}},
			fields:  []string{"name", "value"},
			rows:    [][]string{{"foo", "bar"}},
		},
		{
			name:    "required fields and additional fields present",
			options: &Options{RequiredFields: []string{"value", "name"}},
			fields:  []string{"name", "value", "additonal"},
			rows:    [][]string{{"foo", "bar", "baz"}},
		},
		{
			name:        "unknown field",
			options:     &Options{OptionalFields: []string{"value", "name"}},
			fields:      []string{"name", "unknown"},
			rows:        [][]string{{"foo", "baz"}},
			expectError: true,
		},
		{
			name: "required and optional fields",
			options: &Options{
				RequiredFields: []string{"name", "value"},
				OptionalFields: []string{"tag"},
			},
			fields:      []string{"name", "value", "tag"},
			rows:        [][]string{{"foo", "bar", "baz"}},
			expectError: false,
		},
		{
			name: "required and optional fields, unknown additional field",
			options: &Options{
				RequiredFields: []string{"name", "value"},
				OptionalFields: []string{"tag"},
			},
			fields:      []string{"name", "value", "tag", "unknown"},
			rows:        [][]string{{"foo", "bar", "baz", "qux"}},
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewWithOptions(tc.options, tc.fields, tc.rows...)
			if tc.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}
		})
	}
}

func TestCopy(t *testing.T) {
	fields, rows := testData()

	dt, err := New(fields, rows...)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	ct := dt.Copy()

	if ct == dt {
		t.Fatal("copy points to source")
	}

	if !reflect.DeepEqual(dt, ct) {
		t.Fatalf("copy and source do not contain the same data")
	}
}

func TestRowOperations(t *testing.T) {
	fields, rows := testData()

	dt, err := New(fields, rows...)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	index := dt.FindRow([]string{"4", "5", "6"})
	if index != 1 {
		t.Fatalf("expected index 1, got %d", index)
	}

	dt.RemoveRow(index)

	index = dt.FindRow([]string{"4", "5", "6"})
	if index != -1 {
		t.Fatalf("expected index -1, got %d", index)
	}

	err = dt.AppendRow([]string{"10", "11", "12"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	index = dt.FindRow([]string{"10", "11", "12"})
	if index != 2 {
		t.Fatalf("expected index 2, got %d", index)
	}
}

func TestRows(t *testing.T) {
	fields, rows := testData()

	dt, err := New(fields, rows...)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	s := dt.Rows()

	expected := []map[string]string{
		{"one": "1", "two": "2", "three": "3"},
		{"one": "4", "two": "5", "three": "6"},
		{"one": "7", "two": "8", "three": "9"},
	}

	if !reflect.DeepEqual(s, expected) {
		t.Fatalf("expected %#v, got %#v", expected, s)
	}
}

func testData() ([]string, [][]string) {
	fields := []string{"one", "two", "three"}
	rows := [][]string{
		{"1", "2", "3"},
		{"4", "5", "6"},
		{"7", "8", "9"},
	}

	return fields, rows
}

func buildTable(src [][]string) *gherkin.DataTable {
	rows := make([]*gherkin.TableRow, len(src))
	for i, row := range src {
		cells := make([]*gherkin.TableCell, len(row))
		for j, value := range row {
			cells[j] = &gherkin.TableCell{Value: value}
		}

		rows[i] = &gherkin.TableRow{Cells: cells}
	}

	return &gherkin.DataTable{Rows: rows}
}
