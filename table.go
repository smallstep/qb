package qb

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type table struct {
	Name       string
	Columns    []string
	PrimaryKey string
}

func isPrimaryKey(s string) bool {
	return strings.EqualFold(s, "primaryKey") || strings.EqualFold(s, "pkey")
}

func (t *table) addColumn(name string) error {
	if parts := strings.SplitN(name, ",", 2); len(parts) == 2 && isPrimaryKey(parts[1]) {
		if t.PrimaryKey != "" && t.PrimaryKey != parts[0] {
			return errors.New("table cannot have more than one primary key")
		}
		name = strings.TrimSpace(parts[0])
		t.Columns = append(t.Columns, name)
		t.PrimaryKey = name
		return nil
	}

	t.Columns = append(t.Columns, strings.TrimSpace(name))
	return nil
}

func (t *table) addColumnsFromTable(rt table) error {
	if rt.PrimaryKey != "" {
		if t.PrimaryKey != "" && t.PrimaryKey != rt.PrimaryKey {
			return errors.New("table cannot have more than one primary key")
		}
		t.PrimaryKey = rt.PrimaryKey
	}
	t.Columns = append(t.Columns, rt.Columns...)
	return nil
}

func getTagValue(key string, f reflect.StructField) string {
	s := f.Tag.Get(key)
	if s == "-" {
		return ""
	}
	return s
}

func getTableName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) && i != 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func structOf(i any) (reflect.Value, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Struct:
		return v, nil
	case reflect.Ptr, reflect.Interface:
		elem := v.Elem()
		if elem.Kind() == reflect.Struct {
			return elem, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("%T is neither struct nor does it point to one", i)
}

func fieldColumns(f reflect.StructField, o *options) (table, error) {
	var typ reflect.Type
	switch f.Type.Kind() {
	case reflect.Struct:
		typ = f.Type
	case reflect.Ptr:
		typ = f.Type.Elem()
		if typ.Kind() != reflect.Struct {
			return table{}, nil
		}
	default:
		return table{}, nil
	}

	var t table
	for i, n := 0, typ.NumField(); i < n; i++ {
		field := typ.Field(i)

		// Get the columns in embedded structs
		rt, err := fieldColumns(field, o)
		if err != nil {
			return table{}, err
		}
		if err := t.addColumnsFromTable(rt); err != nil {
			return table{}, err
		}

		// Get the columns
		if name := getTagValue(o.columnTag, field); name != "" {
			if err := t.addColumn(name); err != nil {
				return table{}, err
			}
		}
	}
	return t, nil
}

func getTable(i any, o *options) (table, error) {
	v, err := structOf(i)
	if err != nil {
		return table{}, err
	}

	t := table{Name: o.tableName}
	typ := v.Type()
	for i, n := 0, typ.NumField(); i < n; i++ {
		field := typ.Field(i)

		// Get table if available
		if t.Name == "" {
			if name := getTagValue(o.tableTag, field); name != "" {
				t.Name = name
			}
		}

		// Resolve columns recursively
		rt, err := fieldColumns(field, o)
		if err != nil {
			return table{}, err
		}
		if err := t.addColumnsFromTable(rt); err != nil {
			return table{}, err
		}

		// Get the columns
		if name := getTagValue(o.columnTag, field); name != "" {
			if err := t.addColumn(name); err != nil {
				return table{}, err
			}
		}
	}

	if t.Name == "" {
		t.Name = getTableName(typ.Name())
	}

	return t, nil
}
