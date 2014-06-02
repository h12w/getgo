// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package getgo

import (
	"github.com/hailiang/html-query"
	"io"
	"net/http"
)

// An HTTP crawler task.
// It must provide an HTTP request and a method to handle an HTTP response.
type Task interface {
	Requester
	Handle(resp *http.Response) error
}

// Runner runs Tasks.
// A Runner gets an HTTP request from a Task, get the HTTP response and pass the
// response to the Task's Handle method. When a runner failed to get a response
// object, a nil response must still be passed to the Handle method to notify
// that a transaction must be rolled back if any.
type Runner interface {
	Run(task Task) error
	Close()
}

// A task that should be able to store data with a Storer passed to the Handle
// method.
type StorableTask interface {
	Requester
	Handle(resp *http.Response, s Storer) error
}

// A task that only retrieves a Response's body.
type TextTask interface {
	Requester
	Handle(r io.Reader, s Storer) error
}

// An HTML task should be able to Parse an HTML node tree to a slice of objects.
type HtmlTask interface {
	Requester
	Handle(root *query.Node, s Storer) error
}

// Requester is the interface that returns an HTTP request by Request method.
// The Request method must be implemented to allow repeated calls.
type Requester interface {
	Request() *http.Request
}

// Storer provides the Store method to store an object parsed from an HTTP
// response.
type Storer interface {
	Store(v interface{}) error
}

// Tx is a transaction interface that provides methods for storing objects,
// commit or rollback changes. Notice that there is no Delete method defined.
// Tx's implementation must allow concurrent use.
type Tx interface {
	Storer
	Commit() error
	Rollback() error
}

// Doer processes an HTTP request and returns an HTTP response.
type Doer interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// ErrorHandler is used to call back an external error handler when a task
// fails.
type ErrorHandler interface {
	HandleError(request *http.Request, err error) error
}
