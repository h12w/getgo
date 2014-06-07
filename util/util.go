// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"fmt"
)

func p(v ...interface{}) {
	fmt.Println(v...)
}

func pp(v ...interface{}) {
	for _, item := range v {
		fmt.Println(toJSON(item))
	}
}

func toJSON(v interface{}) string {
	s, _ := json.MarshalIndent(v, "", "  ")
	return string(s)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func checkErrorf(cond bool, format string, a ...interface{}) {
	if cond {
		panic(fmt.Errorf(format, a...))
	}
}
