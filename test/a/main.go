package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func main() {
	s()
	// testRegexp()
	// a()
}

func s() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		s, _ := url.PathUnescape(r.URL.Path)
		fmt.Println(s)
		s, _ = url.QueryUnescape(r.URL.Path)
		fmt.Println(s)
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		fmt.Println(string(b))
		b, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(string(b))
		// r.ParseForm()
		// fmt.Println(r.Form)
		// r.ParseForm()
		// fmt.Println(r.Form)

	})
	http.ListenAndServe(":9090", nil)

}

func a() {
	s := `/a/b/{id3}/c/{p}`
	re := regexp.MustCompile(`({\w+})`)
	params := re.FindAllString(s, -1)
	for i := range params {
		params[i] = params[i][1 : len(params[i])-1]
	}
	fmt.Println(params)
	s = "^" + re.ReplaceAllString(s, `([^/]+)`) + "$"
	fmt.Println(s)
	test := `/a/b/12common/c/678`
	re2 := regexp.MustCompile(s)
	fmt.Println(re2.MatchString(test), re2.FindStringSubmatch(test))
}

func testRegexp() {
	s := `/a/b/12/common/c/678`
	re := regexp.MustCompile(`^/a/b/([^/]+)/c/6([^/]+)$`)
	fmt.Println(re.MatchString(s), re.FindStringSubmatch(s))
}
