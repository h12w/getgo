// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"log"
	"net/http"

	"github.com/hailiang/getgo"
)

// Run runs tasks and print the data fetched.
func Run(tasks ...interface{}) {
	checkError(getgo.Run(runner(), printerTx{}, tasks...))
}

func runner() getgo.Runner {
	client := getgo.NewHTTPLogger(&http.Client{})
	return getgo.SequentialRunner{
		Client: client,
		ErrorHandler: getgo.ErrorHandlerFunc(func(req *http.Request, err error) error {
			log.Printf("Error: %v, Reqeust: %v.\n", err, req.URL)
			return nil
		})}
}

type printerTx struct {
}

func (printerTx) Store(v interface{}) error {
	pp(v)
	return nil
}

func (printerTx) Commit() error {
	p("Commited.")
	return nil
}

func (printerTx) Rollback() error {
	p("Rolled back.")
	return nil
}
