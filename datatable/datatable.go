// Package datatable contains the implementation of a data table tries to solve
// shortcomings of gherkin.DataTable.
package datatable

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jinzhu/copier"
	"github.com/tidwall/pretty"
)

// Options defines field options for the DataTable.
type Options struct {
	OptionalFields []string
	RequiredFields []string
}

// DataTable defines a table with fields names and rows.
type DataTable struct {
	fields []string
	rows   [][]string

	options *Options
}

// New creates a new DataTable with given fields. It optionally accepts initial
// rows.
func New(fields []string, rows ...[]string) (*DataTable, error) {
	return NewWithOptions(nil, fields, rows...)
}

// NewWithOptions create a new DataTable with options and given fields. It
// optionally accepts inital rows.
func NewWithOptions(options *Options, fields []string, rows ...[]string) (*DataTable, error) {
	for _, row := range rows {
		if len(row) != len(fields) {
			return nil, fmt.Errorf("expected row length of %d, got %d", len(fields), len(row))
		}
	}

	dt := &DataTable{
		fields:  fields,
		rows:    rows,
		options: options,
	}

	if err := dt.validateFields(); err != nil {
		return nil, err
	}

	return dt, nil
}

// FromGherkin creates a new DataTable from *gherkin.DataTable
func FromGherkin(dt *gherkin.DataTable) (*DataTable, error) {
	return FromGherkinWithOptions(nil, dt)
}

// FromGherkinWithOptions creates a new DataTable from *gherkin.DataTable with options.
func FromGherkinWithOptions(options *Options, dt *gherkin.DataTable) (*DataTable, error) {
	if len(dt.Rows) < 2 {
		return nil, errors.New("data table must have at least two rows")
	}

	return NewWithOptions(options, values(dt.Rows[0]), rowValues(dt.Rows[1:])...)
}

// validateFields ensures that required fields are present and there are only
// fields listed in the AllowedFields option if the data table was created with
// options.
func (t *DataTable) validateFields() error {
	if t.options == nil {
		return nil
	}

	for _, field := range t.options.RequiredFields {
		if !contains(t.fields, field) {
			return fmt.Errorf(`data table is missing required field %q`, field)
		}
	}

	if len(t.options.OptionalFields) == 0 {
		return nil
	}

	allowedFields := append(t.options.OptionalFields, t.options.RequiredFields...)

	for _, field := range t.fields {
		if !contains(allowedFields, field) {
			return fmt.Errorf(
				`data table contains additional field %q, allowed fields are "%s"`,
				field,
				strings.Join(allowedFields, `", "`),
			)
		}
	}

	return nil
}

// Copy makes a copy of the data table.
func (t *DataTable) Copy() *DataTable {
	c := &DataTable{
		fields: make([]string, len(t.fields)),
		rows:   make([][]string, len(t.rows)),
	}

	copier.Copy(&c.fields, &t.fields)
	copier.Copy(&c.rows, &t.rows)

	return c
}

// FindRow compares given row with all rows in the data table and returns the
// row index if a matching row is found. Returns -1 if row cannot be found.
func (t *DataTable) FindRow(row []string) int {
	for i, r := range t.rows {
		if matchValues(r, row) {
			return i
		}
	}

	return -1
}

// RemoveRow removes the row at given index.
func (t *DataTable) RemoveRow(index int) {
	t.rows = append(t.rows[:index], t.rows[index+1:]...)
}

// AppendRow appends a row to the data table. Will return an error if the
// number of fields does not match the data table's fields.
func (t *DataTable) AppendRow(row []string) error {
	if len(row) != len(t.fields) {
		return fmt.Errorf("expected row length of %d, got %d", len(t.fields), len(row))
	}

	t.rows = append(t.rows, row)

	return nil
}

// Len returns the row count of the data table.
func (t *DataTable) Len() int {
	return len(t.rows)
}

// Fields returns the table fields.
func (t *DataTable) Fields() []string {
	return t.fields
}

// Rows transforms the data table rows into a slice of maps and returns it.
// The map keys are the data table's fields for every row.
func (t *DataTable) Rows() []map[string]string {
	s := make([]map[string]string, len(t.rows))

	for i, row := range t.rows {
		m := make(map[string]string)
		for j, field := range t.fields {
			m[field] = row[j]
		}

		s[i] = m
	}

	return s
}

// RowValues returns the row values.
func (t *DataTable) RowValues() [][]string {
	return t.rows
}

// PrettyJSON is a convenience function for transforming the data table into
// its prettyprinted json representation. Will panic if json marshalling fails.
func (t *DataTable) PrettyJSON() []byte {
	buf, err := json.Marshal(t.Rows())
	if err != nil {
		panic(err)
	}

	return pretty.Pretty(buf)
}

// rowValues converts a slice of *gherkin.TableRow into a slice of string
// slices.
func rowValues(rows []*gherkin.TableRow) [][]string {
	vals := make([][]string, len(rows))
	for i, row := range rows {
		vals[i] = values(row)
	}

	return vals
}

// values converts a *gherkin.TableRow into a slice of strings.
func values(row *gherkin.TableRow) []string {
	values := make([]string, len(row.Cells))
	for i, cell := range row.Cells {
		values[i] = cell.Value
	}

	return values
}

// matchRow returns true if all values in two string slices match pairwise.
func matchValues(a, b []string) bool {
	for i := range a {
		if b[i] != a[i] {
			return false
		}
	}

	return true
}

// contains returns true if haystack contains needle
func contains(haystack []string, needle string) bool {
	for _, element := range haystack {
		if element == needle {
			return true
		}
	}

	return false
}
