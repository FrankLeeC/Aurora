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
// Date: 2017-08-01 19:53:21
// Last Modified by:   Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

// SQLMatchErr no sql matches sqlID
type SQLMatchErr struct {
	sqlID string
}

func (err *SQLMatchErr) Error() string {
	return `sqlID: "` + err.sqlID + `" no sql named "` + err.sqlID + `"`
}

// UnknownDataSourceErr unknown data source
type UnknownDataSourceErr struct {
	dataSource string
}

func (err *UnknownDataSourceErr) Error() string {
	return "unkown data source:" + err.dataSource
}

// DunplicateDataSourceErr dunplicate data source
type DunplicateDataSourceErr struct {
	dataSource string
}

func (err *DunplicateDataSourceErr) Error() string {
	return "duplicated data source:" + err.dataSource
}

// DunplicateSQLIDErr dunplicate sql id
type DunplicateSQLIDErr struct {
	sqlID string
}

func (err *DunplicateSQLIDErr) Error() string {
	return `sqlID: "` + err.sqlID + `" dunplicated sql id`
}

// PointerParamErr ResultSet receiver must be a pointer
type PointerParamErr struct {
	sqlID string
}

func (err *PointerParamErr) Error() string {
	return `sqlID: "` + err.sqlID + `" resultset receiver must be a pointer`
}

// NotSliceInRangeErr empty slice in range
type NotSliceInRangeErr struct {
	sqlID string
}

func (err *NotSliceInRangeErr) Error() string {
	return `sqlID: "` + err.sqlID + `" params in range must be a slice or array`
}

// NoFieldNeedUpdateErr no field need to be update in update set clause
type NoFieldNeedUpdateErr struct {
	sqlID string
}

func (err *NoFieldNeedUpdateErr) Error() string {
	return `sqlID: "` + err.sqlID + `" no field need to be updated`
}

// ParamNotFoundErr params not found in M
type ParamNotFoundErr struct {
	param string
	sqlID string
}

func (err *ParamNotFoundErr) Error() string {
	return `sqlID: "` + err.sqlID + `" param "` + err.param + `" not found`
}

// NotStructErr not a struct
type NotStructErr struct {
	param string
	sqlID string
}

func (err *NotStructErr) Error() string {
	return `sqlID: "` + err.sqlID + `" param "` + err.param + `" is not a struct`
}
