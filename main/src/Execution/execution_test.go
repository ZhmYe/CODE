package Execution

import (
	"bufio"
	"fmt"
	"main/src/Smallbank"
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
func Test4PreemptiveExecutor(t *testing.T) {
	savingsKeys, checkingsKeys := Smallbank.GlobalSmallBank.GetKeys()
	keys := make([]string, 0)
	keys = append(keys, savingsKeys...)
	keys = append(keys, checkingsKeys...)
	fmt.Println(len(keys))
	skews := []float64{0.6}
	generator := NewGenerator(skews)
	txs := generator.GenerateTransactions(10000)
	concurrencys := []int{1}
	for i, _ := range skews {
		fmt.Println("skew=" + strconv.FormatFloat(skews[i], 'f', 2, 64) + "\n")
		for _, concurrency := range concurrencys {
			executor := NewPreemptiveExecutor(keys, concurrency, txs[i])
			finalTimeDuration, finalAbortRate := time.Duration(0), float64(0)
			for k := 0; k < 1; k++ {
				timeDuration, abortRate := executor.Execute()
				finalTimeDuration += timeDuration
				finalAbortRate += abortRate
			}
			finalTimeDuration /= 1
			finalAbortRate /= 1
			fmt.Print("	concurrency=" + strconv.Itoa(concurrency))
			fmt.Print("	Time=" + finalTimeDuration.String())
			fmt.Print("	AbortRate=" + strconv.FormatFloat(finalAbortRate*100, 'f', 2, 64))
			fmt.Println()
		}
	}
}
