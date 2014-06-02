// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package getgo

import (
	"errors"
	"reflect"
	"sync"
)

// TaskGroup makes a group of StorableTask as a single transaction.
type TaskGroup struct {
	tasks []StorableTask
	Tx
}

func NewTaskGroup(tx Tx) *TaskGroup {
	return &TaskGroup{Tx: tx}
}

// Add a StorableTask to TaskGroup.
func (g *TaskGroup) Add(task StorableTask) {
	g.tasks = append(g.tasks, task)
}

// Run all tasks within a TaskGroup.
func (g *TaskGroup) Run(runner Runner) error {
	if len(g.tasks) == 0 {
		return nil
	}
	gtx := newGroupTx(len(g.tasks), g)
	for _, task := range g.tasks {
		if err := runner.Run(Atomized{task, gtx}); err != nil {
			return err
		}
	}
	return nil
}

// groupTx is used for each task within TaskGroup.
type groupTx struct {
	cnt    int
	result bool
	tx     Tx
	mu     sync.Mutex
}

func newGroupTx(size int, tx Tx) *groupTx {
	return &groupTx{cnt: size, result: true, tx: tx}
}

func (t *groupTx) Store(v interface{}) error {
	return t.tx.Store(v)
}

func (t *groupTx) done(result bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.result = t.result && result
	t.cnt--
	if t.cnt == 0 {
		if t.result {
			return t.tx.Commit()
		} else {
			return t.tx.Rollback()
		}
	}
	return nil
}

func (t *groupTx) Commit() error {
	return t.done(true)
}

func (t *groupTx) Rollback() error {
	return t.done(false)
}

// Add either HtmlTask, TextTask or StorableTask to TaskGroup.
func addTask(task interface{}, g *TaskGroup) {
	switch t := task.(type) {
	case HtmlTask:
		g.Add(Storable{Text{t}})
	case TextTask:
		g.Add(Storable{t})
	case StorableTask:
		g.Add(t)
	default:
		panic(errors.New("task is unexpected type: " +
			reflect.TypeOf(task).Name()))
	}
}
