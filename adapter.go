// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package getgo

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hailiang/html-query"
)

// Atomized is an adapter that converts a StorableTask to an atomized Task that
// supports transaction.
type Atomized struct {
	StorableTask
	Tx
}

// Handle implements the Handle method of Task interface.
func (h Atomized) Handle(resp *http.Response) error {
	if resp == nil {
		return h.Tx.Rollback() // response is nil, rollback transaction.
	}
	if err := h.StorableTask.Handle(resp, h.Tx); err != nil {
		h.Tx.Rollback() // ignore rollback error.
		return err
	}
	return h.Tx.Commit()
}

// Storable is an adapter that converts a TextTask to a StorableTask.
type Storable struct {
	TextTask
}

// Handle implements the Handle method of StorableTask interface.
func (b Storable) Handle(resp *http.Response, s Storer) error {
	// Since an HTMLTask definitely uses response's body only, it requires that
	// status 20x is returned.
	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		// no-op.
	default:
		return fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	return b.TextTask.Handle(resp.Body, s)
}

// Text is an adapter that converts an HTMLTask to a TextTask.
type Text struct {
	HTMLTask
}

// Handle implements the Handle method of TextTask interface.
func (t Text) Handle(r io.Reader, s Storer) error {
	root, err := query.Parse(r)
	if err != nil {
		return err
	}
	return t.HTMLTask.Handle(root, s)
}

// ToTask adapts an HTMLTask, TextTask or Task itself to a Task.
func ToTask(t interface{}, tx Tx) Task {
	switch task := t.(type) {
	case HTMLTask:
		return Atomized{Storable{Text{task}}, tx}
	case TextTask:
		return Atomized{Storable{task}, tx}
	case Task:
		return task
	default:
		panic(errors.New("task is unknown type."))
	}

}

// ErrorHandlerFunc converts a function object to a ErrorHandler interface.
type ErrorHandlerFunc func(*http.Request, error) error

// HandleError implements ErrorHandler interface.
func (f ErrorHandlerFunc) HandleError(request *http.Request, err error) error {
	return f(request, err)
}
