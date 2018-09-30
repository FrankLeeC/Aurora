package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

func main() {
	// s()
	testRegexp()
}

func s() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		s, _ := url.PathUnescape(r.URL.Path)
		fmt.Println(s)
		s, _ = url.QueryUnescape(r.URL.Path)
		fmt.Println(s)

	})
	http.ListenAndServe(":9090", nil)

}

func a() {
	s := `/a/b/{id}/c/{p}`
	re := regexp.MustCompile(`({\w+})`)
	params := re.FindAllString(s, -1)
	fmt.Println(params)
	s = re.ReplaceAllString(s, `[^/]+`)
	fmt.Println(s)
}

func testRegexp() {
	s := `/a/b/12common/c/678`
	re := regexp.MustCompile(`^/a/b/([^/]*)/c/6([^/]*)$`)
	fmt.Println(re.MatchString(s), re.FindStringSubmatch(s))
}
