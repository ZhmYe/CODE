package Execution

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func Test4Split(t *testing.T) {
	skews := []float64{0.6, 0.8, 0.99}
	generator := NewGenerator(skews)
	txs := generator.GenerateTransactions(10000)
	concurrencys := []int{2, 4, 8, 16}
	for i, _ := range skews {
		for _, concurrency := range concurrencys {
			executor := NewExecutor(ExecuteWithFabric, concurrency, txs[i])
			executor.SplitTransactions()
		}
	}
}
func Test4Execution(t *testing.T) {
	skews := []float64{0.6, 0.8, 0.99}
	generator := NewGenerator(skews)
	txs := generator.GenerateTransactions(10000)
	concurrencys := []int{2, 4, 8, 16, 32, 64, 128, 256}
	currentTime := time.Now()
	format := "2006-01-02-15-04-05"
	filePath := "./" + currentTime.Format(format) + ".txt"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	writer := bufio.NewWriter(file)

	if err != nil {
		fmt.Printf("open file err=%v\n", err)
		return
	}
	for i, _ := range skews {
		writer.WriteString("skew=" + strconv.FormatFloat(skews[i], 'f', 2, 64) + "\n")
		for _, concurrency := range concurrencys {
			executor := NewExecutor(ExecuteWithFabric, concurrency, txs[i])
			executor.SplitTransactions()
			finalTimeDuration, finalAbortRate := time.Duration(0), float64(0)
			for k := 0; k < 100; k++ {
				timeDuration, abortRate := executor.Execute()
				finalTimeDuration += timeDuration
				finalAbortRate += abortRate
			}
			finalTimeDuration /= 100
			finalAbortRate /= 100
			writer.WriteString("	concurrency=" + strconv.Itoa(concurrency))
			writer.WriteString("	Time=" + finalTimeDuration.String())
			writer.WriteString("	AbortRate=" + strconv.FormatFloat(finalAbortRate*100, 'f', 2, 64))
			writer.WriteString("\n")
		}
	}
	err = writer.Flush()
	if err != nil {
		return
	}
	err = file.Close()
	if err != nil {
		return
	}
}
