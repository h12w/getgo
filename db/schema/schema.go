// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

/*
	Field, Record classes are used to describe a struct value binded to a schema.
*/

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type Field struct {
	Name  string
	Value interface{}
	IsKey bool
}

func NewField(name string, value interface{}, isKey bool) *Field {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			value = nil
		} else {
			value = v.Elem().Interface()
		}
	}
	return &Field{
		Name:  camelToSnake(name),
		Value: value,
		IsKey: isKey}

}

type FieldSelector func(*Field) bool

func Key(f *Field) bool {
	return f.IsKey
}

func NonKey(f *Field) bool {
	return !f.IsKey
}

func DbType(f *Field) bool {
	switch f.Value.(type) {
	case int, *int, string, *string, bool, *bool,
		int8, int16, int32, int64,
		*int8, *int16, *int32, *int64,
		uint8, uint16, uint32, uint64,
		*uint8, *uint16, *uint32, *uint64,
		float32, float64,
		*float32, *float64,
		time.Time, *time.Time:
		return true
	}
	return false
}

func NonNil(f *Field) bool {
	return f.Value != nil
}

type Fields []*Field

func (fs Fields) Filter(conds ...FieldSelector) (fields Fields) {
nextField:
	for _, field := range fs {
		for _, cond := range conds {
			if !cond(field) {
				continue nextField
			}
		}
		fields = append(fields, field)
	}
	return
}

func (fs Fields) Values() []interface{} {
	values := make([]interface{}, len(fs))
	for i, f := range fs {
		values[i] = f.Value
	}
	return values
}

type Record struct {
	Name   string
	Fields Fields
}

type SqlTag struct {
	Pk bool
}

func parseSqlTag(tag string) *SqlTag {
	sqlTag := &SqlTag{}
	specs := strings.Split(tag, ",")
	for _, spec := range specs {
		if spec == "pk" {
			sqlTag.Pk = true
		}
	}
	return sqlTag
}

func NewRecord(s interface{}) *Record {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf(
			"record must be initialized from a struct. %v, %v.", v.Kind(), v))
	}

	t := v.Type()
	name := camelToSnake(t.Name())

	// extract key value pairs
	fields := make(Fields, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sqlTag := parseSqlTag(t.Field(i).Tag.Get("sql"))
		fields[i] = NewField(t.Field(i).Name, v.Field(i).Interface(), sqlTag.Pk)
	}

	if len(fields) > 0 &&
		fields[0].Name == "id" {
		fields[0].IsKey = true
	}

	return &Record{
		Name:   name,
		Fields: fields,
	}
}

func (r *Record) SetKey(names ...string) *Record {
	for _, name := range names {
		name = camelToSnake(name)
		for i := range r.Fields {
			if r.Fields[i].Name == name {
				r.Fields[i].IsKey = true
			}
		}
	}
	return r
}

var lowerUpper = regexp.MustCompile(`([\P{Lu}])([\p{Lu}])`)

// convert name from camel case to snake case
func camelToSnake(s string) string {
	return strings.ToLower(lowerUpper.ReplaceAllString(s, `${1}_${2}`))
}
