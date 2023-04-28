package qb

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	idColumn        = "id"
	createdAtColumn = "created_at"
	deletedAtColumn = "deleted_at"
)

// QueryBuilder provides a simple list of SQL queries that can be used by the
// models. It requires tables with the columns id, created_at, and deleted_at.
type QueryBuilder struct {
	Table         string
	Columns       []string
	SelectDeleted bool
}

type options struct {
	tableName string
	tableTag  string
	columnTag string
}

func defaultOptions() *options {
	return &options{
		tableTag:  "dbtable",
		columnTag: "db",
	}
}

// Option is the type used to pass options to the New and Must functions.
type Option func(o *options)

// TableName sets the table name to use.
func TableName(name string) Option {
	return func(o *options) {
		if name != "" {
			o.tableName = name
		}
	}
}

// TableTag sets the tag key used to get the table name. It defaults to
// "dbtable".
func TableTag(key string) Option {
	return func(o *options) {
		if key != "" {
			o.tableTag = key
		}
	}
}

// ColumnTag sets the tag key used to get a column name. It defaults to
// "db".
func WithColumnTag(key string) Option {
	return func(o *options) {
		if key != "" {
			o.columnTag = key
		}
	}
}

// New returns a new query builder configured with the fields tags in the given
// struct. By default it uses the tag "dbtable" for the table name and "db" for
// the column names.
func New(i any, opts ...Option) (*QueryBuilder, error) {
	o := defaultOptions()
	for _, fn := range opts {
		fn(o)
	}
	t, err := getTable(i, o)
	if err != nil {
		return nil, err
	}
	return NewQueryBuilder(t.Name, t.Columns), nil
}

// Must returns a new query builder configured with the fields tags in the given
// struct. By default it uses the tag "dbtable" for the table name and "db" for
// the column names.
//
// Must will panic if i is not an struct.
func Must(i any, opts ...Option) *QueryBuilder {
	qb, err := New(i, opts...)
	if err != nil {
		panic(err)
	}
	return qb
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
	if !q.SelectDeleted {
		return fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL", q.columns(), q.Table)
	}
	return fmt.Sprintf("SELECT %s FROM %s", q.columns(), q.Table)
}

// Insert returns the query to insert a record.
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

// Insert returns the query to insert a record using named values.
func (q *QueryBuilder) NamedInsert() string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.Table, q.columns(), q.namedValues())
}

// NamedInsertWithReturning returns the query to insert a record using named
// values, the query will return the id.
func (q *QueryBuilder) NamedInsertWithReturning() string {
	var columns, values []string
	for _, name := range q.Columns {
		if name != idColumn {
			columns = append(columns, name)
			values = append(values, ":"+name)
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", q.Table, join(columns), join(values))
}

// Update returns the query to update a record. Update won't update neither the
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

// NamedUpdate returns the query to update a record using named values. Update
// won't update neither the id nor the created_at column.
func (q *QueryBuilder) NamedUpdate() string {
	var values []string
	for _, name := range q.Columns {
		if name != idColumn && name != createdAtColumn {
			values = append(values, name+" = :"+name)
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE id = :id", q.Table, join(values))
}

// Delete returns the query to mark a record as deleted.
func (q *QueryBuilder) Delete() string {
	return fmt.Sprintf("UPDATE %s SET deleted_at = $1 WHERE id = $2", q.Table)
}

// HardDelete returns the query to delete a row by id.
func (q *QueryBuilder) HardDelete() string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = $1", q.Table)
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

func (q *QueryBuilder) namedValues() string {
	n := len(q.Columns)
	c := make([]string, n)
	for i, s := range q.Columns {
		c[i] = ":" + s
	}
	return join(c)
}

func join(s []string) string {
	return strings.Join(s, ", ")
}
