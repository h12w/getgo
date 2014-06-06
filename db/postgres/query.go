// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"bytes"
	"database/sql"
	"fmt"
	. "github.com/hailiang/getgo/db/schema"
	"strings"
)

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Query struct {
	Cmd  string
	Args []interface{}
}

func (q *Query) Do(ex Execer) (sql.Result, error) {
	return ex.Exec(q.Cmd, q.Args...)
}

// Limitation: The current implementation does not handle confilicts between
// transactions. So there must be an external lock to make sure there is only
// one transaction at the same time.
func Upsert(tx Execer, s interface{}) error {
	var r *Record
	switch s.(type) {
	case *Record:
		r = s.(*Record)
	default:
		r = NewRecord(s)
	}

	// ignore nil record
	if r == nil {
		return nil
	}

	q := InsertIgnoreQuery(r)
	result, err := q.Do(tx)
	if err != nil {
		return fmt.Errorf("%v -> %v.", err, q)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		q := UpdateQuery(r)
		result, err := q.Do(tx)
		if err != nil {
			return err
		}

		n, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if n == 0 {
			return fmt.Errorf("failed to update row: %v.", q)
		}
	}

	return nil
}

func Delete(tx Execer, s interface{}) error {
	var r *Record
	switch s.(type) {
	case *Record:
		r = s.(*Record)
	default:
		r = NewRecord(s)
	}

	// ignore nil record
	if r == nil {
		return nil
	}

	q := DeleteQuery(r)
	_, err := q.Do(tx)
	if err != nil {
		return fmt.Errorf("%v -> %v.", err, q)
	}
	return nil
}

func InsertIgnoreQuery(r *Record) *Query {
	fields := r.Fields.Filter(DbType, NonNil)
	pkeys := r.Fields.Filter(Key)
	return &Query{
		Cmd: pg(query("INSERT INTO", quote(r.Name), brace(fieldList(fields)), "SELECT", placeholderList(fields),
			"WHERE NOT EXISTS", brace(query("SELECT 1 FROM", quote(r.Name), "WHERE", fieldEqualList(pkeys))))),
		Args: append(fields.Values(), pkeys.Values()...),
	}
}

func UpdateQuery(r *Record) *Query {
	pkeys := r.Fields.Filter(Key)
	rest := r.Fields.Filter(NonKey, DbType, NonNil)
	if len(rest) == 0 {
		rest = pkeys
	}
	return &Query{
		Cmd:  pg(query("UPDATE", quote(r.Name), "SET", brace(fieldList(rest)), "=", brace(placeholderList(rest)), "WHERE", fieldEqualList(pkeys))),
		Args: append(rest.Values(), pkeys.Values()...),
	}
}

func DeleteQuery(r *Record) *Query {
	fields := r.Fields.Filter(DbType, NonNil)
	return &Query{
		Cmd: pg(query("DELETE FROM", quote(r.Name), "WHERE",
			fieldEqualList(fields))),
		Args: fields.Values(),
	}
}

func fieldList(fs Fields) string {
	return list(len(fs), ", ", func(i int) string { return quote(fs[i].Name) })
}

func placeholderList(fs Fields) string {
	return list(len(fs), ", ", func(int) string { return "?" })
}

func fieldEqualList(fs Fields) string {
	return list(len(fs), " AND ",
		func(i int) string {
			return quote(fs[i].Name) + "=?"
		})
}

func query(args ...string) string {
	return strings.Join(args, " ")
}

func quote(s string) string {
	return `"` + s + `"`
}

func brace(s string) string {
	return "(" + s + ")"
}

func pg(s string) string {
	var b bytes.Buffer
	i := 1
	for _, r := range s {
		if r != '?' {
			b.WriteRune(r)
		} else {
			b.WriteString(fmt.Sprint("$", i))
			i++
		}
	}
	return b.String()
}

func list(cnt int, sep string, get func(i int) string) string {
	var b bytes.Buffer
	for i := 0; i < cnt-1; i++ {
		b.WriteString(get(i))
		b.WriteString(sep)
	}
	if cnt > 0 {
		b.WriteString(get(cnt - 1))
	}
	return b.String()
}
