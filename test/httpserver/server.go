package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/FrankLeeC/Aurora/httpserver"
)

func main() {
	s := httpserver.NewHTTPServer(9090)
	s.Route("/test", test)
	s.Route("/ok", ok)
	s.DynamicFilter("/{url}", &testFilter{})
	s.Finish(func(e error) {
		fmt.Printf("close server error: %v\n", e)
	})
	s.ServeHTTP()
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	fmt.Printf("receive sig: %v\n", sig)
	s.Shutdown()
}

func test(rsp *httpserver.Response, req *httpserver.Request) (uint, error) {
	rsp.Write([]byte("test"))
	return http.StatusOK, nil
}

func ok(rsp *httpserver.Response, req *httpserver.Request) (uint, error) {
	rsp.Write([]byte("ok"))
	return http.StatusOK, nil
}

type testFilter struct {
}

func (a *testFilter) Before(rsp *httpserver.Response, r *httpserver.Request) bool {
	rsp.Header().Add("O", "2")
	return true
}

func (a *testFilter) After(rsp *httpserver.Response, r *httpserver.Request) {
	rsp.Header().Add("h", "1")
}
