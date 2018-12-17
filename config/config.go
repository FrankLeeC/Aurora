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

// Package config config util
// Author: Frank Lee
// Date: 2017-08-07 10:43:26
// Last Modified by:   Frank Lee
// Last Modified time: 2017-08-07 10:43:26
package config

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	runMode    string
	runModeSet bool
	findNode   bool
	properties = make(map[string]interface{})
	global     = make(map[string]interface{})
	keyStack   = list.New()
	stackDepth = 0

	commentSegmentRegex = regexp.MustCompile(".*\\s+#.*")
	commentStartRegex   = regexp.MustCompile("\\s+#")

	nodeRegex = regexp.MustCompile("^\\[+\\w*\\]+$")
	evalRegex = regexp.MustCompile(`(?U)eval\(\d+([+\-*/]\d+)*?\)`)
	exp       = regexp.MustCompile(`[\-+*/]`)
)

// func main() {
//     fmt.Println(properties)
//     fmt.Println("===================")
//     fmt.Println(global)
//     fmt.Println("===================")
//     fmt.Println(GetString("mysql>default>port"))
//     fmt.Println(GetString("a"))
// }

func init() {
	runModeSet = false
	findNode = false
	parseConfigFile()
}

// GetString get string value by key
func GetString(key string) string {
	s, contains := getFromMode(key)
	if !contains {
		s, contains = getFromGlobal(key)
	}
	if contains {
		return s
	}
	fmt.Println("value of key", key, "is not set!")
	return ""
}

// GetInt get int value by key. If key is not set, return 0. strconv.Atoi error, return 0.
func GetInt(key string) int {
	s, contains := getFromMode(key)
	if !contains {
		s, contains = getFromGlobal(key)
	}
	if contains {
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println("GetInt error:", err)
			return 0
		}
		return i
	}
	fmt.Println("value of", key, "is not set!")
	return 0
}

// GetInt64 get int64 value by key. If key is not set, return 0. strconv.ParseInt error, return 0.
func GetInt64(key string) int64 {
	s, contains := getFromMode(key)
	if !contains {
		s, contains = getFromGlobal(key)
	}
	if contains {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			fmt.Println("GetInt64 error:", err)
			return 0
		}
		return i
	}
	fmt.Println("values of", key, "is not set!")
	return 0
}

// GetBool get bool value by key. If key is not set, return false. strconv.ParseBool error, return false.
func GetBool(key string) bool {
	s, contains := getFromMode(key)
	if !contains {
		s, contains = getFromGlobal(key)
	}
	if contains {
		b, err := strconv.ParseBool(s)
		if err != nil {
			fmt.Println("GetBool error:", err)
			return false
		}
		return b
	}
	fmt.Println("value of", key, "is not set!")
	return false
}

// GetEval get eval value by key. If key is not set, return 0. strconv.ParseFloat error, return 0.
func GetEval(key string) float64 {
	e, contains := getEvalFromMode(key)
	if !contains {
		e, contains = getEvalFromGlobal(key)
	}
	if contains {
		return e
	}
	fmt.Println("value of", key, "is not set!")
	return 0
}

// Contains contains key or not
func Contains(key string) bool {
	_, contains := getFromMode(key)
	if !contains {
		_, contains = getFromGlobal(key)
	}
	return contains
}

// GetAll if not recurse, only return k-v starting with exactly key, else return all k-v starting with key.
func GetAll(key string, recurse bool) map[string]interface{} {
	key = runMode + ">" + key
	m := make(map[string]interface{})
	for k, v := range properties {
		if strings.HasPrefix(k, key) {
			if recurse {
				m[k] = v
			} else {
				t := strings.Replace(k, key, "", 1)
				if len(t) >= 2 {
					t = t[1:]
					if !strings.Contains(t, ">") {
						m[k] = v
					}
				}
			}
		}
	}
	return m
}

func getFromMode(key string) (string, bool) {
	if v, contains := properties[runMode+">"+key]; contains {
		return v.(string), true
	}
	return "", false
}

func getFromGlobal(key string) (string, bool) {
	if v, contains := global[key]; contains {
		return v.(string), true
	}
	return "", false
}

func getEvalFromMode(key string) (float64, bool) {
	if v, contains := properties[runMode+">"+key]; contains {
		return v.(float64), true
	}
	return 0, false
}

func getEvalFromGlobal(key string) (float64, bool) {
	if v, contains := global[key]; contains {
		return v.(float64), true
	}
	return 0, false
}

// GetRunMode get current run mode
func GetRunMode() string {
	return runMode
}

func parseConfigFile() {
	file, err := os.Open("./app.conf")
	if err != nil {
		fmt.Println("config file not found!")
		os.Exit(2)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if isValidLine(line) {
			line = processIfContainsComment(line)
			if !findNode { // before first node
				if isEntryLine(line) {
					_key, value := getKeyValue(line)
					if _key == "RunMode" {
						runMode = value.(string)
						runModeSet = true
					} else {
						global[_key] = value
					}
				} else if isNode(line) {
					findNode = true
					parseOneLine(line)
				}
			} else {
				if !runModeSet {
					fmt.Println("RunMode must set before any node!")
					os.Exit(2)
				}
				parseOneLine(line)
			}
		}

	}
}

func processIfContainsComment(line string) string {
	if containsComment(line) {
		i := commentStartIndex(line)
		if i > 0 { // TODO if # follows =, there mustn't be space between = and #
			line = line[0:i]
		}
	}
	return line
}

func parseOneLine(line string) {
	nodeDepth := getNodeDepth(line)
	if nodeDepth > 0 { // this line is a node
		if nodeDepth <= stackDepth {
			refreshStack(nodeDepth, getNodeContent(line, nodeDepth))
		} else {
			pushStack(getNodeContent(line, nodeDepth))
		}
	} else { // this line is an entry
		_key, value := getKeyValue(line)
		baseKey := getBaseKey()
		properties[baseKey+_key] = value
	}
}

func pushStack(content string) {
	keyStack.PushBack(content)
	stackDepth = keyStack.Len()
}

func refreshStack(depth int, content string) {
	if keyStack.Len() == 0 {
		keyStack.PushFront(content)
	} else {
		var removeKey *list.Element
		for i := 0; i < depth; i++ {
			if removeKey == nil {
				removeKey = keyStack.Front()
			} else {
				removeKey = removeKey.Next()
			}
		}
		keyStack.InsertBefore(content, removeKey)
		for keyStack.Len() > depth {
			keyStack.Remove(keyStack.Back())
		}
	}

	stackDepth = keyStack.Len()
}

func getBaseKey() string {
	key := ""
	for e := keyStack.Front(); e != nil; e = e.Next() {
		key += e.Value.(string) + ">"
	}
	return key
}

// getKeyValue get key and value
func getKeyValue(line string) (string, interface{}) {
	entries := strings.Split(line, "=")
	v := strings.TrimSpace(entries[1])
	v = strings.Replace(v, " ", "", -1)
	v = strings.TrimSpace(v)
	if evalRegex.MatchString(v) { // eval expression
		eval := v[5 : len(v)-1]
		return strings.TrimSpace(entries[0]), evaluate(eval)
	} else {
		return strings.TrimSpace(entries[0]), strings.TrimSpace(entries[1])
	}
}

func evaluate(s string) float64 {
	nus := exp.Split(s, -1)
	opes := exp.FindAllString(s, -1)
	var numbers = make([]float64, 0, len(nus))
	for i := range nus {
		f, err := strconv.ParseFloat(nus[i], 64)
		if err != nil {
			fmt.Println("ParseFloat err:", err)
			os.Exit(1)
		}
		numbers = append(numbers, f)
	}
	//fmt.Println(numbers)
	//fmt.Println(opes)

	length := len(opes)
	var c bool
	for {
		c = false
		for i := range opes {
			if opes[i] == "/" && i < length-1 && opes[i+1] == " *" {
				c = true
				t := append(numbers[0:i], numbers[i]*numbers[i+2]/numbers[i+1])
				if i < length-2 {
					numbers = append(t, numbers[i+3:]...)
					opes = append(opes[0:i], opes[i+2:]...)
				} else {
					numbers = t
					opes = opes[0:i]
				}
				break
			}
		}
		if !c {
			break
		}
	}

	length = len(opes)
	for {
		c = false
		for i := range opes {
			if opes[i] == "*" {
				c = true
				t := append(numbers[0:i], numbers[i]*numbers[i+1])
				if i < length-1 {
					numbers = append(t, numbers[i+2:]...)
					opes = append(opes[0:i], opes[i+1:]...)
				} else {
					numbers = t
					opes = opes[0:i]
				}
				break
			} else if opes[i] == "/" {
				c = true
				t := append(numbers[0:i], numbers[i]/numbers[i+1])
				if i < length-1 {
					numbers = append(t, numbers[i+2:]...)
					opes = append(opes[0:i], opes[i+1:]...)
				} else {
					numbers = t
					opes = opes[0:i]
				}
				break
			}
		}
		if !c {
			break
		}
	}

	length = len(opes)
	for {
		c = false
		for i := range opes {
			var t []float64
			if opes[i] == "+" {
				c = true
				t = append(numbers[0:i], numbers[i]+numbers[i+1])
			} else if opes[i] == "-" {
				c = true
				t = append(numbers[0:i], numbers[i]-numbers[i+1])
			}
			if i < length-1 {
				numbers = append(t, numbers[i+2:]...)
				opes = append(opes[0:i], opes[i+1:]...)
				break
			} else {
				numbers = t
				opes = opes[0:i]
				break
			}
		}
		//fmt.Println(numbers)
		if !c || len(opes) == 0 {
			break
		}
	}

	return numbers[0]
}

func getNodeContent(line string, depth int) string {
	return line[depth : len(line)-depth]
}

// not comment line and contains 'key = value'
func isValidLine(line string) bool {
	return !isCommentLine(line) && (isEntryLine(line) || isNode(line))
}

// contains 'key = value'
func isEntryLine(line string) bool {
	return strings.Contains(line, "=") && strings.Index(line, "=") > 0 && strings.Index(line, "=") < len(line)-1
}

// line starting with # is comment
func isCommentLine(line string) bool {
	return strings.HasPrefix(line, "#")
}

// comment after properties
func containsComment(line string) bool {
	return commentSegmentRegex.MatchString(line)
}

// get index of comment segment
func commentStartIndex(line string) int {
	return commentStartRegex.FindStringIndex(line)[0]
}

// must start with [, end with ] and contains content. if line == '[]', it's not a node
func isNode(line string) bool {
	return nodeRegex.MatchString(line) && len(line) > 2
}

// if line is a node, return depth, return 0 otherwise
func getNodeDepth(line string) int {
	depth := 0
	for {
		if isNode(line) {
			depth++
			line = line[1 : len(line)-1]
		} else {
			break
		}
	}
	return depth
}
