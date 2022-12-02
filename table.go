package qb

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type table struct {
	Name    string
	Columns []string
}

func getTagValue(key string, f reflect.StructField) string {
	s := f.Tag.Get(key)
	if s == "" || s == "-" {
		return ""
	}
	parts := strings.SplitN(s, ",", 2)
	return strings.TrimSpace(parts[0])
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

func fieldColumns(f reflect.StructField, o *options) []string {
	var typ reflect.Type
	switch f.Type.Kind() {
	case reflect.Struct:
		typ = f.Type
	case reflect.Ptr:
		typ = f.Type.Elem()
		if typ.Kind() != reflect.Struct {
			return nil
		}
	default:
		return nil
	}

	var columns []string
	for i, n := 0, typ.NumField(); i < n; i++ {
		field := typ.Field(i)

		cols := fieldColumns(field, o)
		columns = append(columns, cols...)

		// Get the columns
		if name := getTagValue(o.columnTag, field); name != "" {
			columns = append(columns, name)
		}
	}
	return columns
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
		cols := fieldColumns(field, o)
		t.Columns = append(t.Columns, cols...)

		// Get the columns
		if name := getTagValue(o.columnTag, field); name != "" {
			t.Columns = append(t.Columns, name)
		}
	}

	if t.Name == "" {
		t.Name = getTableName(typ.Name())
	}

	return t, nil
}
