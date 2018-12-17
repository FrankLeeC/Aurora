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

// Package http http tools
// Author: Frank Lee
// Date: 2017-11-17 09:16:30
// Last Modified by:   Frank Lee
// Last Modified time: 2017-11-17 09:16:30
package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// ContentType content-type
type ContentType int

const (

	// JSON application/json
	JSON ContentType = 0

	// FORM application/x-www-form-urlencoded
	FORM ContentType = 1
)

// Post post
// param url: url
// param params: params
// param t: content-type http.JSON/http.FORM
// return []byte: http response body
// return error: errors
func Post(url, params string, t ContentType) ([]byte, error) {
	c := getContentType(t)
	rsp, err := http.Post(url, c, strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// Get get
// param url: url
// return []byte: http response body
// return error: errors
func Get(url string) ([]byte, error) {
	rsp, err := http.Get(url)
	if err != nil {
		return nil, err

	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// GetInTime get timeout
// param url: url
// param time: timeout
// return []byte: http response body
// return error: errors
func GetInTime(url string, time *time.Duration) ([]byte, error) {
	if time == nil {
		return Get(url)
	}
	clt := getClt(false, time)
	rsp, err := clt.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// PostInTime post timeout
// param url: url
// param params: params
// param t content-type http.JSON/http.FORM
// param time timeout
// return []byte: http response body
// return error: errors
func PostInTime(url, params string, t ContentType, time *time.Duration) ([]byte, error) {
	if time == nil {
		return Post(url, params, t)
	}
	c := getContentType(t)
	clt := getClt(false, time)
	rsp, err := clt.Post(url, c, strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// PostInsecure https post insecurely with timeout
// param url: url
// param params: params
// param t content-type http.JSON/http.FORM
// return []byte: http response body
// return error: errors
func PostInsecure(url, params string, t ContentType) ([]byte, error) {
	return PostInsecureInTime(url, params, t, nil)
}

// PostInsecureInTime https post insecurely with timeout
// param url: url
// param params: params
// param t content-type http.JSON/http.FORM
// param time: timeout
// return []byte: http response body
// return error: errors
func PostInsecureInTime(url, params string, t ContentType, time *time.Duration) ([]byte, error) {
	s := strings.TrimSpace(url)
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "http://") {
		return PostInTime(url, params, t, time)
	}
	c := getContentType(t)
	clt := getClt(true, time)
	rsp, err := clt.Post(url, c, strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

// GetInsecure https get insecurely with timeout
// param url: url
// return []byte: response body
// return error: errors
func GetInsecure(url string) ([]byte, error) {
	return GetInsecureInTime(url, nil)
}

// GetInsecureInTime https get insecurely with timeout
// param url: url
// param time: timeout
// return []byte: response body
// return error: errors
func GetInsecureInTime(url string, time *time.Duration) ([]byte, error) {
	s := strings.TrimSpace(url)
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "http://") {
		return GetInTime(url, time)
	}
	clt := getClt(true, time)
	rsp, err := clt.Get(url)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return ioutil.ReadAll(rsp.Body)
}

func getClt(insecure bool, time *time.Duration) *http.Client {
	clt := &http.Client{}
	if time != nil {
		clt.Timeout = *time
	}
	if insecure {
		p := x509.NewCertPool()
		tp := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            p,
				InsecureSkipVerify: true,
			},
		}
		clt.Transport = tp
	}
	return clt
}

func getContentType(t ContentType) string {
	var c string
	switch t {
	case JSON:
		c = "application/json;charset=utf-8"
	case FORM:
		c = "application/x-www-form-urlencoded"
	}
	return c
}
