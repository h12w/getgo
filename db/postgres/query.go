// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	sc "github.com/hailiang/getgo/db/schema"
)

// execer is an interface that satisfies the Exec method of sql.Tx.
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// query stores a SQL query.
type query struct {
	Cmd  string
	Args []interface{}
}

// Do executes the query on an execer provided as an argument.
func (q *query) Do(ex execer) (sql.Result, error) {
	return ex.Exec(q.Cmd, q.Args...)
}

// upsert insert a record or update it if given primary key exists.
// Limitation: The current implementation does not handle confilicts between
// transactions. So there must be an external lock to make sure there is only
// one transaction at the same time.
func upsert(tx execer, s interface{}) error {
	var r *sc.Record
	switch s.(type) {
	case *sc.Record:
		r = s.(*sc.Record)
	default:
		r = sc.NewRecord(s)
	}

	// ignore nil record
	if r == nil {
		return nil
	}

	q := insertIgnoreQuery(r)
	result, err := q.Do(tx)
	if err != nil {
		return fmt.Errorf("%v -> %v.", err, q)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		q := updateQuery(r)
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

// delete deletes a record of a given primary key.
func deleteRecord(tx execer, s interface{}) error {
	var r *sc.Record
	switch s.(type) {
	case *sc.Record:
		r = s.(*sc.Record)
	default:
		r = sc.NewRecord(s)
	}

	// ignore nil record
	if r == nil {
		return nil
	}

	q := deleteQuery(r)
	_, err := q.Do(tx)
	if err != nil {
		return fmt.Errorf("%v -> %v.", err, q)
	}
	return nil
}

// insertIgnoreQuery returns a query that inserts a record when the primary keys
// not exist, otherwise ignore it.
func insertIgnoreQuery(r *sc.Record) *query {
	fields := r.Fields.Filter(sc.DbType, sc.NonNil)
	pkeys := r.Fields.Filter(sc.Key)
	return &query{
		Cmd: pg(join("INSERT INTO", quote(r.Name), brace(fieldList(fields)), "SELECT", placeholderList(fields),
			"WHERE NOT EXISTS", brace(join("SELECT 1 FROM", quote(r.Name), "WHERE", fieldEqualList(pkeys))))),
		Args: append(fields.Values(), pkeys.Values()...),
	}
}

// updateQuery returns a query that updates a record of the same primary keys.
func updateQuery(r *sc.Record) *query {
	pkeys := r.Fields.Filter(sc.Key)
	rest := r.Fields.Filter(sc.NonKey, sc.DbType, sc.NonNil)
	if len(rest) == 0 {
		rest = pkeys
	}
	return &query{
		Cmd:  pg(join("UPDATE", quote(r.Name), "SET", brace(fieldList(rest)), "=", brace(placeholderList(rest)), "WHERE", fieldEqualList(pkeys))),
		Args: append(rest.Values(), pkeys.Values()...),
	}
}

// deleteQuery return a query that deletes a record.
func deleteQuery(r *sc.Record) *query {
	fields := r.Fields.Filter(sc.DbType, sc.NonNil)
	return &query{
		Cmd: pg(join("DELETE FROM", quote(r.Name), "WHERE",
			fieldEqualList(fields))),
		Args: fields.Values(),
	}
}

func fieldList(fs sc.Fields) string {
	return list(len(fs), ", ", func(i int) string { return quote(fs[i].Name) })
}

func placeholderList(fs sc.Fields) string {
	return list(len(fs), ", ", func(int) string { return "?" })
}

func fieldEqualList(fs sc.Fields) string {
	return list(len(fs), " AND ",
		func(i int) string {
			return quote(fs[i].Name) + "=?"
		})
}

func join(args ...string) string {
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
