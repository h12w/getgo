// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package postgres

import (
	"fmt"
	"testing"

	"github.com/hailiang/getgo/db/schema"
)

func TestSqlstore(t *testing.T) {
	type StructType struct {
		ID      int `sql:"pk"`
		IDOther int `sql:"pk"`
		SVal    string
		IVal    int
		PVal    *int
	}
	s := &StructType{ID: 1, IDOther: 2, SVal: "S", IVal: 9}
	r := schema.NewRecord(s) //.SetKey("Id", "IdOther")

	fmt.Println(r)
	for _, f := range r.Fields {
		fmt.Println(f)
	}

	fmt.Println(insertIgnoreQuery(r))
	fmt.Println(updateQuery(r))
	fmt.Println(deleteQuery(r))
}
