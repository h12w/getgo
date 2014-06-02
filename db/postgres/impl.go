// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"
	"sync"
	"bitbucket.org/hailiang/getgo/db"
)

func Open(dataSourceName string) (db.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &dbImpl{db: db}, nil
}

type dbImpl struct {
	db *sql.DB
	mu sync.Mutex
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}

func (d *dbImpl) Begin() (db.Tx, error) {
	return &txImpl{db: d}, nil
}

func (d *dbImpl) Lock() {
	d.mu.Lock()
}

func (d *dbImpl) Unlock() {
	d.mu.Unlock()
}

type txImpl struct {
	db *dbImpl
	ss []interface{}
	ds []interface{}
	mu sync.Mutex
}

func (t *txImpl) Store(v interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ss = append(t.ss, v)
	return nil
}

func (t *txImpl) Delete(v interface{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ds = append(t.ds, v)
	return nil
}

func (t *txImpl) Commit() error {
	t.db.Lock()
	defer t.db.Unlock()

	rawdb := t.db.db
	tx, err := rawdb.Begin()
	if err != nil {
		return err
	}
	if err := storeRecords(t.ss, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := deleteRecords(t.ds, tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (t *txImpl) Rollback() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ss = nil
	t.ds = nil
	return nil
}

// TODO: merge storeRecords and Delete records

func storeRecords(vs []interface{}, tx *sql.Tx) error {
	if len(vs) == 0 {
		return nil
	}
	for _, v := range vs {
		if err := Upsert(tx, v); err != nil {
			return err
		}
	}
	return nil
}

func deleteRecords(vs []interface{}, tx *sql.Tx) error {
	if len(vs) == 0 {
		return nil
	}
	for _, v := range vs {
		if err := Delete(tx, v); err != nil {
			return err
		}
	}
	return nil
}
