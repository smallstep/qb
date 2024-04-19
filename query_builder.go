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

// BindParam represents the binding parameter in SQL queries.
type BindParam int

const (
	// Dollar is the binding parameter type used in PostgreSQL, these parameters
	// use the character $ and number with the positional number starting in 1.
	// They look like $1, $2, ...
	DOLLAR BindParam = iota + 1
	// QUESTION is the binding parameter type used in mysql and sqlite3, this
	// parameters just the character ?.
	QUESTION
)

// QueryBuilder provides a simple list of SQL queries that can be used by the
// models. It requires tables with the columns id, created_at, and deleted_at.
type QueryBuilder struct {
	Table         string
	Columns       []string
	SelectDeleted bool
	PrimaryKey    string
	BindType      BindParam
}

type options struct {
	tableName string
	tableTag  string
	columnTag string
	bindType  BindParam
}

func defaultOptions() *options {
	return &options{
		tableTag:  "dbtable",
		columnTag: "db",
		bindType:  DOLLAR,
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
func ColumnTag(key string) Option {
	return func(o *options) {
		if key != "" {
			o.columnTag = key
		}
	}
}

// BindType defines the binding parameter type used. It defaults to DOLLAR.
func BindType(t BindParam) Option {
	return func(o *options) {
		if t != 0 {
			o.bindType = t
		}
	}
}

// WithColumnTag sets the tag key used to get a column name. It defaults to
// "db".
//
// Deprecated: use ColumnTag.
func WithColumnTag(key string) Option {
	return ColumnTag(key)
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
	qb := NewQueryBuilder(t.Name, t.Columns)
	if t.PrimaryKey != "" {
		qb.PrimaryKey = t.PrimaryKey
	}
	if o.bindType != 0 {
		qb.BindType = o.bindType
	}
	return qb, nil
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
		PrimaryKey:    idColumn,
		BindType:      DOLLAR,
	}
}

// Queries returns the queries for select by id, insert,
// update, and delete.
func (q *QueryBuilder) Queries() (string, string, string, string) {
	return q.Select(), q.Insert(), q.Update(), q.Delete()
}

// Select returns the query to get a record by id.
func (q *QueryBuilder) Select() string {
	s := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s", q.columns(), q.Table, q.idColumn(), q.bind(1))
	if !q.SelectDeleted {
		s += " AND deleted_at IS NULL"
	}
	return s
}

// SelectBy returns a query to get a record by the given column name.
func (q *QueryBuilder) SelectBy(name string, extraNames ...string) string {
	s := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s", q.columns(), q.Table, name, q.bind(1))
	// Append extra names.
	for i, n := range extraNames {
		s += fmt.Sprintf(" AND %s = %s", n, q.bind(i+2))
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
	var idName = q.idColumn()
	var columns, values []string
	for _, name := range q.Columns {
		if name != idName {
			columns = append(columns, name)
			values = append(values, q.bind(pos))
			pos++
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s", q.Table, join(columns), join(values), idName)
}

// Insert returns the query to insert a record using named values.
func (q *QueryBuilder) NamedInsert() string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", q.Table, q.columns(), q.namedValues())
}

// NamedInsertWithReturning returns the query to insert a record using named
// values, the query will return the id.
func (q *QueryBuilder) NamedInsertWithReturning() string {
	var idName = q.idColumn()
	var columns, values []string
	for _, name := range q.Columns {
		if name != idName {
			columns = append(columns, name)
			values = append(values, ":"+name)
		}
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING %s", q.Table, join(columns), join(values), idName)
}

// Update returns the query to update a record. Update won't update neither the
// id nor the created_at column.
func (q *QueryBuilder) Update() string {
	var v []string
	var idName = q.idColumn()
	pos := 1
	for _, name := range q.Columns {
		if name != idName && name != createdAtColumn {
			v = append(v, name+" = "+q.bind(pos))
			pos++
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", q.Table, join(v), q.idColumn(), q.bind(pos))
}

// NamedUpdate returns the query to update a record using named values. Update
// won't update neither the id nor the created_at column.
func (q *QueryBuilder) NamedUpdate() string {
	var values []string
	var idName = q.idColumn()
	for _, name := range q.Columns {
		if name != idName && name != createdAtColumn {
			values = append(values, name+" = :"+name)
		}
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = :%s", q.Table, join(values), q.idColumn(), idName)
}

// Delete returns the query to mark a record as deleted.
func (q *QueryBuilder) Delete() string {
	return fmt.Sprintf("UPDATE %s SET deleted_at = %s WHERE %s = %s", q.Table, q.bind(1), q.idColumn(), q.bind(2))
}

// HardDelete returns the query to delete a row by id.
func (q *QueryBuilder) HardDelete() string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = %s", q.Table, q.idColumn(), q.bind(1))
}

func (q *QueryBuilder) idColumn() string {
	if q.PrimaryKey != "" {
		return q.PrimaryKey
	}
	return idColumn
}

func (q *QueryBuilder) bind(i int) string {
	switch q.BindType {
	case QUESTION:
		return "?"
	default:
		return "$" + strconv.Itoa(i)
	}
}

func (q *QueryBuilder) columns() string {
	return strings.Join(q.Columns, ", ")
}

func (q *QueryBuilder) values() string {
	n := len(q.Columns)
	c := make([]string, n)
	for i := 0; i < n; i++ {
		c[i] = q.bind(i + 1)
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
