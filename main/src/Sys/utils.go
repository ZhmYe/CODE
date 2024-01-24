package Sys

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
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
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		return getCurrentAbPathByCaller()
	}
	return dir
}
func SetCPU(n int) int {
	numCpus := n
	if numCpus > runtime.NumCPU() {
		numCpus = runtime.NumCPU() // 如果超过了系统支持的最大CPU数，则只能设置为系统支持的最大CPU数
	}
	runtime.GOMAXPROCS(numCpus)
	fmt.Printf("Set CPU Number：%d\n", runtime.GOMAXPROCS(-1))
	return numCpus
}
func GoRoutineSleep() {
	tmp := 0
	for k := 0; k < 100000; k++ {
		tmp++
	}
}
