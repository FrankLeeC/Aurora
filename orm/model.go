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
// Date: 2017-06-01 20:02:27
// Last Modified by:   Frank Lee
// Last Modified time: 2018-04-09 22:28:27
package orm

import (
	"bytes"
	"container/list"
	"fmt"
	"strings"
	"time"
	// "fmt"

	"reflect"
)

// M map of paramaters. Use it if parameter is not a struct
type M map[string]interface{}

// AsM cast struct to M using params
func AsM(a interface{}, params []string) M {
	m := M{}
	rv := reflect.ValueOf(a)
	for _, v := range params {
		m[v] = rv.FieldByName(v).Interface()
	}
	return m
}

// gql gql
type gql struct {
	tableNames map[string]string
	sqlID      string
	sql        string // executable sql
	m          M
	params     []interface{}
	useTran    bool
	exec       *executor
	returnID   bool
	showSQL    bool
	dataSource string
	execType   string
	e          error
}

// InitGQL return a pointer to new gql
func InitGQL() *gql {
	g := new(gql)
	g.tableNames = make(map[string]string)
	g.returnID = false
	g.dataSource = "default"
	return g
}

// ReturnLastID return last id
func (g *gql) ReturnLastID() *gql {
	g.returnID = true
	g.exec.returnID = true
	return g
}

// ShowSQL print sql and params
func (g *gql) ShowSQL(b bool) *gql {
	g.showSQL = b
	return g
}

// M pass a M
func (g *gql) M(m M) *gql {
	g.m = m
	return g
}

// Tables pass table names
func (g *gql) Tables(t map[string]string) *gql {
	g.tableNames = t
	return g
}

// Insert insert into mysql
func (g *gql) Insert() *gql {
	g.execType = "insert"
	return g
}

// Update update
func (g *gql) Update() *gql {
	g.execType = "update"
	return g
}

// Delete delete
func (g *gql) Delete() *gql {
	g.execType = "delete"
	return g
}

// One execute gql and set the first result into a
func (g *gql) One(a interface{}) (int64, error) {
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return 0, &PointerParamErr{}
	}
	g.execType = "query"
	db, err := getConn(g.dataSource)
	if err != nil {
		ormLog.Error("connect to database error: %s", err.Error())
		return 0, err
	}
	exec := initExecutor(db, g.dataSource)
	g.exec = exec
	err = g.build()
	if err != nil {
		ormLog.Error(err.Error())
		return 0, err
	}
	g.exec.execType = g.execType
	s := time.Now()
	count, err := g.executeQuery(a)
	e := time.Now()
	if SHOWSQL || g.showSQL {
		ormLog.Info("sqlID: %s, sql: %s", g.sqlID, g.sql)
		ormLog.Info("sqlID: %s, params: %s", g.sqlID, g.params)
		ormLog.Info("sqlID: %s, execute time: %s", g.sqlID, e.Sub(s))
	}
	return count, err
}

// All execute gql and set all results into a
func (g *gql) All(a interface{}) (int64, error) {
	if reflect.ValueOf(a).Kind() != reflect.Ptr {
		return 0, &PointerParamErr{}
	}
	g.execType = "query"
	db, err := getConn(g.dataSource)
	if err != nil {
		ormLog.Error("connect to database error: %s", err.Error())
		return 0, err
	}
	exec := initExecutor(db, g.dataSource)
	g.exec = exec
	err = g.build()
	if err != nil {
		ormLog.Error(err.Error())
		return 0, err
	}
	g.exec.execType = g.execType
	s := time.Now()
	count, err := g.executeQueryMany(a)
	e := time.Now()
	if SHOWSQL || g.showSQL {
		ormLog.Info("sqlID: %s, sql: %s", g.sqlID, g.sql)
		ormLog.Info("sqlID: %s, params: %s", g.sqlID, g.params)
		ormLog.Info("sqlID: %s, execute time: %s", g.sqlID, e.Sub(s))
	}
	return count, err
}

// Result delete, update, insert
func (g *gql) Result() (int64, error) {
	// fmt.Println("===========================in model", gql.sql)
	db, err := getConn(g.dataSource)
	if err != nil {
		ormLog.Error("connect to database error: %s", err.Error())
		return 0, err
	}
	exec := initExecutor(db, g.dataSource)
	g.exec = exec
	err = g.build()
	if err != nil {
		ormLog.Error(err.Error())
		return 0, err
	}
	g.exec.sql = g.sql
	g.exec.params = g.params
	g.exec.execType = g.execType
	s := time.Now()
	c, err := g.exec.insertUpdateDelete()
	e := time.Now()
	if SHOWSQL || g.showSQL {
		ormLog.Info("sqlID: %s, sql: %s", g.sqlID, g.sql)
		ormLog.Info("sqlID: %s, params: %s", g.sqlID, g.params)
		ormLog.Info("sqlID: %s, execute time: %s", g.sqlID, e.Sub(s))
	}
	return c, err
}

// Use set sql id
func (g *gql) Use(sqlID string) *gql {
	g.sqlID = sqlID
	return g
}

// UseDataSource switch datasource
func (g *gql) UseDataSource(dataSource string) *gql {
	g.dataSource = dataSource
	return g
}

func (g *gql) executeQuery(a interface{}) (int64, error) {
	return g.exec.executeQuery(g.sql, g.params, a)
}

func (g *gql) executeQueryMany(a interface{}) (int64, error) {
	return g.exec.executeQueryMany(g.sql, g.params, a)
}

func (g *gql) build() error {
	entry, err := getEntry(g.sqlID)
	if err != nil {
		return err
	}
	// entry.replaceTableName(g.tableNames)
	g.sql, g.params, g.e = entry.build(g.m, g.tableNames)
	return g.e
}

// SQL sql struct
type builder struct {
	sql             string
	raw             string
	rangeParamCount map[string]int
	useWhere        bool
	setWhere        bool
	params          []interface{}
	paramMap        map[string]interface{}
	processOver     bool
}

func initBuilder() *builder {
	b := new(builder)
	b.setWhere = false
	b.paramMap = make(map[string]interface{})
	b.params = make([]interface{}, 0)
	b.rangeParamCount = make(map[string]int)
	b.processOver = false
	return b
}

// Option option
type Option struct {
	MaxOpenedConnection int // max open conn
	MaxIdleConnection   int
	MaxLifeTime         int // minutes
}

type handler interface {
	handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error)
	len() int // length of children
	children() []handler
	addHandler(h handler)
	name() string
}

type ifHandler struct {
	id  string
	s   string
	val string
	l   []handler
}

func (a *ifHandler) judge(m M, kvs []kv) bool {
	s, f := meetCondition(a.s, m, kvs, ifconst)
	if f {
		a.val = s[0]
	}
	return f
}

func (a *ifHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	if a.judge(m, kvs) {
		if a.l != nil {
			var buf bytes.Buffer
			params := make([]interface{}, 0)
			for i := range a.l {
				v1, v2, err := a.l[i].handle(m, tm, kvs)
				if err != nil {
					return "", nil, err
				}
				buf.WriteString(v1)
				if v2 != nil {
					params = append(params, v2...)
				}
			}
			return buf.String(), params, nil
		}
		return "", nil, nil
	}
	return "", nil, nil
}

func (a *ifHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *ifHandler) len() int {
	return len(a.l)
}

func (a *ifHandler) children() []handler {
	return a.l
}

func (a *ifHandler) name() string {
	return "ifHandler" + a.s
}

type ifNotNilHandler struct {
	id string
	s  string
	l  []handler
}

func (a *ifNotNilHandler) judge(m M, kvs []kv) bool {
	_, f := meetCondition(a.s, m, kvs, ifnotnilconst)
	// fmt.Println(a.s, f)
	return f
}

func (a *ifNotNilHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	if a.judge(m, kvs) {
		if a.l != nil {
			var buf bytes.Buffer
			params := make([]interface{}, 0)
			for i := range a.l {
				v1, v2, err := a.l[i].handle(m, tm, kvs)
				if err != nil {
					return "", nil, err
				}
				buf.WriteString(v1)
				if v2 != nil {
					params = append(params, v2...)
				}
			}
			return buf.String(), params, nil
		}
		return "", nil, nil
	}
	return "", nil, nil
}

func (a *ifNotNilHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *ifNotNilHandler) len() int {
	return len(a.l)
}

func (a *ifNotNilHandler) children() []handler {
	return a.l
}

func (a *ifNotNilHandler) name() string {
	return "ifNotNilHandler" + a.s
}

type setHandler struct {
	id string
	l  []handler
}

func (a *setHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	if a.l != nil {
		var buf bytes.Buffer
		tmpArr := make([]string, 0)
		params := make([]interface{}, 0)
		for i := range a.l {
			v1, v2, err := a.l[i].handle(m, tm, kvs)
			if err != nil {
				return "", nil, err
			}
			v1 = strings.TrimSpace(v1)
			if strings.HasSuffix(v1, ",") {
				v1 = v1[0 : len(v1)-1]
			}
			tmpArr = append(tmpArr, v1)
			if v2 != nil {
				params = append(params, v2...)
			}
		}
		realArr := make([]string, 0)
		for i := range tmpArr {
			if "" != strings.TrimSpace(tmpArr[i]) {
				realArr = append(realArr, tmpArr[i])
			}
		}
		buf.WriteString(strings.Join(realArr, ","+space))
		if strings.TrimSpace(buf.String()) == "" {
			return "", nil, &NoFieldNeedUpdateErr{a.id}
		}
		return space + "SET" + space + buf.String(), params, nil
	}
	return "", nil, &NoFieldNeedUpdateErr{a.id}
}

func (a *setHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *setHandler) len() int {
	return len(a.l)
}

func (a *setHandler) children() []handler {
	return a.l
}

func (a *setHandler) name() string {
	return "setHandler"
}

type whereHandler struct {
	id string
	l  []handler
}

func (a *whereHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	// var buf bytes.Buffer
	tmpArr := make([]string, 0)
	params := make([]interface{}, 0)
	tmpHArr := make([]string, 0)
	for i := range a.l {
		v1, v2, err := a.l[i].handle(m, tm, kvs)
		if err != nil {
			return "", nil, err
		}
		v1 = strings.TrimSpace(v1)
		// if andRegexp.MatchString(v1) {
		// 	v1 = v1[4:]
		// }
		tmpArr = append(tmpArr, v1)
		tmpHArr = append(tmpHArr, a.l[i].name())
		if v2 != nil {
			params = append(params, v2...)
		}
	}
	realArr := make([]string, 0)
	realHArr := make([]string, 0)
	for i := range tmpArr {
		if strings.TrimSpace(tmpArr[i]) != "" {
			realArr = append(realArr, tmpArr[i])
			realHArr = append(realHArr, tmpHArr[i])
		}
	}
	if len(realArr) <= 0 {
		return "", nil, nil
	}
	var realBuf bytes.Buffer
	if andRegexp.MatchString(realArr[0]) {
		realArr[0] = realArr[0][4:]
	} else if orRegexp.MatchString(realArr[0]) {
		realArr[0] = realArr[0][3:]
	}
	realBuf.WriteString(space + realArr[0] + space)
	for k := 1; k < len(realArr); k++ {
		// if realHArr[k] != "rangeHandler" {
		// 	// realBuf.WriteString(space + "AND" + space + realArr[k])
		// 	realBuf.WriteString(space + realArr[k])
		// } else {
		// 	realBuf.WriteString(space + realArr[k])
		// }
		realBuf.WriteString(space + realArr[k])
	}
	// buf.WriteString(strings.Join(realArr, space+"AND"+space))
	if strings.TrimSpace(realBuf.String()) == "" {
		return "", nil, nil
	}
	return space + "WHERE" + space + realBuf.String(), params, nil
}

func (a *whereHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *whereHandler) len() int {
	return len(a.l)
}

func (a *whereHandler) children() []handler {
	return a.l
}

func (a *whereHandler) name() string {
	return "whereHandler"
}

type rangeHandler struct {
	id    string
	s     string
	val   string // 1st
	tmp   string // 2nd
	num   int    // number of params between [ and ]
	left  string // 3rd
	right string // 4th
	l     []handler
}

func (a *rangeHandler) judge(m M, kvs []kv) bool {
	_, f := meetCondition(a.s, m, kvs, rangeconst)
	return f
}

func (a *rangeHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	if a.judge(m, kvs) {
		var v interface{}
		f := false
		if kvs != nil {
			for i := len(kvs) - 1; i >= 0; i-- { // 栈的方式遍历，才能满足作用域的要求。
				if kvs[i].k == a.val {
					v = kvs[i].v
					f = true
					break
				}
			}
		}
		if !f {
			if t, c := m[a.val]; c {
				v = t
			}
		}
		if v != nil {
			slice := reflect.ValueOf(v)
			if slice.Kind() != reflect.Slice && slice.Kind() != reflect.Array {
				return "", nil, &NotSliceInRangeErr{a.id}
			}
			var buf bytes.Buffer
			tmpArr := make([]string, 0)
			params := make([]interface{}, 0)
			for i := 0; i < slice.Len(); i++ {
				tmp := slice.Index(i).Interface()
				kvs = append(kvs, kv{a.tmp, tmp})
				var innerBuf bytes.Buffer
				for j := range a.l {
					v1, v2, err := a.l[j].handle(m, tm, kvs)
					if err != nil {
						return "", nil, err
					}
					innerBuf.WriteString(v1)
					if v2 != nil {
						params = append(params, v2...)
					}
				}
				tmpArr = append(tmpArr, innerBuf.String())
			}
			buf.WriteString(a.left) // if a.num == 2, a.left is ""
			buf.WriteString(strings.Join(tmpArr, ","))
			buf.WriteString(a.right) // if a.num == 2, a.right is ""
			return buf.String(), params, nil
		}
		return "", nil, &ParamNotFoundErr{param: a.val, sqlID: a.id}
	}
	return "", nil, nil
}

func (a *rangeHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *rangeHandler) len() int {
	return len(a.l)
}

func (a *rangeHandler) children() []handler {
	return a.l
}

func (a *rangeHandler) name() string {
	return "rangeHandler"
}

type paramHandler struct {
	id string
	s  string
	l  []handler
}

func (a *paramHandler) handle(m M, tm map[string]string, kvs []kv) (string, []interface{}, error) {
	params := make([]interface{}, 0)
	sql := a.replaceTableName(tm)
	for paramRegexp.MatchString(sql) {
		tmp := strings.TrimSpace(sql)
		i := strings.Index(tmp, "#")
		before := space
		after := space
		if i > 0 {
			before += tmp[0:i]
		}
		tmp = tmp[i+1:]
		i = strings.Index(tmp, "#")
		param := strings.TrimSpace(tmp[0:i])
		if i < len(tmp)-1 {
			after += tmp[i+1:]
		}
		if strings.Contains(param, ".") { // struct
			pf := strings.Split(param, ".")
			v, f := getValue(pf[0], m, kvs)
			if !f {
				return "", nil, &ParamNotFoundErr{param: pf[0], sqlID: a.id}
			}
			if reflect.TypeOf(v).Kind() != reflect.Struct {
				return "", nil, &NotStructErr{param: pf[0], sqlID: a.id}
			}
			params = append(params, reflect.ValueOf(v).FieldByName(pf[1]).Interface())
		} else {
			v, f := getValue(param, m, kvs)
			if !f {
				return "", nil, &ParamNotFoundErr{param: param, sqlID: a.id}
			}
			params = append(params, v)
		}
		sql = before + space + q + space + after
	}
	return sql, params, nil
}

func (a *paramHandler) addHandler(h handler) {
	if a.l == nil {
		a.l = make([]handler, 0)
	}
	a.l = append(a.l, h)
}

func (a *paramHandler) len() int {
	return len(a.l)
}

func (a *paramHandler) children() []handler {
	return a.l
}

func (a *paramHandler) name() string {
	return "paramHandler"
}

func (a *paramHandler) replaceTableName(m map[string]string) string {
	sql := a.s
	for tableNameRegexp.MatchString(sql) {
		loc := tableNameRegexp.FindIndex([]byte(sql))
		placeholder := sql[loc[0]+1 : loc[1]-1]
		tableName := m[placeholder[6:]]
		sql = sql[0:loc[0]] + space + tableName + space + sql[loc[1]:]
	}
	return sql
}

type gqlEntry struct {
	id string
	hs []handler // 针对每一个树，执行深度优先遍历，生成sql
	i  int       // current handler index
	e  error
}

func newGqlEntry(id string) gqlEntry {
	return gqlEntry{
		id: id,
		hs: make([]handler, 0),
		i:  0,
		e:  nil,
	}
}

func (a *gqlEntry) getHandler(l, i int) handler {
	queue := list.New()
	queue.PushBack(a.hs[a.i])
	layer := 0
	var p handler
	for queue.Len() > 0 {
		if layer == l {
			for j := 0; j < i; j++ {
				tmp := queue.Front()
				queue.Remove(tmp)
			}
			p = queue.Front().Value.(handler)
			goto out
		}
		tmp := make([]handler, 0)
		for e := queue.Front(); e != nil; e = e.Next() {
			if e.Value.(handler).len() > 0 {
				tmp = append(tmp, e.Value.(handler).children()...)
			}
		}
		queue = queue.Init()
		for i := range tmp {
			queue.PushBack(tmp[i])
		}
		layer++
	}
out:
	queue = queue.Init()
	queue = nil // free memory
	return p
}

// l 层数（0开始）
// i 第l层第i（0开始）个
func (a *gqlEntry) pushHandler(h handler, l, i int) {
	p := a.getHandler(l, i)
	p.addHandler(h)
}

func (a *gqlEntry) newHandler(h handler) {
	a.hs = append(a.hs, h)
}

func (a *gqlEntry) commitHandlerTree() {
	a.i++
}

func (a *gqlEntry) print() {
	for i := range a.hs {
		fmt.Println(i, a.hs[i].name())
		for j := 0; j < a.hs[i].len(); j++ {
			fmt.Println(i, j, a.hs[i].children()[j].name())
			for k := 0; k < a.hs[i].children()[j].len(); k++ {
				fmt.Println(i, j, k, a.hs[i].children()[j].children()[k].name())
			}
		}
	}
}

func (a *gqlEntry) build(m M, tm map[string]string) (string, []interface{}, error) {
	// a.print()
	var sql string
	params := make([]interface{}, 0)
	tmpArr := make([]string, 0)
	for i := range a.hs {
		kvs := make([]kv, 0)
		tmp, p, err := a.hs[i].handle(m, tm, kvs)
		if err != nil {
			a.e = err
			sql = ""
			params = nil
			goto out
		}
		if tmp != "" {
			tmpArr = append(tmpArr, tmp)
		}
		if p != nil {
			params = append(params, p...)
		}
	}
	sql = strings.Join(tmpArr, space)
out:
	return sql, params, a.e
}

type kv struct {
	k string
	v interface{}
}

func getValue(name string, m M, kvs []kv) (interface{}, bool) {
	var v interface{}
	f := false
	if kvs != nil {
		for i := len(kvs) - 1; i >= 0; i-- {
			if kvs[i].k == name {
				v = kvs[i].v
				f = true
				break
			}
		}
	}
	if !f {
		v, f = m[name]
	}
	return v, f
}
