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
// Date: 2017-08-01 20:02:41
// Last Modified by:   Frank Lee
// Last Modified time: 2018-04-29 22:54
package orm

import (
	"bufio"
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	currentNode         string
	currentUseWhere     = false
	currentWhereIndex   = 0
	nodeRegex           = regexp.MustCompile("^---\\s*\\[\\w*\\]\\s*$")
	commentSegmentRegex = regexp.MustCompile(".*\\s+--.*")
	commentStartRegex   = regexp.MustCompile("\\s+--")
	sqlMap              = make(map[string]gqlEntry)
	sqlQueue            = list.New()
	space               = " "
	q                   = "?"
	sqlIDMap            = make(map[string]int)
)

func parse() {
	// dir := filepath.Dir("../sql/") + "/"
	dir := filepath.Dir(strings.Replace(os.Args[0], "\\", "/", -1)) + "/sql/"
	fl, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("read directory error: %s", err)
	}
	for _, fi := range fl {
		if strings.HasSuffix(fi.Name(), ".sql") {
			readFile(dir + fi.Name())
			currentNode = ""
		}
	}
}

func readFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); (line != "") && (!isCommentLine(line)) {
			line = processCommentSegment(line)
			if whereRegexp.MatchString(line) {
				currentUseWhere = true
			}
			if isNode(line) {
				if currentNode != "" {
					flush()
					currentUseWhere = false
					currentWhereIndex = 0
				}
				getSQLID(line)
			} else {
				pushSQLSegment(line)
				currentWhereIndex++
			}
		}
	}
	flush()
}

// isCommentLine comment line starts with --. Line startswith --- is not comment.
func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "--") && !strings.HasPrefix(line, "---")
}

// processCommentSegment clean comment segment
func processCommentSegment(line string) string {
	if commentSegmentRegex.MatchString(line) && !strings.HasPrefix(line, "---") {
		i := commentStartRegex.FindIndex([]byte(line))[0]
		line = line[0:i]
	}
	return line
}

// isNode is this line a sql id
func isNode(line string) bool {
	return nodeRegex.MatchString(line)
}

// flush flush current gqlEntry
func flush() {
	entry := newGqlEntry(currentNode)
	if sqlQueue.Len() > 0 {
		depth := 0
		treeDepArr := make([]int, 3, 3) // 保存每层handler的个数，当depth++后，treeDepArr[depth]++。当depth == 0时重新初始化
		counterHandler := false         // if counterHandler == false, use paramHandler
		for e := sqlQueue.Front(); e != nil; e = e.Next() {
			tmp := fmt.Sprintf("%s", e.Value)
			if b, t := match(tmp); b {
				counterHandler = true
				depth++
				if depth > len(treeDepArr) {
					treeDepArr = append(treeDepArr, 0)
				}
				treeDepArr[depth-1]++

				var h handler
				if t == ifconst {
					h = &ifHandler{s: tmp, id: currentNode}
				} else if t == ifnotnilconst {
					h = &ifNotNilHandler{s: tmp, id: currentNode}
				} else if t == rangeconst {
					ss := strings.Split(expression(tmp), ",")
					if len(ss) == 4 {
						h = &rangeHandler{s: tmp, id: currentNode, val: strings.TrimSpace(ss[0]), tmp: strings.TrimSpace(ss[1]), left: strings.TrimSpace(ss[2]), right: strings.TrimSpace(ss[3]), num: 4}
					} else {
						h = &rangeHandler{s: tmp, id: currentNode, val: strings.TrimSpace(ss[0]), tmp: strings.TrimSpace(ss[1]), left: "", right: "", num: 2}
					}
				} else if t == setconst {
					h = &setHandler{id: currentNode}
				} else if t == whereconst {
					h = &whereHandler{id: currentNode}
				}

				if depth > 1 { // 当前层大于1
					entry.pushHandler(h, depth-1-1, treeDepArr[depth-1-1]-1)
				} else { // 当是第一层时，newHandlerTree
					entry.newHandler(h)
				}
			} else if end(tmp) {
				depth--
				if depth == 0 { // end
					entry.commitHandlerTree()
					counterHandler = false
					treeDepArr = make([]int, 3, 3)
				}
			} else {
				h := &paramHandler{s: tmp, id: currentNode}
				if !counterHandler {
					entry.newHandler(h)
					entry.commitHandlerTree()
				} else {
					entry.pushHandler(h, depth-1, treeDepArr[depth-1]-1)
					treeDepArr[depth]++
				}
			}
		}
		// re-initialize
		sqlQueue = sqlQueue.Init()
	}
	sqlMap[currentNode] = entry
}

// pushSQLSegment push sql segment into queue
func pushSQLSegment(line string) {
	sqlQueue.PushBack(line)
}

// getSQLID get sql id from node
func getSQLID(node string) {
	start := strings.Index(node, "[")
	end := strings.LastIndex(node, "]")
	sqlID := node[start+1 : end]
	if _, contains := sqlIDMap[sqlID]; contains {
		err := DunplicateSQLIDErr{sqlID}
		panic(err.Error())
	} else {
		sqlIDMap[sqlID] = 0
	}
	currentNode = strings.TrimSpace(node[start+1 : end])
}

func getEntry(key string) (gqlEntry, error) {
	if e, contains := sqlMap[key]; contains {
		// 返回深拷贝的值，否则在并发情况下，会出现协程安全问题
		return e, nil
	}
	ormLog.Error("no match sql with: %s", key)
	return gqlEntry{}, &SQLMatchErr{key}
}
