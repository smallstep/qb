package qb

import (
	"reflect"
	"testing"
)

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
