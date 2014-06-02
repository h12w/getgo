// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package getgo

import (
	"net/http"
	"sync"
)

// Retry number when failed to fetch a page.
var RetryNum = 3

// A simple single threaded task runner.
type SequentialRunner struct {
	Client Doer
	ErrorHandler
}

func (r SequentialRunner) Run(task Task) error {
	req := task.Request()
	resp, err := r.Client.Do(req)
	if err != nil {
		task.Handle(nil) // notify that the fetch has failed, ignore the error.
		if err = r.HandleError(req, err); err != nil {
			return err
		} else {
			return nil
		}
	}
	defer resp.Body.Close()
	if err := task.Handle(resp); err != nil {
		if err = r.HandleError(req, err); err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func (r SequentialRunner) Close() {
}

// Concurrent runner.
type ConcurrentRunner struct {
	seq SequentialRunner
	ch  chan Task
	wg  *sync.WaitGroup
}

func NewConcurrentRunner(workerNum int, client Doer, errHandler ErrorHandler) ConcurrentRunner {
	r := ConcurrentRunner{SequentialRunner{RetryDoer{client, RetryNum}, errHandler}, make(chan Task), new(sync.WaitGroup)}
	r.wg.Add(workerNum)
	for i := 0; i < workerNum; i++ {
		go r.work()
	}
	return r
}

func (r ConcurrentRunner) Run(task Task) error {
	r.ch <- task
	return nil
}

func (r ConcurrentRunner) Close() {
	close(r.ch)
	r.wg.Wait()
}

func (r ConcurrentRunner) work() {
	defer r.wg.Done()
	for task := range r.ch {
		_ = r.seq.Run(task) // errors are ignored here, handled by error handler.
	}
}

// RetryDoer wraps a Doer and implements the retry operation for Do method.
type RetryDoer struct {
	Doer
	RetryTime int
}

func (d RetryDoer) Do(req *http.Request) (resp *http.Response, err error) {
	for i := 0; i < d.RetryTime; i++ {
		resp, err = d.Doer.Do(req)
		if err == nil {
			return resp, nil
		}
	}
	return nil, err
}

// Run either HtmlTask, TextTask or Task. tx is commited if successful or
// rollbacked if failed.
func Run(runner Runner, tx Tx, tasks ...interface{}) error {
	switch len(tasks) {
	case 0:
		return nil
	case 1:
		return runner.Run(ToTask(tasks[0], tx))
	}
	// more than 1 tasks
	tg := NewTaskGroup(tx)
	for _, task := range tasks {
		addTask(task, tg)
	}
	return tg.Run(runner)
}
