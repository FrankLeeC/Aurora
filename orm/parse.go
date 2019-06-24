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
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	whereRegexp    = regexp.MustCompile("(?U)---\\s*where")
	whereEndRegexp = regexp.MustCompile("---\\s*endwhere")
	andRegexp      = regexp.MustCompile("(?U)and\\s+|And\\s+|AND\\s+|ANd\\s+|AnD\\s+|aND\\s+|aNd\\s+|anD\\s+")
	orRegexp       = regexp.MustCompile("(?U)or\\s+|Or\\s+|oR\\s+|OR\\s+")
	// ifNotNilRegexp  = regexp.MustCompile("(?U)---\\s*ifnotnil\\s+\\[.*\\].*---\\s*endif") // shortest match
	ifNotNilRegexp = regexp.MustCompile("(?U)---\\s*ifnotnil\\s+\\[.*\\]") // shortest match
	ifEndRegexp    = regexp.MustCompile("---\\s*endif")
	// ifRegexp        = regexp.MustCompile("(?U)---\\s*if\\s+\\[.*\\].*---\\s*endif")
	ifRegexp        = regexp.MustCompile("(?U)---\\s*if\\s+\\[.*\\]")
	setRegexp       = regexp.MustCompile("(?U)---\\s*set")
	setEndRegexp    = regexp.MustCompile("---\\s*endset")
	rangeRegexp     = regexp.MustCompile("(?U)---\\s*range\\s+\\[.*\\]")
	rangeEndRegexp  = regexp.MustCompile("---\\s*endrange")
	tableNameRegexp = regexp.MustCompile("@table_\\w+@")

	paramRegexp = regexp.MustCompile(`(<=?|>=?|!?=|\s*)\s*#\w+(\.\w+)?#`)

	notEqualRegexp = regexp.MustCompile("(?U)\\w+\\s*!=.*")
	equalRegexp    = regexp.MustCompile("(?U)\\w+\\s*==.*")
	smallerRegexp  = regexp.MustCompile("(?U)\\w+\\s*<.*")
	largerRegexp   = regexp.MustCompile("(?U)\\w+\\s*>.*")
	leRegexp       = regexp.MustCompile("(?U)\\w+\\s*<=.*")
	geRegexp       = regexp.MustCompile("(?U)\\w+\\s*>=.*")

	floatRegexp = regexp.MustCompile("^\\d+\\.\\d+$")
	intRegexp   = regexp.MustCompile("^\\d+$")

	// SHOWSQL whether or not show sql
	SHOWSQL = false
)

const (
	ifconst = iota
	ifnotnilconst
	rangeconst
	setconst
	whereconst
)

// judge ==, !=, <=, <, >=, >
func judge(param M, kvs []kv, expression string) (string, bool) {
	var pv interface{}
	var value string
	var variable string
	var flag int
	if equalRegexp.MatchString(expression) { // ==
		flag = 0
		i := strings.Index(expression, "==")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+2:])
	} else if notEqualRegexp.MatchString(expression) { //  !=
		flag = 1
		i := strings.Index(expression, "!=")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+2:])
	} else if leRegexp.MatchString(expression) { // <=
		flag = 2
		i := strings.Index(expression, "<=")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+2:])
	} else if smallerRegexp.MatchString(expression) { // <
		flag = 3
		i := strings.Index(expression, "<")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+1:])
	} else if geRegexp.MatchString(expression) { // >=
		flag = 4
		i := strings.Index(expression, ">=")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+2:])
	} else { // >
		flag = 5
		i := strings.Index(expression, ">")
		variable = strings.TrimSpace(expression[0:i])
		value = strings.TrimSpace(expression[i+1:])
	}
	var f bool
	pv, f = getValue(variable, param, kvs)
	if !f {
		return variable, false
	}
	// pv = param[variable]
	// if pv == nil {
	// 	return false
	// }
	if isString(value) { // string
		if flag == 0 {
			return variable, pv.(string) == value
		} else if flag == 1 {
			return variable, pv.(string) != value
		} else if flag == 2 {
			return variable, strings.Compare(pv.(string), value) <= 0
		} else if flag == 3 {
			return variable, strings.Compare(pv.(string), value) < 0
		} else if flag == 4 {
			return variable, strings.Compare(pv.(string), value) >= 0
		}
		return variable, strings.Compare(pv.(string), value) > 0
	} else if isFloat(value) {
		switch pv.(type) {
		case float32:
			if flag == 0 {
				return variable, float64(pv.(float32)) == getFloatValue(value)
			} else if flag == 1 {
				return variable, float64(pv.(float32)) != getFloatValue(value)
			} else if flag == 2 {
				return variable, float64(pv.(float32)) <= getFloatValue(value)
			} else if flag == 3 {
				return variable, float64(pv.(float32)) < getFloatValue(value)
			} else if flag == 4 {
				return variable, float64(pv.(float32)) >= getFloatValue(value)
			}
			return variable, float64(pv.(float32)) > getFloatValue(value)
		case float64:
			if flag == 0 {
				return variable, pv.(float64) == getFloatValue(value)
			} else if flag == 1 {
				return variable, pv.(float64) != getFloatValue(value)
			} else if flag == 2 {
				return variable, pv.(float64) <= getFloatValue(value)
			} else if flag == 3 {
				return variable, pv.(float64) < getFloatValue(value)
			} else if flag == 4 {
				return variable, pv.(float64) >= getFloatValue(value)
			}
			return variable, pv.(float64) > getFloatValue(value)
		}
	} else if isInt(value) {
		switch pv.(type) {
		case uint8:
			if flag == 0 {
				return variable, uint64(pv.(uint8)) == getUintValue(value)
			} else if flag == 1 {
				return variable, uint64(pv.(uint8)) != getUintValue(value)
			} else if flag == 2 {
				return variable, uint64(pv.(uint8)) <= getUintValue(value)
			} else if flag == 3 {
				return variable, uint64(pv.(uint8)) < getUintValue(value)
			} else if flag == 4 {
				return variable, uint64(pv.(uint8)) >= getUintValue(value)
			}
			return variable, uint64(pv.(uint8)) > getUintValue(value)
		case uint:
			if flag == 0 {
				return variable, uint64(pv.(uint)) == getUintValue(value)
			} else if flag == 1 {
				return variable, uint64(pv.(uint)) != getUintValue(value)
			} else if flag == 2 {
				return variable, uint64(pv.(uint)) <= getUintValue(value)
			} else if flag == 3 {
				return variable, uint64(pv.(uint)) < getUintValue(value)
			} else if flag == 4 {
				return variable, uint64(pv.(uint)) >= getUintValue(value)
			}
			return variable, uint64(pv.(uint)) > getUintValue(value)
		case uint16:
			if flag == 0 {
				return variable, uint64(pv.(uint16)) == getUintValue(value)
			} else if flag == 1 {
				return variable, uint64(pv.(uint16)) != getUintValue(value)
			} else if flag == 2 {
				return variable, uint64(pv.(uint16)) <= getUintValue(value)
			} else if flag == 3 {
				return variable, uint64(pv.(uint16)) < getUintValue(value)
			} else if flag == 4 {
				return variable, uint64(pv.(uint16)) >= getUintValue(value)
			}
			return variable, uint64(pv.(uint16)) > getUintValue(value)
		case uint32:
			if flag == 0 {
				return variable, uint64(pv.(uint32)) == getUintValue(value)
			} else if flag == 1 {
				return variable, uint64(pv.(uint32)) != getUintValue(value)
			} else if flag == 2 {
				return variable, uint64(pv.(uint32)) <= getUintValue(value)
			} else if flag == 3 {
				return variable, uint64(pv.(uint32)) < getUintValue(value)
			} else if flag == 4 {
				return variable, uint64(pv.(uint32)) >= getUintValue(value)
			}
			return variable, uint64(pv.(uint32)) > getUintValue(value)
		case uint64:
			if flag == 0 {
				return variable, uint64(pv.(uint64)) == getUintValue(value)
			} else if flag == 1 {
				return variable, uint64(pv.(uint64)) != getUintValue(value)
			} else if flag == 2 {
				return variable, uint64(pv.(uint64)) <= getUintValue(value)
			} else if flag == 3 {
				return variable, uint64(pv.(uint64)) < getUintValue(value)
			} else if flag == 4 {
				return variable, uint64(pv.(uint64)) >= getUintValue(value)
			}
			return "", uint64(pv.(uint64)) > getUintValue(value)
		case int8:
			if flag == 0 {
				return variable, int64(pv.(int8)) == getIntValue(value)
			} else if flag == 1 {
				return variable, int64(pv.(int8)) != getIntValue(value)
			} else if flag == 2 {
				return variable, int64(pv.(int8)) <= getIntValue(value)
			} else if flag == 3 {
				return variable, int64(pv.(int8)) < getIntValue(value)
			} else if flag == 4 {
				return variable, int64(pv.(int8)) >= getIntValue(value)
			}
			return variable, int64(pv.(int8)) > getIntValue(value)
		case int:
			if flag == 0 {
				return variable, int64(pv.(int)) == getIntValue(value)
			} else if flag == 1 {
				return variable, int64(pv.(int)) != getIntValue(value)
			} else if flag == 2 {
				return variable, int64(pv.(int)) <= getIntValue(value)
			} else if flag == 3 {
				return variable, int64(pv.(int)) < getIntValue(value)
			} else if flag == 4 {
				return variable, int64(pv.(int)) >= getIntValue(value)
			}
			return variable, int64(pv.(int)) > getIntValue(value)
		case int16:
			if flag == 0 {
				return variable, int64(pv.(int16)) == getIntValue(value)
			} else if flag == 1 {
				return variable, int64(pv.(int16)) != getIntValue(value)
			} else if flag == 2 {
				return variable, int64(pv.(int16)) <= getIntValue(value)
			} else if flag == 3 {
				return variable, int64(pv.(int16)) < getIntValue(value)
			} else if flag == 4 {
				return variable, int64(pv.(int16)) >= getIntValue(value)
			}
			return variable, int64(pv.(int16)) > getIntValue(value)
		case int32:
			if flag == 0 {
				return variable, int64(pv.(int32)) == getIntValue(value)
			} else if flag == 1 {
				return variable, int64(pv.(int32)) != getIntValue(value)
			} else if flag == 2 {
				return variable, int64(pv.(int32)) <= getIntValue(value)
			} else if flag == 3 {
				return variable, int64(pv.(int32)) < getIntValue(value)
			} else if flag == 4 {
				return variable, int64(pv.(int32)) >= getIntValue(value)
			}
			return variable, int64(pv.(int32)) > getIntValue(value)
		case int64:
			if flag == 0 {
				return variable, int64(pv.(int64)) == getIntValue(value)
			} else if flag == 1 {
				return variable, int64(pv.(int64)) != getIntValue(value)
			} else if flag == 2 {
				return variable, int64(pv.(int64)) <= getIntValue(value)
			} else if flag == 3 {
				return variable, int64(pv.(int64)) < getIntValue(value)
			} else if flag == 4 {
				return variable, int64(pv.(int64)) >= getIntValue(value)
			}
			return variable, int64(pv.(int64)) > getIntValue(value)
		}
	} else if isBool(value) {
		if flag == 0 {
			return variable, pv.(bool) == getBoolValue(value)
		}
		return variable, pv.(bool) != getBoolValue(value)
	}
	// default is string
	if flag == 0 {
		return variable, pv.(string) == value
	} else if flag == 1 {
		return variable, pv.(string) != value
	} else if flag == 2 {
		return variable, strings.Compare(pv.(string), value) <= 0
	} else if flag == 3 {
		return variable, strings.Compare(pv.(string), value) < 0
	} else if flag == 4 {
		return variable, strings.Compare(pv.(string), value) >= 0
	}
	return variable, strings.Compare(pv.(string), value) > 0
}

// used in judgement
func isString(s string) bool {
	return strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")
}

// used in judgement
func isFloat(s string) bool {
	return floatRegexp.MatchString(s)
}

// used in judgement
func isInt(s string) bool {
	return intRegexp.MatchString(s)
}

// used in judgement
func isBool(s string) bool {
	return s == "True" || s == "true" || s == "TRUE" || s == "T" || s == "F" || s == "False" || s == "false" || s == "FALSE"
}

func getFloatValue(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func getUintValue(v string) uint64 {
	i, _ := strconv.ParseUint(v, 10, 64)
	return i
}

func getIntValue(v string) int64 {
	i, _ := strconv.ParseInt(v, 10, 64)
	return i
}

func getBoolValue(v string) bool {
	return v == "TRUE" || v == "T" || v == "true" || v == "True"
}

func match(s string) (bool, int) {
	if ifRegexp.MatchString(s) {
		return true, ifconst
	}
	if ifNotNilRegexp.MatchString(s) {
		return true, ifnotnilconst
	}
	if rangeRegexp.MatchString(s) {
		return true, rangeconst
	}
	if setRegexp.MatchString(s) {
		return true, setconst
	}
	if whereRegexp.MatchString(s) {
		return true, whereconst
	}
	return false, -1
}

func content(s string, r *regexp.Regexp) string {
	return space + s[strings.Index(s, "]")+1:r.FindIndex([]byte(s))[0]] + space
}

func before(s string, r *regexp.Regexp) string {
	return space + s[0:r.FindIndex([]byte(s))[0]] + space
}

func after(s string, r *regexp.Regexp) string {
	return space + s[r.FindIndex([]byte(s))[1]:] + space
}

func expression(s string) string {
	a := strings.Index(s, "[")
	b := strings.Index(s, "]")
	return strings.TrimSpace(s[a+1 : b])
}

// return expression and whether meet condition or not
func meetCondition(s string, m M, kvs []kv, t int) ([]string, bool) {
	c := expression(s)
	if t == ifconst {
		s, f := judge(m, kvs, c)
		return []string{s}, f
	} else if t == ifnotnilconst {
		c = strings.TrimSpace(c)
		var r = false
		for i := range kvs {
			if kvs[i].k == c {
				r = true
				break
			}
		}
		if !r {
			_, r = m[c]
		}
		return []string{c}, r
	} else if t == rangeconst {
		c = strings.TrimSpace(c)
		cc := strings.Split(c, ",")
		rangeParam := cc[0]
		rp, r := m[rangeParam]
		if !r {
			return nil, false
		}
		slice := reflect.ValueOf(rp)
		if slice.Len() <= 0 {
			return nil, false
		}
		return cc, true
	}
	return nil, false
}

func end(s string) bool {
	b := ifEndRegexp.MatchString(s)
	if !b {
		b = rangeEndRegexp.MatchString(s)
	}
	if !b {
		b = setEndRegexp.MatchString(s)
	}
	if !b {
		b = whereEndRegexp.MatchString(s)
	}
	return b
}
