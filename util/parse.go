// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hailiang/html-query"
	ex "github.com/hailiang/html-query/expr"
)

// MustLoadHTML loads and parses an HTML file.
func MustLoadHTML(file string) *query.Node {
	f, err := os.Open(file)
	checkError(err)
	defer f.Close()

	node, err := query.Parse(f)
	checkError(err)
	return node
}

// MustLoadJSON loads and parses a JSON file.
func MustLoadJSON(file string, v interface{}) {
	f, err := os.Open(file)
	checkError(err)
	defer f.Close()

	err = json.NewDecoder(f).Decode(&v)
	checkError(err)
}

// DumpAll dumps all the nodes in the HTML DOM that satisfy the checkers provided.
func DumpAll(n *query.Node, cs ...ex.Checker) {
	it := n.Descendants(cs...)
	for node := it.Next(); node != nil; node = it.Next() {
		path := []*query.Node{}
		for par := node; par != nil; par = par.Parent() {
			path = append([]*query.Node{par}, path...)
		}
		for level, n := range path {
			for i := 0; i < level-1; i++ {
				fmt.Print("    ")
			}
			fmt.Println(*n.RenderTagOnly())
		}
	}
}

// DumpAllText dumps all the text node in the HTML DOM that satisfy the regular
// expression provided.
func DumpAllText(n *query.Node, pat string) {
	DumpAll(n, append([]ex.Checker{ex.TextNode}, ex.Text(pat))...)
}
