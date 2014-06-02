// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"fmt"
	"testing"
	"getgo/db"
)

func Test_sqlstore(t *testing.T) {
	type StructType struct {
		Id      int `sql:"pk"`
		IdOther int `sql:"pk"`
		SVal    string
		IVal    int
		PVal    *int
	}
	s := &StructType{Id: 1, IdOther: 2, SVal: "S", IVal: 9}
	r := db.NewRecord(s)//.SetKey("Id", "IdOther")

	fmt.Println(r)
	for _, f := range r.Fields {
		fmt.Println(f)
	}

	fmt.Println(InsertIgnoreQuery(r))
	fmt.Println(UpdateQuery(r))
	fmt.Println(DeleteQuery(r))
}
