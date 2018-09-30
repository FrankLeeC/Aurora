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
	DynamicParams map[string]string
}

type Response struct {
	b            []byte
	rw           http.ResponseWriter
	code         int
	writeBytes   bool
	writeCode    bool
	returnedCode uint
	err          error
}

func (a *Response) Write(b []byte) {
	a.writeBytes = true
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
	rsp.WriteHeader(404)
	rsp.Write([]byte(`page not found`))
	return 404, nil
}

type handler struct {
	notFound            func(rsp *Response, req *Request) (uint, error)
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

func newHandler() *handler {
	return &handler{
		notFound:      defaultNotFound,
		plainFilters:  make(map[string][]Filter),
		regexFilters:  make(map[string][]Filter),
		plainHandlers: make(map[string]func(rsp *Response, req *Request) (uint, error)),
		regexHandlers: make(map[string]func(rsp *Response, req *Request) (uint, error)),
	}
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
	a.sortedHandlerRegex = make([]*regexp.Regexp, 0, len(a.sortedHandlerPattern))
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
	a.sortedFilterRegex = make([]*regexp.Regexp, 0, len(a.sortedFilterPattern))
	for _, realRegex := range a.sortedFilterPattern {
		a.sortedFilterRegex = append(a.sortedFilterRegex, regexp.MustCompile(realRegex))
	}
}

func (a *handler) prepare() {
	a.prepareFilters()
	a.preparaHandlers()
}

func (a *handler) route(path string, f func(rsp *Response, req *Request) (uint, error)) {
	a.plainHandlers[path] = f
}

func (a *handler) dynamicRoute(pattern string, f func(rsp *Response, req *Request) (uint, error)) {
	a.regexHandlers[pattern] = f
}

func (a *handler) filter(path string, f Filter) {
	var fs []Filter
	var c bool
	if fs, c = a.plainFilters[path]; !c {
		fs = make([]Filter, 0)
	}
	fs = append(fs, f)
	a.plainFilters[path] = fs

}

func (a *handler) dynamicFilter(pattern string, f Filter) {
	var fs []Filter
	var c bool
	if fs, c = a.regexFilters[pattern]; !c {
		fs = make([]Filter, 0)
	}
	fs = append(fs, f)
	a.regexFilters[pattern] = fs

}

func (a *handler) matchPlainHandler(url string) func(rsp *Response, req *Request) (uint, error) {
	for k, f := range a.plainHandlers {
		if k == url {
			return f
		}
	}
	return nil
}

func (a *handler) matchRegexpHandler(url string) (*regexp.Regexp, []string, func(rsp *Response, req *Request) (uint, error)) {
	for i, regex := range a.sortedHandlerRegex {
		if regex.MatchString(url) {
			pattern := a.handlerRegexPattern[a.sortedHandlerPattern[i]]
			return regex, a.handlerPatternParams[pattern], a.regexHandlers[pattern]
		}
	}
	return nil, nil, nil
}

func (a *handler) matchPlainFilter(url string) []Filter {
	for k, f := range a.plainFilters {
		if k == url {
			return f
		}
	}
	return nil
}

func (a *handler) matchRegexpFilter(url string) (*regexp.Regexp, []string, []Filter) {
	for i, regex := range a.sortedFilterRegex {
		if regex.MatchString(url) {
			pattern := a.filterRegexPattern[a.sortedFilterPattern[i]]
			return regex, a.filterPatternParams[pattern], a.regexFilters[pattern]
		}
	}

	return nil, nil, nil
}

func (a *handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	url, err := url.QueryUnescape(r.URL.Path)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	filters := a.matchPlainFilter(url)
	var filterParams []string
	var filterRegex *regexp.Regexp
	if filters == nil || len(filters) <= 0 {
		filterRegex, filterParams, filters = a.matchRegexpFilter(url)
	}

	h := a.matchPlainHandler(url)
	var handlerParams []string
	var handlerRegex *regexp.Regexp
	if h == nil {
		handlerRegex, handlerParams, h = a.matchRegexpHandler(url)
	}

	req := &Request{r, make(map[string]string)}
	rsp := &Response{rw: rw}

	if filters != nil && len(filters) > 0 {
		if filterRegex != nil {
			tmp := filterRegex.FindStringSubmatch(url)
			values := make([]string, 0)
			if len(tmp) > 1 {
				for i := 1; i < len(tmp); i++ {
					values = append(values, tmp[i])
				}
			}
			for i := range values {
				req.DynamicParams[filterParams[i]] = values[i]
			}
		}

		for _, f := range filters {
			if !f.Before(rsp, req) {
				return
			}
		}
	}

	if h != nil {
		if handlerRegex != nil {
			tmp := handlerRegex.FindStringSubmatch(url)
			values := make([]string, 0)
			if len(tmp) > 1 {
				for i := 1; i < len(tmp); i++ {
					values = append(values, tmp[i])
				}
			}
			for i := range values {
				if _, c := req.DynamicParams[handlerParams[i]]; !c {
					req.DynamicParams[handlerParams[i]] = values[i]
				}
			}
		}
		rsp.returnedCode, rsp.err = h(rsp, req)
	} else {
		rsp.returnedCode, rsp.err = a.notFound(rsp, req)
	}

	if filters != nil && len(filters) > 0 {
		for _, f := range filters {
			f.After(rsp, req)
		}
	}

	if rsp.writeCode {
		rsp.rw.WriteHeader(rsp.code)
	}

	if rsp.writeBytes {
		rsp.rw.Write(rsp.b)
	}

}

func NewHTTPServer(port int) *HTTPServer {
	h := newHandler()
	http.Handle("/", h)
	s := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: http.DefaultServeMux,
	}
	return &HTTPServer{s: s, defaultHandler: h, block: true}
}

type HTTPServer struct {
	s              *http.Server
	defaultHandler *handler
	// handlers       map[string]*handler
	block  bool
	finish func(err error)
}

func (a *HTTPServer) Route(path string, f func(rsp *Response, req *Request) (uint, error)) {
	a.defaultHandler.route(path, f)
}

func (a *HTTPServer) DynamicRoute(pattern string, f func(rsp *Response, req *Request) (uint, error)) {
	a.defaultHandler.dynamicRoute(pattern, f)
}

func (a *HTTPServer) DynamicFilter(pattern string, f Filter) {
	a.defaultHandler.dynamicFilter(pattern, f)
}

func (a *HTTPServer) Filter(path string, f Filter) {
	a.defaultHandler.filter(path, f)
}

func (a *HTTPServer) NotFound(f func(rsp *Response, req *Request) (uint, error)) {
	a.defaultHandler.notFound = f
}

func (a *HTTPServer) ServeHTTP() {
	a.defaultHandler.prepare()
	if !a.block {
		go func() {
			a.doServeHTTP()
		}()
		return
	}
	a.doServeHTTP()
}

func (a *HTTPServer) ServeHTTPS(certFile, keyFile string) {
	a.defaultHandler.prepare()
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
