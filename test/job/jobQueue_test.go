package test

import (
	"strings"
	"testing"

	"github.com/FrankLeeC/Aurora/job"

	"github.com/go-redis/redis"
)

var clt *redis.ClusterClient

func TestJob(t *testing.T) {
	n := 10000
	clt = initRedis()
	clt.Del("test_job_queue")
	c := make(chan int, n)
	jq := job.NewJobQueue(200, 2)
	for i := 0; i < n; i++ {
		jq.Submit(&task{c: c})
	}
	for i := 0; i < n; i++ {
		<-c
	}
	t.Log("--------rs:", clt.Get("test_job_queue").Val())
}

func initRedis() *redis.ClusterClient {
	addr := "***"
	password := "***"
	client := redis.NewClusterClient((&redis.ClusterOptions{
		Addrs:    strings.Split(addr, ","),
		Password: password,
	}))
	_, err := client.Ping().Result()
	if err != nil {
		return nil
	}
	return client
}

type task struct {
	r int64
	c chan int
}

func (t *task) Work() {
	t.r = clt.Incr("test_job_queue").Val()
	t.c <- 1
}
