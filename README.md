Getgo: a concurrent, simple and extensible web scraping framework
=================================================================
[![GoDoc](https://godoc.org/github.com/hailiang/getgo?status.png)](https://godoc.org/github.com/hailiang/getgo)
[![Build Status](https://travis-ci.org/hailiang/getgo.svg?branch=master)](https://travis-ci.org/hailiang/getgo)

Getgo is a concurrent, simple and extensible web scraping framework written in [Go](http://golang.org).

Quick start
-----------
###Get Getgo
```bash
go get -u github.com/hailiang/getgo
```

###Define a task
This example is under the examples/goblog directory. To use Getgo to scrap structured
data from a web page, just define the structured data as a Go struct (golangBlogEntry),
and define a corresponding task (golangBlogIndexTask).
```go
type golangBlogEntry struct {
	Title string
	URL   string
	Tags  *string
}

type golangBlogIndexTask struct {
	// Variables in task URL, e.g. page number
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
```

###Run the task
Use util.Run to run the task and print all the result to standard output.
```go
	util.Run(golangBlogIndexTask{})
```
To store the parsed result to a database, a storage backend satisfying getgo.Tx
interface should be provided to the getgo.Run method.

Understand Getgo
----------------
A getgo.Task is an interface to represent an HTTP crawler task that provides an
HTTP request and a method to handle the HTTP response.
```go
type Task interface {
	Requester
	Handle(resp *http.Response) error
}

type Requester interface {
	Request() *http.Request
}
```

A getgo.Runner is responsible to run a getgo.Task. There are two concrete runners
provided: SequentialRunner and ConcurrentRunner.
```go
type Runner interface {
	Run(task Task) error // Run runs a task
	Close()              // Close closes the runner
}
```

A task that stores data into a storage backend should satisfy getgo.StorableTask
interface.
```go
type StorableTask interface {
	Requester
	Handle(resp *http.Response, s Storer) error
}
```

A storage backend is simply an object satisfying getgo.Tx interface.
```go
type Storer interface {
	Store(v interface{}) error
}

type Tx interface {
	Storer
	Commit() error
	Rollback() error
}
```

See getgo.Run method to understand how a StorableTask is combined with a storage
backend and adapted to become a normal Task to allow a Runner to run it.

There are currently a PostgreSQL storage backend provided by Getgo, and it is
not hard to support more backends (See getgo/db package for details).

The easier way to define a task for an HTML page is to define a task satisfying
getgo.HTMLTask rather than getgo.Task, there are adapters to convert internally
an HTMLTask to a Task so that a Runner can run an HTMLTask. The Handle method of
HTMLTask provides an already parsed HTML DOM object (by html-query package).
```go
type HTMLTask interface {
	Requester
	Handle(root *query.Node, s Storer) error
}
```

Similarly, a task for retrieving a JSON page should satisfy getgo.TextTask
interface. An io.Reader is provided to be decoded by the encoding/json package.
```go
type TextTask interface {
	Requester
	Handle(r io.Reader, s Storer) error
}
```
