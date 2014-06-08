package main

import (
	"fmt"
	"net/http"

	"github.com/hailiang/getgo"
	"github.com/hailiang/getgo/util"
	"github.com/hailiang/html-query"
	"github.com/hailiang/html-query/expr"
)

func main() {
	util.Run(golangBlogIndexTask{})
}

var (
	_Id    = expr.Id
	_Class = expr.Class
)

// golangBlogEntry represents a record for storing a blog entry.
type golangBlogEntry struct {
	Title string
	URL   string
	Tags  *string
}

// golangBlogIndexTask retrieves the blog index of Go blog.
type golangBlogIndexTask struct {
	// Task variables, e.g. page number
}

func (t golangBlogIndexTask) Request() *http.Request {
	return getReq(`http://blog.golang.org/index`)
}

func (t golangBlogIndexTask) Handle(root *query.Node, s getgo.Storer) (err error) {
	root.Div(_Id("content")).Children(_Class("blogtitle")).For(func(item *query.Node) {
		title := item.Ahref().Text()
		url := item.Ahref().Href()
		tags := item.Span(_Class("tags")).Text()
		if url != nil && title != nil {
			store(&golangBlogEntry{Title: *title, URL: *url, Tags: tags}, s, &err)
		}
	})
	return
}

func getReq(template string, args ...interface{}) *http.Request {
	req, err := http.NewRequest("GET", fmt.Sprintf(template, args...), nil)
	checkError(err)
	return req
}

func store(v interface{}, s getgo.Storer, err *error) {
	if *err == nil {
		*err = s.Store(v)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
