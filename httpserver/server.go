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

package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

type Filter interface {
	Before(rsp *Response, r *Request) bool
	After(rsp *Response, r *Request)
}

type Request struct {
	*http.Request
}

type Response struct {
	b         []byte
	rw        http.ResponseWriter
	code      int
	writeCode bool
}

func (a *Response) Write(b []byte) {
	if a.b == nil {
		a.b = make([]byte, 0, 1024)
	}
	a.b = append(a.b, b...)
}

func (a *Response) Header() http.Header {
	return a.rw.Header()
}

func (a *Response) WriteHeader(statusCode int) {
	a.code = statusCode
	a.writeCode = true
}

func (a *Response) Byte() []byte {
	rsp.Write([]byte("page not found"))
	return a.b
}

func defaultNotFound(rsp *Response, req *Request) (uint, error) {
	return 404, nil
}

type handler struct {
	plainFilters        map[string][]Filter
	regexpFilters       map[string][]Filter
	sortedFilterPattern []string

	plainHandlers        map[string]func(rsp *Response, req *Request) (uint, error)
	regexpHandlers       map[string]func(rsp *Response, req *Request) (uint, error) // raw pattern -> func
	sortedHandlerPattern []string                                                   // regexp string
	sortedHandlerRegexp  []regexp.Regexp
	regexpHandlerPattern map[string]string   // regexp string -> raw pattern
	params               map[string][]string // raw pattern -> params
}

func (a *handler) prepare() {
	s := ``
	for regexpStr, _ := range a.regexpHandlers {

	}
}

func (a *handler) route(pattern string, f func(rsp *Response, req *Request) (uint, error)) {
	a.plainHandlers[pattern] = f
}

func (a *handler) matchPlainHandler(url string) func(rsp *Response, req *Request) (uint, error) {
	for k, f := range a.handlers {
		if k == url {
			return f
		}
	}
	return nil
}

func (a *handler) matchRegexpHandler(url string) func(rsp *Response, req *Request) (uint, error) {
	for i, pattern := range a.sortedRegexp {
		if pattern.MatchString(url) {
			return a.regexpHandlers[a.regexpPattern[a.sortedPattern[i]]]
		}
	}
	return nil
}

func (a *handler) matchPlainFilter(url string) []Filter {
	for k, f := range a.plainFilters {
		if k == url {
			return f
		}
	}
	return nil
}

func (a *handler) matchRegexpFilter(url string) []Filter {
	for k, f := range a.regexpFilters {
		if k.MatchString(url) {
			return f
		}
	}
	return nil
}

func (a *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	url, err := url.QueryUnescape(r.URL.Path)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	f := a.matchPlainHandler(url)

	if f == nil {
		f = notFound
	}

	filters := a.matchPlainFilter(url)
	if filters == nil || len(filters) <= 0 {
		filters = a.matchRegexpFilter(r.URL)
	}

	req := &Request{r}
	rsp := &Response{rw: rw}

	if filters != nil && len(f) > 0 {
		for _, f := range filters {
			if !f.Before(rsp, req) {
				return
			}
		}
	}

	code, err := f(rsp, req)

	if filters != nil && len(f) > 0 {
		for _, f := range filters {
			f.After(rsp, req)
		}
	}

	if rsp.writeCode {
		rsp.rw.WriteHeader(rsp.code)
	}

	rsp.rw.Write(rsp.b)
	fmt.Println(code, err)
}

func NewHTTPServer(port int) *HTTPServer {
	h := &handler{f: defaultHandlerFunc}
	http.Handle("/", h)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: http.DefaultServeMux,
	}
	return &HTTPServer{s: s, defaultHandler: h, handlers: make(map[string]*handler), block: true}
}

func defaultHandlerFunc(rsp *Response, req *Request) (uint, error) {
	url := req.URL
	// for k, v := range
	return 0, nil
}

type HTTPServer struct {
	s              *http.Server
	defaultHandler *handler
	handlers       map[string]*handler
	block          bool
	finish         func(err error)
}

func (a *HTTPServer) Route(pattern string, f func(rsp *Response, req *Request) (uint, error)) {
	h := &handler{f: f}
	a.handlers[pattern] = h
}

func (a *HTTPServer) AddPatternFilter(r *regexp.Regexp, f Filter) {
	for k, h := range a.handlers {
		if r.Match([]byte(k)) {
			h.addFilter(f)
		}
	}
}

func (a *HTTPServer) AddFilter(pattern string, f Filter) {
	for k, h := range a.handlers {
		if k == pattern {
			h.addFilter(f)
		}
	}
}

func (a *HTTPServer) ServeHTTP() {
	if !a.block {
		go func() {
			a.doServeHTTP()
		}()
		return
	}
	a.doServeHTTP()
}

func (a *HTTPServer) ServeHTTPS(certFile, keyFile string) {
	if !a.block {
		go func() {
			a.doServeHTTPS(certFile, keyFile)
		}()
		return
	}
	a.doServeHTTPS(certFile, keyFile)
}

func (a *HTTPServer) doServeHTTP() {
	err := a.s.ListenAndServe()
	if err != nil {
		fmt.Printf("ListenAndServe error:%v\n", err.Error())
	}
}

func (a *HTTPServer) doServeHTTPS(certFile, keyFile string) {
	err := a.s.ListenAndServeTLS(certFile, keyFile)
	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("ListenAndServe error:%v\n", err.Error())
	}
}

func (a *HTTPServer) Shutdown() {
	err := a.s.Shutdown(context.TODO())
	a.finish(err)
}

func (a *HTTPServer) Finish(f func(e error)) {
	a.block = false
	a.finish = f
}
