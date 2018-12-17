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
// Date: 2017-08-01 20:02:48
// Last Modified by:   Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

import (
	"database/sql"
	"reflect"
	"time"
)

// Transaction transaction
type Transaction struct {
	tx         *sql.Tx
	err        error
	exec       *executor
	showSQL    bool
	dataSource string
	f          bool
}

// InitTran init transaction
func InitTran(dataSource string) *Transaction {
	tran := new(Transaction)
	db, err := getConn(dataSource)
	if err != nil {
		ormLog.Error("connect to database error: %s", err.Error())
		return nil
	}
	tran.exec = initExecutor(db, dataSource)
	tran.tx, err = db.Begin()
	if err != nil {
		ormLog.Error("begin transaction error: %s", err.Error())
		// fmt.Println("begin transaction error:", err)
		return nil
	}
	return tran
}

// ShowSQL show sql
func (tran *Transaction) ShowSQL(b bool) *Transaction {
	tran.showSQL = b
	return tran
}

func (tran *Transaction) insertUpdateDelete(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	if tran.err == nil {
		gql := InitGQL().Tables(tableNames).Use(sqlID).M(params)
		err := gql.build()
		if err != nil {
			tran.err = err
			ormLog.Error(err.Error())
			return 0, err
		}
		sql := gql.sql
		p := gql.params
		tran.exec.sql = sql
		tran.exec.params = p
		s := time.Now()
		c, err := tran.exec.insertUpdateDelete()
		e := time.Now()
		tran.err = err
		if SHOWSQL || tran.showSQL {
			ormLog.Info("sqlID: %s, sql: %s", sqlID, sql)
			ormLog.Info("sqlID: %s, params: %s", sqlID, p)
			ormLog.Info("sqlID: %s, execute time: %s", sqlID, e.Sub(s))
		}
		return c, err
	}
	return 0, tran.err
}

// Insert insert
func (tran *Transaction) Insert(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	return tran.insertUpdateDelete(sqlID, params, result, tableNames)
}

// Update update
func (tran *Transaction) Update(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	return tran.insertUpdateDelete(sqlID, params, result, tableNames)
}

// Delete delete
func (tran *Transaction) Delete(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	return tran.insertUpdateDelete(sqlID, params, result, tableNames)
}

// One select one
func (tran *Transaction) One(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	if reflect.ValueOf(result).Kind() != reflect.Ptr {
		return 0, &PointerParamErr{}
	}
	if tran.err == nil {
		gql := InitGQL().Tables(tableNames).Use(sqlID).M(params)
		err := gql.build()
		if err != nil {
			tran.err = err
			ormLog.Error(err.Error())
			return 0, err
		}
		sql := gql.sql
		p := gql.params
		tran.exec.execType = "query"
		s := time.Now()
		count, err := tran.exec.executeQuery(sql, p, result)
		e := time.Now()
		tran.err = err
		if SHOWSQL || tran.showSQL {
			ormLog.Info("sqlID: %s, sql: %s", sqlID, sql)
			ormLog.Info("sqlID: %s, params: %s", sqlID, p)
			ormLog.Info("sqlID: %s, execute time: %s", sqlID, e.Sub(s))
		}
		return count, err
	}
	return 0, tran.err
}

// All select many
func (tran *Transaction) All(sqlID string, params M, result interface{}, tableNames map[string]string) (int64, error) {
	if reflect.ValueOf(result).Kind() != reflect.Ptr {
		return 0, &PointerParamErr{}
	}
	if tran.err == nil {
		gql := InitGQL().Tables(tableNames).Use(sqlID).M(params)
		err := gql.build()
		if err != nil {
			tran.err = err
			ormLog.Error(err.Error())
			return 0, err
		}
		sql := gql.sql
		p := gql.params
		tran.exec.execType = "query"
		s := time.Now()
		count, err := tran.exec.executeQueryMany(sql, p, result)
		e := time.Now()
		tran.err = err
		if SHOWSQL || tran.showSQL {
			ormLog.Info("sqlID: %s, sql: %s", sqlID, sql)
			ormLog.Info("sqlID: %s, params: %s", sqlID, p)
			ormLog.Info("sqlID: %s, execute time: %s", sqlID, e.Sub(s))
		}
		return count, err
	}
	return 0, tran.err
}

// Commit commit transaction
func (tran *Transaction) Commit() error {
	if tran.err != nil {
		e := tran.tx.Rollback()
		if e != nil {
			ormLog.Error("rollback error: %s", e.Error())
		}
		return e
	}
	e := tran.tx.Commit()
	if e != nil {
		ormLog.Error("commit error: %s", e.Error())
	}
	return e
}

// Error return the first error in db transaction
func (tran *Transaction) Error() error {
	return tran.err
}
