package main

import (
	"fmt"
	"main/src/Config"
	"main/src/Evaluate"
	"main/src/Sys"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	Config.GlobalConfig.LogPath = Sys.GetCurrentAbPath()
	rand.Seed(time.Now().UnixNano())
	fmt.Printf("Max CPU Number:%d\n", runtime.NumCPU())
	//Evaluate.RunE1WithDifferentParams([]int{2, 4, 8, 16, 32}, "log/E1/")
	//Evaluate.RunE2WithDifferentParams([]int{100, 200, 500, 1000, 2000, 5000}, "log/E2/")
	Evaluate.EvaluateLatencyAndTpsInDifferentInstance("log/E3/")
	//Evaluate.EvaluateAbortRateAndTpsWithDifferentInstanceMode("log/E4/")
}
