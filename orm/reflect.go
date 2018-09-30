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

// Package orm orm
// Author: Frank Lee
// Date: 2017-08-01 20:02:45
// Last Modified by:   Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

import (
	"reflect"
	"strings"
	"time"
)

// slice must be a slice or a struct
func build(columns []string, p []interface{}, slice interface{}) {
	m := make(map[string]interface{})
	for i := 0; i < len(columns); i++ {
		m[columns[i]] = p[i]
	}
	val := reflect.ValueOf(slice) // pointer to T
	ind := reflect.Indirect(val)  // T
	var v reflect.Value
	kind := ind.Kind()
	if reflect.Slice == kind {
		v = reflect.New(ind.Type().Elem())
	} else if reflect.Struct == kind {
		v = reflect.New(ind.Type())
	}
	_v := reflect.TypeOf(v.Interface())
	num := _v.Elem().NumField()
	for index := 0; index < num; index++ {
		column := strings.TrimSpace(_v.Elem().Field(index).Tag.Get("column"))
		cv := m[column]
		if nil == cv { // if not selected
			continue
		}
		_type := v.Elem().Field(index).Type().Name()
		field := v.Elem().Field(index)
		if field.CanSet() {
			switch _type {
			case "string":
				_cv := reflect.ValueOf(cv)
				_b := make([]byte, 0, _cv.Len())
				for i := 0; i < _cv.Len(); i++ {
					_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
				}
				field.SetString(string(_b))
			case "int":
				field.SetInt(int64(cv.(int64)))
			case "int8":
				field.SetInt(int64(cv.(int8)))
			case "int16":
				field.SetInt(int64(cv.(int16)))
			case "int32":
				if _val32, ok := cv.(int32); ok {
					field.SetInt(int64(_val32))
				} else if _val64, ok := cv.(int64); ok {
					field.SetInt(int64(_val64))
				}
			case "int64":
				if _val32, ok := cv.(int32); ok {
					field.SetInt(int64(_val32))
				} else if _val64, ok := cv.(int64); ok {
					field.SetInt(int64(_val64))
				}
			case "float32":
				if _val32, ok := cv.(float32); ok {
					field.SetFloat(float64(_val32))
				} else if _val64, ok := cv.(float64); ok {
					field.SetFloat(float64(_val64))
				}
			case "float64":
				if _val32, ok := cv.(float32); ok {
					field.SetFloat(float64(_val32))
				} else if _val64, ok := cv.(float64); ok {
					field.SetFloat(float64(_val64))
				}
			case "uint":
				field.SetUint(uint64(cv.(uint)))
			case "uint8":
				field.SetUint(uint64(cv.(uint8)))
			case "uint16":
				field.SetUint(uint64(cv.(uint16)))
			case "uint32":
				if _val32, ok := cv.(uint32); ok {
					field.SetUint(uint64(_val32))
				} else if _val64, ok := cv.(uint64); ok {
					field.SetUint(uint64(_val64))
				}
			case "uint64":
				if _val32, ok := cv.(uint32); ok {
					field.SetUint(uint64(_val32))
				} else if _val64, ok := cv.(uint64); ok {
					field.SetUint(uint64(_val64))
				}
			case "Time":
				_cv := reflect.ValueOf(cv)
				_b := make([]byte, 0, _cv.Len())
				for i := 0; i < _cv.Len(); i++ {
					_b = append(_b, byte(_cv.Index(i).Interface().(uint8)))
				}
				ts := string(_b)
				t, _ := time.ParseInLocation("2006-01-02 15:04:05", ts, time.Local)
				field.Set(reflect.ValueOf(t))
			}
		}
	}
	if reflect.Slice == kind {
		val = reflect.Append(reflect.Indirect(val), v.Elem())
		ind.Set(val)
	} else if reflect.Struct == kind {
		ind.Set(v.Elem())
	}
}
