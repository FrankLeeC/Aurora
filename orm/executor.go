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
// Date: 2017-08-01 20:02:22
// Last Modified by: Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

import (
	"database/sql"
)

var (
	url      string
	user     string
	password string
)

type executor struct {
	db         *sql.DB
	sql        string
	useTrans   bool
	dataSource string
	params     []interface{}
	returnID   bool
	execType   string // insert update delete query
}

func initExecutor(db *sql.DB, dataSource string) *executor {
	exec := new(executor)
	exec.db = db
	exec.dataSource = dataSource
	exec.returnID = false
	return exec
}

func (exec *executor) executeQuery(sql string, params []interface{}, a interface{}) (int64, error) {
	exec.sql = sql
	exec.params = params
	return exec.query(a, true, nil)
}

func (exec *executor) executeQueryMany(sql string, params []interface{}, a interface{}) (int64, error) {
	exec.sql = sql
	exec.params = params
	return exec.query(nil, false, a)
}

func (exec *executor) query(i interface{}, single bool, slice interface{}) (int64, error) {
	var count int64
	st, err := exec.db.Prepare(exec.sql)
	if st != nil {
		defer st.Close()
	}
	if err != nil {
		ormLog.Error("prepare error: %s", err.Error())
		// fmt.Println("prepare error:", err)
		return 0, err
	}
	var row *sql.Rows
	if len(exec.params) > 0 {
		row, err = st.Query(exec.params...)
	} else {
		row, err = st.Query()
	}
	if row != nil {
		defer row.Close()
	}
	if err != nil {
		ormLog.Error("query error: %s", err.Error())
		// fmt.Println("query error:", err)
		return 0, err
	}
	if single {
		if row.Next() {
			count = 1
			columns, err := row.Columns()
			if err != nil {
				return 0, err
			}
			_res := make([]interface{}, len(columns))
			_v := make([]interface{}, len(columns))
			for i := range _v {
				_res[i] = &_v[i]
			}
			row.Scan(_res...)
			build(columns, _v, i)
		}
	} else {
		for row.Next() {
			count++
			columns, err := row.Columns()
			if err != nil {
				return 0, err
			}
			_res := make([]interface{}, len(columns))
			_v := make([]interface{}, len(columns))
			for i := range _v {
				_res[i] = &_v[i]
			}
			row.Scan(_res...)
			build(columns, _v, slice)
		}
	}
	return count, nil
}

func (exec *executor) insertUpdateDelete() (int64, error) {
	st, err := exec.db.Prepare(exec.sql)
	if st != nil {
		defer st.Close()
	}
	if err != nil {
		ormLog.Error("error: %s", err.Error())
		// fmt.Println("error:", err)
		return 0, err
	}
	// fmt.Println("-----------------------------------------in executor", exec.params)
	re, err := st.Exec(exec.params...)
	if err != nil {
		ormLog.Error("error: %s", err.Error())
		// fmt.Println("error:", err)
		return 0, err
	}
	if exec.execType == "insert" && exec.returnID {
		return re.LastInsertId()
	}
	return re.RowsAffected()
}
