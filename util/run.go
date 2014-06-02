// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"bitbucket.org/hailiang/getgo"
	"log"
	"net/http"
)

func Run(tasks ...interface{}) {
	checkError(getgo.Run(runner(), PrinterTx{}, tasks...))
}

func runner() getgo.Runner {
	client := getgo.NewHttpLogger(&http.Client{})
	return getgo.SequentialRunner{
		client,
		getgo.ErrorHandlerFunc(func(req *http.Request, err error) error {
			log.Printf("Error: %v, Reqeust: %v.\n", err, req.URL)
			return nil
		})}
}

type PrinterTx struct {
}

func (PrinterTx) Store(v interface{}) error {
	pp(v)
	return nil
}

func (PrinterTx) Commit() error {
	p("Commited.")
	return nil
}

func (PrinterTx) Rollback() error {
	p("Rolled back.")
	return nil
}
