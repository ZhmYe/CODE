package main

import (
	"fmt"
	"log"
	"main/src/Config"
	"main/src/Evaluate"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

// 获取系统临时目录，兼容go run
func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}
func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}
	return dir
}
func main() {
	rootPath := getCurrentAbPath()
	Config.GlobalConfig.LogPath = rootPath
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
	Evaluate.EvaluateTpsAndAbortNumberWithDifferentConcurrency("log/")
}
