package http_test

import (
	"Aurora/tools/http"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	if b, err := http.Get("http://www.baidu.com"); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println(string(b))
	}
}

func TestBytes(t *testing.T) {
	s := "中"
	a, _ := ioutil.ReadAll(strings.NewReader(s))
	b, _ := ioutil.ReadAll(bytes.NewReader([]byte(s)))
	t.Log(a, b)
}
