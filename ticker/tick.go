/*
MIT License

Copyright (c) 2018 Frank Lee

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package ticker ticker
package ticker

import "time"

import "errors"

// Refresher refresh
type Refresher interface {
	Refresh()
}

// Option ticker option
type Option struct {
	MaxRetry int           // max count
	Duration time.Duration // interval
}

// Ticker ticker
type Ticker struct {
	h *handler
}

// Start start ticker
func (a *Ticker) Start() {
	go a.h.run()

}

// Stop stop ticker
func (a *Ticker) Stop() {
	a.h.s <- true
}

// New run a refresh tick
func New(r Refresher, p *Option) *Ticker {
	if r == nil {
		panic(errors.New("nil Refresher"))
	}
	if p == nil {
		p = defaultOption()
	}
	a := &handler{r, p.Duration, p.MaxRetry, 0, make(chan bool)}
	return &Ticker{a}
}

func defaultOption() *Option {
	return &Option{2, time.Second * 1}
}

type handler struct {
	r Refresher
	t time.Duration
	c int
	i int
	s chan bool
}

func (a *handler) restart() {
	if a.c == 0 {
		return
	}
	if a.c < 0 {
		go a.run()
	}
	if a.i < a.c {
		a.i++
		go a.run()
	}
}

func (a *handler) run() {
	defer func() {
		e := recover()
		if err, ok := e.(error); ok && err != nil {
			a.restart()
		}
	}()
	for {
		select {
		case <-time.After(a.t):
			a.r.Refresh()
		case <-a.s:
			goto over
		}
	}
over:
	return
}
