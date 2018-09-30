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
// Date: 2017-08-01 19:53:05
// Last Modified by: Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

import (
	"database/sql"
	"time"

	"github.com/FrankLeeC/Aurora/log"
)

var (
	datasource = make(map[string]string)

	dbMap map[string]*sql.DB

	maxOpenConn = 10 // max opened connections

	maxIdleConn = 3 // max opened connections

	maxLifeTime = 5 // maximum amount of time a connection may be reuse. If maxLifeTime <= 0, connections are reused forever.

	ormLog *log.Logger
)

func init() {
	ormLog = log.NewLogger("./AURORA_ORM_LOG/orm.log", nil)
	parse()
	dbMap = make(map[string]*sql.DB)
	maxOpenConn = 10
	maxIdleConn = 3
	maxLifeTime = 10
}

// RegisterDataSource register data source
// Option is optional
// name name of data source
// s uri
func RegisterDataSource(name, s string, o *Option) error {
	if _, contains := datasource[name]; contains {
		return &DunplicateDataSourceErr{dataSource: name}
	}
	db, err := getDatabase(s)
	if err != nil {
		ormLog.Info("register mysql datasource: %s, error: %s", name, err.Error())
		// fmt.Println("register mysql datasource", name, "error:", err)
		return err
	}
	err = db.Ping()
	if err != nil {
		ormLog.Info("register mysql datasource: %s, error: %s", name, err.Error())
		// fmt.Println("register mysql datasource", name, "error:", err)
		return err
	}
	if o != nil {
		db.SetMaxOpenConns(o.MaxIdleConnection)
		db.SetMaxIdleConns(o.MaxIdleConnection)
		db.SetConnMaxLifetime(time.Duration(o.MaxLifeTime) * time.Minute)
	} else {
		db.SetMaxOpenConns(maxOpenConn)
		db.SetMaxIdleConns(maxIdleConn)
		db.SetConnMaxLifetime(time.Duration(maxLifeTime) * time.Minute)
	}
	datasource[name] = s
	dbMap[name] = db
	return nil
}

// sql.DB is a pool
func getDatabase(s string) (*sql.DB, error) {
	return sql.Open("mysql", s)
}

func getConn(s string) (*sql.DB, error) {
	if v, contains := dbMap[s]; contains {
		return v, nil
	}
	return nil, &UnknownDataSourceErr{dataSource: s}
}
