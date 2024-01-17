package main

import (
	"fmt"
	"main/src/Evaluate"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	maxProcs := runtime.NumCPU()
	fmt.Printf("当前系统支持的最大CPU数：%d\n", maxProcs)

	// 设置最大CPU数为4
	numCpus := 4
	if numCpus > maxProcs {
		numCpus = maxProcs // 如果超过了系统支持的最大CPU数，则只能设置为系统支持的最大CPU数
	}
	runtime.GOMAXPROCS(numCpus)
	fmt.Printf("已设置最大CPU数为：%d\n", runtime.GOMAXPROCS(-1))
	Evaluate.EvaluateTpsAndAbortNumberWithDifferentConcurrency("data/")
}
