package qb

import (
	"reflect"
	"testing"
	"time"
)

type testInterface interface {
	qbTable()
}

type testTable struct {
	ID    string `dbtable:"users" db:"id" table:"foo" col:"foo_id"`
	Name  string `db:"name" col:"foo_name"`
	Email string `db:"email" col:"foo_email"`
}

func (t *testTable) qbTable() {}

type testTableNoName struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

type testModel struct {
	ID string `db:"id"`
	TestModelWithTime
}

type TestModelWithTime struct {
	CreatedAt time.Time `db:"created_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type testModelType struct {
	testModel `dbtable:"model"`
	Name      string `db:"name"`
	Email     string `db:"email"`
}

type testModelTypePtr struct {
	*string
	*testModel `dbtable:"model"`
	Name       string `db:"name"`
	Email      string `db:"email"`
}

func TestNew(t *testing.T) {
	testTableInterface := func() testInterface {
		return &testTable{}
	}

	s := "string"

	type args struct {
		i    interface{}
		opts []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *QueryBuilder
		wantErr bool
	}{
		{"ok", args{testTable{}, nil}, &QueryBuilder{
			Table:         "users",
			Columns:       []string{"id", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with interface", args{testTableInterface(), nil}, &QueryBuilder{
			Table:         "users",
			Columns:       []string{"id", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with no name", args{testTableNoName{}, nil}, &QueryBuilder{
			Table:         "test_table_no_name",
			Columns:       []string{"id", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with model", args{testModelType{}, nil}, &QueryBuilder{
			Table:         "model",
			Columns:       []string{"id", "created_at", "deleted_at", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with model ptr", args{testModelTypePtr{string: &s}, nil}, &QueryBuilder{
			Table:         "model",
			Columns:       []string{"id", "created_at", "deleted_at", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with table name", args{&testTable{}, []Option{WithTableName("mytable")}}, &QueryBuilder{
			Table:         "mytable",
			Columns:       []string{"id", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with options", args{testTable{}, []Option{WithTableTag("table"), WithColumnTag("col")}}, &QueryBuilder{
			Table:         "foo",
			Columns:       []string{"foo_id", "foo_name", "foo_email"},
			SelectDeleted: false,
		}, false},
		{"fail", args{"not a struct", nil}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.i, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMust(t *testing.T) {
	type args struct {
		i    interface{}
		opts []Option
	}
	tests := []struct {
		name      string
		args      args
		want      *QueryBuilder
		wantPanic bool
	}{
		{"ok", args{testTable{}, nil}, &QueryBuilder{
			Table:         "users",
			Columns:       []string{"id", "name", "email"},
			SelectDeleted: false,
		}, false},
		{"ok with options", args{testTable{}, []Option{WithTableName("foo"), WithTableTag("table"), WithColumnTag("col")}}, &QueryBuilder{
			Table:         "foo",
			Columns:       []string{"foo_id", "foo_name", "foo_email"},
			SelectDeleted: false,
		}, false},
		{"fail", args{"not a struct", nil}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r != nil != tt.wantPanic {
					t.Errorf("Must() panic = %v, wantErr %v", r, tt.wantPanic)
				}
			}()

			got := Must(tt.args.i, tt.args.opts...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewQueryBuilder(t *testing.T) {
	type args struct {
		table   string
		columns []string
	}
	tests := []struct {
		name string
		args args
		want *QueryBuilder
	}{
		{"ok", args{"users", []string{"id", "name", "email"}}, &QueryBuilder{Table: "users", Columns: []string{"id", "name", "email"}, SelectDeleted: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryBuilder(tt.args.table, tt.args.columns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_Queries(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  string
		want2  string
		want3  string
	}{
		{"selectDeleted", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, true},
			"SELECT id, name, email, created_at, deleted_at FROM users WHERE id = $1",
			"INSERT INTO users (id, name, email, created_at, deleted_at) VALUES ($1, $2, $3, $4, $5)",
			"UPDATE users SET name = $1, email = $2, deleted_at = $3 WHERE id = $4",
			"UPDATE users SET deleted_at = $1 WHERE id = $2"},
		{"noSelectDeleted", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false},
			"SELECT id, name, email, created_at, deleted_at FROM users WHERE id = $1 AND deleted_at IS NULL",
			"INSERT INTO users (id, name, email, created_at, deleted_at) VALUES ($1, $2, $3, $4, $5)",
			"UPDATE users SET name = $1, email = $2, deleted_at = $3 WHERE id = $4",
			"UPDATE users SET deleted_at = $1 WHERE id = $2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			got, got1, got2, got3 := q.Queries()
			if got != tt.want {
				t.Errorf("QueryBuilder.Queries() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("QueryBuilder.Queries() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("QueryBuilder.Queries() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("QueryBuilder.Queries() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}

func TestQueryBuilder_SelectBy(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	type args struct {
		name       string
		extraNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"selectDeleted", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, true}, args{"email", nil}, "SELECT id, name, email, created_at, deleted_at FROM users WHERE email = $1"},
		{"noSelectDeleted", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, args{"email", nil}, "SELECT id, name, email, created_at, deleted_at FROM users WHERE email = $1 AND deleted_at IS NULL"},
		{"extra names", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, args{"name", []string{"email"}}, "SELECT id, name, email, created_at, deleted_at FROM users WHERE name = $1 AND email = $2 AND deleted_at IS NULL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.SelectBy(tt.args.name, tt.args.extraNames...); got != tt.want {
				t.Errorf("QueryBuilder.SelectBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_SelectAll(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"all", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, true}, "SELECT id, name, email, created_at, deleted_at FROM users"},
		{"non deleted", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, "SELECT id, name, email, created_at, deleted_at FROM users WHERE deleted_at IS NULL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.SelectAll(); got != tt.want {
				t.Errorf("QueryBuilder.SelectAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_InsertWithReturning(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"ok", fields{"users", []string{"id", "name", "email"}, false}, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"},
		{"ok no id", fields{"users", []string{"name", "email"}, false}, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.InsertWithReturning(); got != tt.want {
				t.Errorf("QueryBuilder.InsertWithReturning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_NamedInsert(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"ok", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, "INSERT INTO users (id, name, email, created_at, deleted_at) VALUES (:id, :name, :email, :created_at, :deleted_at)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.NamedInsert(); got != tt.want {
				t.Errorf("QueryBuilder.NamedInsert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_NamedInsertWithReturning(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"ok", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, "INSERT INTO users (name, email, created_at, deleted_at) VALUES (:name, :email, :created_at, :deleted_at) RETURNING id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.NamedInsertWithReturning(); got != tt.want {
				t.Errorf("QueryBuilder.NamedInsertWithReturning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryBuilder_NamedUpdate(t *testing.T) {
	type fields struct {
		Table         string
		Columns       []string
		SelectDeleted bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"ok", fields{"users", []string{"id", "name", "email", "created_at", "deleted_at"}, false}, "UPDATE users SET name = :name, email = :email, deleted_at = :deleted_at WHERE id = :id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QueryBuilder{
				Table:         tt.fields.Table,
				Columns:       tt.fields.Columns,
				SelectDeleted: tt.fields.SelectDeleted,
			}
			if got := q.NamedUpdate(); got != tt.want {
				t.Errorf("QueryBuilder.NamedUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
