package Evaluate

import (
	"main/src/Consensus"
	"main/src/Execution"
	"main/src/Logger"
	"main/src/Smallbank"
	"main/src/Sys"
	"strconv"
	"time"
)

var globalSmallBank = Smallbank.GlobalSmallBank

// EvaluateTpsAndAbortNumberWithDifferentConcurrency E1
func EvaluateTpsAndAbortNumberWithDifferentConcurrency(CPUNumber int, path string) {
	skews := []float64{0.6}
	generator := Execution.NewGenerator(skews)
	txs := generator.GenerateTransactions(10000)
	concurrencys := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}
	logger := Logger.NewLogger(path)
	Sys.SetCPU(CPUNumber)
	logger.Write("CPU=" + strconv.Itoa(CPUNumber))
	logger.Wrap()
	for i, _ := range skews {
		logger.Write("skew=" + strconv.FormatFloat(skews[i], 'f', 2, 64) + "\n")
		for _, concurrency := range concurrencys {
			executor := Execution.NewExecutor(Execution.ExecuteWithFabric, concurrency, txs[i])
			executor.SplitTransactions()
			finalTimeDuration, finalAbortRate := time.Duration(0), float64(0)
			for k := 0; k < 100; k++ {
				timeDuration, abortRate := executor.SimpleExecute()
				finalTimeDuration += timeDuration
				finalAbortRate += abortRate
			}
			finalTimeDuration /= 100
			finalAbortRate /= 100
			logger.Write("\tconcurrency=" + strconv.Itoa(concurrency))
			logger.Write("\tTime=" + finalTimeDuration.String())
			logger.Write("\tAbortRate=" + strconv.FormatFloat(finalAbortRate*100, 'f', 2, 64))
			logger.Wrap()
		}
	}
	logger.Finish()
}

// EvaluateAbortRateAndTpsWithDifferentBlockSize E2
func EvaluateAbortRateAndTpsWithDifferentBlockSize(blockSize int, path string) {
	skews := []float64{0.6, 0.7, 0.8, 0.9, 0.99}
	generator := Execution.NewGenerator(skews)
	txs := generator.GenerateTransactions(blockSize) // 简单的用一定数量的交易来代替区块
	// 这里只设置全并发和少量并发(64)
	CompleteConcurrency, OptimalConcurrency := blockSize, 64
	logger := Logger.NewLogger(path)
	Sys.SetCPU(8)
	logger.Write("BlockSize=" + strconv.Itoa(blockSize))
	logger.Wrap()
	for i, _ := range skews {
		logger.Write("skew=" + strconv.FormatFloat(skews[i], 'f', 2, 64) + "\n")
		// 首先全量并发测延时和abort rate
		executor := Execution.NewExecutor(Execution.ExecuteWithFabric, CompleteConcurrency, txs[i])
		executor.SplitTransactions()
		finalTimeDuration, finalAbortRate := time.Duration(0), float64(0)
		for k := 0; k < 100; k++ {
			timeDuration, abortRate := executor.SimpleExecute()
			finalTimeDuration += timeDuration
			finalAbortRate += abortRate
		}
		finalTimeDuration /= 100
		finalAbortRate /= 100
		logger.Write("\tComplete Concurrency")
		logger.Wrap()
		logger.Write("\t\tTime=" + finalTimeDuration.String())
		logger.Write("\t\tAbortRate=" + strconv.FormatFloat(finalAbortRate*100, 'f', 2, 64))
		logger.Wrap()
		// 然后少量并发测延时和abort rate
		executor = Execution.NewExecutor(Execution.ExecuteWithFabric, OptimalConcurrency, txs[i])
		executor.SplitTransactions()
		finalTimeDuration, finalAbortRate = time.Duration(0), float64(0)
		for k := 0; k < 100; k++ {
			timeDuration, abortRate := executor.SimpleExecute()
			finalTimeDuration += timeDuration
			finalAbortRate += abortRate
		}
		finalTimeDuration /= 100
		finalAbortRate /= 100
		logger.Write("\tOptimal Concurrency")
		logger.Wrap()
		logger.Write("\t\tTime=" + finalTimeDuration.String())
		logger.Write("\t\tAbortRate=" + strconv.FormatFloat(finalAbortRate*100, 'f', 2, 64))
		logger.Wrap()
	}
	logger.Finish()
}

// EvaluateLatencyAndTpsInDifferentInstance E3
func EvaluateLatencyAndTpsInDifferentInstance(path string) {
	Sys.SetCPU(8)
	peer := Consensus.NewPeer(4, 64, 10000, 0.6)
	reports := peer.Run()
	logger := Logger.NewLogger(path)
	logger.Write("Concurrency=128")
	logger.Wrap()
	for i, report := range reports {
		logger.Write("Instance " + strconv.Itoa(i))
		logger.Write("\tConcurrency=" + strconv.Itoa(report.GetConcurrency()))
		logger.Write("\tTime=" + report.GetProcessTime().String())
		logger.Write("\tTx Number=" + strconv.Itoa(report.GetProcessTransactionNumber()))
		logger.Write("\tAbort Rate=" + strconv.FormatFloat(report.GetAbortRate()*100, 'f', 2, 64))
		logger.Wrap()
	}
	logger.Finish()
}

func EvaluateAbortRateAndTpsWithDifferentInstanceMode(path string) {
	Sys.SetCPU(8)
	modes := []Consensus.InstanceMode{Consensus.Default, Consensus.PreBatched, Consensus.Preemptive}
	logger := Logger.NewLogger(path)
	logger.Write("Concurrency=64")
	logger.Wrap()
	for _, mode := range modes {
		generator := Execution.NewGenerator([]float64{0.6})
		transactions := generator.GenerateTransactions(1000)[0]
		instance := Consensus.NewInstance()
		instance.SetMode(mode)
		switch mode {
		case Consensus.Default:
			logger.Write("mode=Default")
		case Consensus.PreBatched:
			logger.Write("mode=PreBatched")
		case Consensus.Preemptive:
			logger.Write("mode=Preemptive")

		}
		logger.Wrap()
		instance.SetConcurrency(64)
		instance.InjectTransactions(transactions)
		instance.Run()
		report := instance.GetReport()
		logger.Write("\tTx Number=" + strconv.Itoa(report.GetProcessTransactionNumber()))
		logger.Write("\tTime=" + report.GetProcessTime().String())
		logger.Write("\tAbort Rate=" + strconv.FormatFloat(report.GetAbortRate()*100, 'f', 2, 64))
		logger.Wrap()
	}
	logger.Finish()
}
