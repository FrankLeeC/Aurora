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
	return a.b
}

func defaultNotFound(rsp *Response, req *Request) (uint, error) {
	return 404, nil
}

type handler struct {
	plainFilters        map[string][]Filter
	regexFilters        map[string][]Filter
	sortedFilterPattern []string // regex string
	sortedFilterRegex   []*regexp.Regexp
	filterRegexPattern  map[string]string   // regex string -> raw pattern
	filterPatternParams map[string][]string // raw pattern -> params

	plainHandlers        map[string]func(rsp *Response, req *Request) (uint, error)
	regexHandlers        map[string]func(rsp *Response, req *Request) (uint, error) // raw pattern -> func
	sortedHandlerPattern []string                                                   // regex string
	sortedHandlerRegex   []*regexp.Regexp
	handlerRegexPattern  map[string]string   // regex string -> raw pattern
	handlerPatternParams map[string][]string // raw pattern -> params
}

func quicksort(a []string) {
	if len(a) <= 1 {
		return
	}
	i := 0
	j := len(a) - 1
	p := 0
	for i <= j {
		for j >= 0 {
			if len(a[j]) > len(a[p]) {
				a[j], a[p] = a[p], a[j]
				p = j
				j--
				break
			}
			j--
		}
		for i <= j {
			if len(a[i]) < len(a[p]) {
				a[i], a[p] = a[p], a[i]
				p = i
				i++
				break
			}
			i++
		}
	}
	quicksort(a[0:p])
	quicksort(a[p+1:])
}

func (a *handler) preparaHandlers() {
	if len(a.regexHandlers) <= 0 {
		return
	}
	preg := regexp.MustCompile(`({\w+})`)
	a.handlerPatternParams = make(map[string][]string, len(a.regexHandlers))
	a.handlerRegexPattern = make(map[string]string, len(a.regexHandlers))
	a.sortedHandlerPattern = make([]string, 0, len(a.regexHandlers))

	for rawPattern := range a.regexHandlers {
		params := preg.FindAllString(rawPattern, -1)
		if params != nil && len(params) > 0 {
			for i := range params {
				params[i] = params[i][1 : len(params[i])-1]
			}
			a.handlerPatternParams[rawPattern] = params
			realRegex := "^" + preg.ReplaceAllString(rawPattern, `([^/]+)`) + "$"
			a.handlerRegexPattern[realRegex] = rawPattern
			a.sortedHandlerPattern = append(a.sortedHandlerPattern, realRegex)
		} else {
			a.plainHandlers[rawPattern] = a.regexHandlers[rawPattern]
		}
	}
	quicksort(a.sortedHandlerPattern)
	for _, realRegex := range a.sortedHandlerPattern {
		a.sortedHandlerRegex = append(a.sortedHandlerRegex, regexp.MustCompile(realRegex))
	}
}

func (a *handler) prepareFilters() {
	if len(a.regexFilters) <= 0 {
		return
	}
	preg := regexp.MustCompile(`({\w+})`)
	a.filterPatternParams = make(map[string][]string, len(a.regexFilters))
	a.filterRegexPattern = make(map[string]string, len(a.regexFilters))
	a.sortedFilterPattern = make([]string, 0, len(a.regexFilters))

	for rawPattern := range a.regexFilters {
		params := preg.FindAllString(rawPattern, -1)
		if params != nil && len(params) > 0 {
			for i := range params {
				params[i] = params[i][1 : len(params[i])-1]
			}
			a.filterPatternParams[rawPattern] = params
			realRegex := "^" + preg.ReplaceAllString(rawPattern, `([^/]+)`) + "$"
			a.filterRegexPattern[realRegex] = rawPattern
			a.sortedFilterPattern = append(a.sortedFilterPattern, realRegex)
		} else {
			a.plainFilters[rawPattern] = a.regexFilters[rawPattern]
		}
	}
	quicksort(a.sortedFilterPattern)
	for _, realRegex := range a.sortedFilterPattern {
		a.sortedFilterRegex = append(a.sortedFilterRegex, regexp.MustCompile(realRegex))
	}
}

func (a *handler) prepare() {
	a.prepareFilters()
	a.preparaHandlers()
}

func (a *handler) route(pattern string, f func(rsp *Response, req *Request) (uint, error)) {
	a.plainHandlers[pattern] = f
}

func (a *handler) matchPlainHandler(url string) func(rsp *Response, req *Request) (uint, error) {
	for k, f := range a.plainHandlers {
		if k == url {
			return f
		}
	}
	return nil
}

func (a *handler) matchRegexpHandler(url string) func(rsp *Response, req *Request) (uint, error) {
	for i, pattern := range a.sortedHandlerRegex {
		if pattern.MatchString(url) {
			return a.regexHandlers[a.handlerRegexPattern[a.sortedHandlerPattern[i]]]
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
