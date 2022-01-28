package qb

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	idColumn        = "id"
	createdAtColumn = "created_at"
)

// QueryBuilder provides a simple list of SQL queries that can be used by the
// models. It requires tables with the columns id, created_at, and deleted_at.
type QueryBuilder struct {
	Table         string
	Columns       []string
	SelectDeleted bool
}

// NewQueryBuilder returns a new query builder configured with the given table
// and columns.
func NewQueryBuilder(table string, columns []string) *QueryBuilder {
	return &QueryBuilder{
		Table:         table,
		Columns:       columns,
		SelectDeleted: false,
	}
}

// Queries returns the queries for select by id, insert,
// update, and delete.
func (q *QueryBuilder) Queries() (string, string, string, string) {
	return q.Select(), q.Insert(), q.Update(), q.Delete()
}

// Select returns the query to get a record by id.
func (q *QueryBuilder) Select() string {
	s := fmt.Sprintf("SELECT %s FROM %s WHERE id = $1", q.columns(), q.Table)
	if !q.SelectDeleted {
		s += " AND deleted_at IS NULL"
	}
	return s
}

// SelectBy returns a query to get a record by the given column name.
func (q *QueryBuilder) SelectBy(name string, extraNames ...string) string {
	s := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", q.columns(), q.Table, name)
	// Append extra names.
	for i, n := range extraNames {
		s += fmt.Sprintf(" AND %s = $%d", n, i+2)
	}
	if !q.SelectDeleted {
		s += " AND deleted_at IS NULL"
	}
	return s
}

// SelectAll returns a query to get all entries in a table.
func (q *QueryBuilder) SelectAll() string {
	if q.SelectDeleted {
		return fmt.Sprintf("SELECT %s FROM %s", q.columns(), q.Table)
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL", q.columns(), q.Table)
}

// Insert returns the query to insert an record.
func (q *QueryBuilder) Insert() string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.Table, q.columns(), q.values())
}

// InsertWithReturning returns the query to insert that returns the id.
func (q *QueryBuilder) InsertWithReturning() string {
	var pos = 1
	var columns, values []string
	for _, name := range q.Columns {
		if name != idColumn {
			columns = append(columns, name)
			values = append(values, "$"+strconv.Itoa(pos))
			pos++
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", q.Table, join(columns), join(values))
}

// Update returns the query to update an record. Update won't update neither the
// id nor the created_at column.
func (q *QueryBuilder) Update() string {
	var v []string
	pos := 1
	for _, name := range q.Columns {
		if name != idColumn && name != createdAtColumn {
			v = append(v, name+" = $"+strconv.Itoa(pos))
			pos++
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", q.Table, join(v), pos)
}

// Delete returns the query to mark a record as deleted.
func (q *QueryBuilder) Delete() string {
	return fmt.Sprintf("UPDATE %s SET deleted_at = $1 WHERE id = $2", q.Table)
}

func (q *QueryBuilder) columns() string {
	return strings.Join(q.Columns, ", ")
}

func (q *QueryBuilder) values() string {
	n := len(q.Columns)
	c := make([]string, n)
	for i := 0; i < n; i++ {
		c[i] = "$" + strconv.Itoa(i+1)
	}
	return join(c)
}

func join(s []string) string {
	return strings.Join(s, ", ")
}
