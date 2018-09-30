package main

import (
	"fmt"

	"github.com/FrankLeeC/Aurora/config"
)

func main() {
	fmt.Printf("RunMode=%s\n", config.GetRunMode())
	fmt.Printf("foo=%s\n", config.GetString("foo"))
	fmt.Printf("abc=%d\n", config.GetInt("abc"))
	fmt.Printf("port=%s\n", config.GetString("port"))
	fmt.Printf("fzz=%f\n", config.GetEval("fzz"))
	fmt.Printf("fxx=%f\n", config.GetEval("fxx"))

	fmt.Printf("mysql>defaultPagesize=%d\n", config.GetInt("mysql>defaultPagesize"))
	fmt.Printf("mysql>source1>uri=%s\n", config.GetString("mysql>source1>uri"))
	fmt.Printf("mysql>source1>useSomething=%v\n", config.GetBool("mysql>source1>useSomething"))

	fmt.Printf("mysql>source2>uri=%s\n", config.GetString("mysql>source2>uri"))

	fmt.Println("--------------------")
	m := config.GetAll("mysql", false)
	fmt.Printf("getAll: %v\n", m)

	fmt.Println("----------------------")
	m = config.GetAll("mysql", true)
	fmt.Printf("getAll recursely: %v\n", m)
}
