// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package getgo

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const avgCounterCap = 50

// HTTPLogger wraps an HTTP client and logs the request and network speed.
type HTTPLogger struct {
	client         Doer
	startTime      time.Time
	mu             sync.Mutex
	totalReqCount  int
	totalByteCount int
	avgByteCounter *avgCounter
}

// NewHTTPLogger creates an HTTPLogger by inspecting the connection's Read
// method of an http.Client.
func NewHTTPLogger(client *http.Client) *HTTPLogger {
	httpLogger := &HTTPLogger{
		client:         client,
		startTime:      time.Now(),
		avgByteCounter: newAvgCounter()}
	if client.Transport == nil {
		client.Transport = &http.Transport{}
	}
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.Dial = httpLogger.wrappedDial
	}
	return httpLogger
}
func (l *HTTPLogger) wrappedDial(network, address string) (net.Conn, error) {
	conn, err := net.Dial(network, address)
	if err == nil {
		return &readFilter{conn, l.gotBytes}, err
	}
	return conn, err
}

func (l *HTTPLogger) gotBytes(b []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.totalByteCount += len(b)
	l.avgByteCounter.Add(len(b), time.Now())
}

// Do implements the Doer interface.
func (l *HTTPLogger) Do(req *http.Request) (resp *http.Response, err error) {
	resp, err = l.client.Do(req)
	l.measure(resp, err)
	l.log(req)
	return resp, err
}

func (l *HTTPLogger) measure(resp *http.Response, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err == nil {
		l.totalReqCount++
	}
}

func (l *HTTPLogger) log(req *http.Request) {
	sec := time.Now().Sub(l.startTime).Seconds()
	reqSpeed := int(float64(l.totalReqCount) / sec)
	kbSpeed := int(float64(l.totalByteCount) / sec / 1000)
	fmt.Printf("[%dKB, %d, %dKB] %s\n",
		int(l.avgByteCounter.PerSecond()/1000), reqSpeed, kbSpeed, req.URL)
}

type readFilter struct {
	net.Conn
	read func(b []byte)
}

func (l *readFilter) Read(b []byte) (n int, err error) {
	n, err = l.Conn.Read(b)
	if err == nil {
		l.read(b[:n])
	}
	return n, err
}

type nt struct {
	n int
	t time.Time
}

type avgCounter struct {
	ring  []nt
	tail  int
	total int
	mu    sync.Mutex
}

func newAvgCounter() *avgCounter {
	ring := make([]nt, avgCounterCap)
	ring[len(ring)-1].t = time.Now()
	return &avgCounter{ring: ring}
}

func (c *avgCounter) Add(n int, t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total -= c.ring[c.tail].n
	c.total += n
	c.ring[c.tail] = nt{n, t}
	c.tail++
	if c.tail >= len(c.ring) {
		c.tail = 0
	}
}

func (c *avgCounter) PerSecond() float64 {
	head, tail := c.tail-1, c.tail
	if head == -1 {
		head = len(c.ring) - 1
	}
	sec := c.ring[head].t.Sub(c.ring[tail].t).Seconds()
	if sec == 0 {
		return 0
	}
	return float64(c.total) / sec
}
