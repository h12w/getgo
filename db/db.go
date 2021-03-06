// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package db contains common interface that all implementations under db directory
must satisfy.
*/
package db

// DB is a general interface for arbitrary database type.
type DB interface {
	Begin() (Tx, error) // Begin returns an object of transaction interface.
}

// Storer provides the Store method to store an object parsed from an HTTP
// response.
type Storer interface {
	Store(v interface{}) error
}

// Deleter provides the Delete method to delete an object.
type Deleter interface {
	Delete(v interface{}) error
}

// Tx is a transaction interface that provides methods for storing objects,
// commit or rollback changes. Tx's implementation must allow concurrent use.
type Tx interface {
	Storer
	Deleter
	Commit() error
	Rollback() error
}
