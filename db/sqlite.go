package db

import (
	"fmt"
	"strings"

	"github.com/alicebob/sqlittle"
)

type columIndex struct {
	col      *sqlittle.TableColumn
	rowIndex int
	rowid    bool
}

// Regroups a Record to a Row, filling in missing columns as needed.
func toRow(rowid int64, cis []columIndex, r sqlittle.Record) Row {
	row := make(Row, len(cis))
	for i, c := range cis {
		if c.rowid {
			row[i] = rowid
			continue
		}
		if len(r) <= c.rowIndex {
			// use 'DEFAULT' when the record is too short
			row[i] = c.col.Default
		} else {
			row[i] = r[c.rowIndex]
		}
	}
	return row
}

// given column names returns the index in a Row this column is expected, and
// the column definition. Allows 'rowid' alias.
func toColumnIndexRowid(s *sqlittle.Schema, columns []string) ([]columIndex, error) {
	res := make([]columIndex, 0, len(columns))
	for _, c := range columns {
		n := s.Column(c)
		if n < 0 {
			cup := strings.ToUpper(c)
			if cup == "ROWID" || cup == "OID" || cup == "_ROWID_" {
				res = append(res, columIndex{nil, n, true})
				continue
			} else {
				return nil, fmt.Errorf("no such column: %q", c)
			}
		}
		res = append(res, columIndex{&s.Columns[n], n, false})
	}
	return res, nil
}

// given column names returns the index of this column in a row in the index (and
// the column definition). For database order of the columns depends on the
// primary key.
func toColumnIndexNonRowid(s *sqlittle.Schema, columns []string) ([]columIndex, error) {
	stored := columnStoreOrder(s) // column indexes in disk order
	res := make([]columIndex, 0, len(columns))
	for _, c := range columns {
		n := s.Column(c)
		if n < 0 {
			return nil, fmt.Errorf("no such column: %q", c)
		}
		res = append(res, columIndex{&s.Columns[n], stored[n], false})
	}
	return res, nil
}

// for non-rowid tables only:
// given an index gives back the indexes in a row which form the primary key.
func pkColumns(schema *sqlittle.Schema, ind *sqlittle.SchemaIndex) []int {
	if !schema.WithoutRowid {
		panic("can't call pkColumns on an rowid table")
	}

	var res []int
	for _, c := range schema.PK {
		if in := ind.Column(c.Column); in < 0 {
			ind.Columns = append(ind.Columns, c)
			res = append(res, len(ind.Columns)-1)
		} else {
			res = append(res, in)
		}
	}
	return res
}

// given a non-rowid table, gives the order columns are stored on disk
func columnStoreOrder(schema *sqlittle.Schema) []int {
	if !schema.WithoutRowid {
		panic("can't call columnStoreOrder on an rowid table")
	}

	// all PK columns come first, then all other columns, in order
	var cols = make([]string, 0, len(schema.Columns))
	for _, c := range schema.PK {
		cols = append(cols, strings.ToLower(c.Column))
	}
loop:
	for _, c := range schema.Columns {
		n := strings.ToLower(c.Column)
		for _, oc := range cols {
			if oc == n {
				continue loop
			}
		}
		cols = append(cols, n)
	}

	res := make([]int, len(cols))
loop2:
	for i, c := range schema.Columns {
		n := strings.ToLower(c.Column)
		for j, oc := range cols {
			if oc == n {
				res[i] = j
				continue loop2
			}
		}
	}
	return res
}
