package Execution

import (
	"fmt"
	"main/src/Algorithm/Fabric"
	"main/src/Algorithm/FabricPP"
	"main/src/Algorithm/Harmony"
	"main/src/Algorithm/Nezha"
	"main/src/Blockchain"
	"main/src/Smallbank"
	"strconv"
	"sync"
	"time"
)

type ExecuteMethod int

const (
	ExecuteWithFabric ExecuteMethod = iota
	ExecuteWithFabricpp
	ExecuteWithNezha
	ExecuteWithHarmony
)

type Executor struct {
	method       ExecuteMethod
	concurrency  int
	transactions []*Blockchain.Transaction
	split        [][]*Blockchain.Transaction
}

func NewExecutor(method ExecuteMethod, concurrency int, transactions []*Blockchain.Transaction) *Executor {
	executor := new(Executor)
	executor.method = method
	executor.transactions = transactions
	executor.concurrency = concurrency
	executor.split = make([][]*Blockchain.Transaction, 0)
	return executor
}
func (e *Executor) SplitTransactions() {
	index := 0
	for {
		if index+e.concurrency >= len(e.transactions) {
			// 这里超过下标，将剩下的看成一个整体
			e.split = append(e.split, e.transactions[index:])
			break
		}
		e.split = append(e.split, e.transactions[index:index+e.concurrency])
		index += e.concurrency
	}
	fmt.Println("Split Finished, Split Number: " + strconv.Itoa(len(e.split)))
}
func (e *Executor) Execute() (time.Duration, float64) {
	// 不同split之间串行, split内部并行并使用某种abort算法
	startTime := time.Now()
	for _, transactions := range e.split {
		var wg4tx sync.WaitGroup
		wg4tx.Add(len(transactions))
		for _, tx := range transactions {
			tmpTx := tx
			go func(tx *Blockchain.Transaction) {
				defer wg4tx.Done()
				for _, op := range tx.Ops {
					if op.Type == Blockchain.OpRead {
						readResult, _ := strconv.Atoi(Smallbank.GlobalSmallBank.Read(op.Key))
						op.ReadResult = strconv.Itoa(readResult)
					}
					if op.Type == Blockchain.OpWrite {
						readResult, _ := strconv.Atoi(Smallbank.GlobalSmallBank.Read(op.Key))
						amount, _ := strconv.Atoi(op.Val)
						WriteResult := readResult + amount
						op.WriteResult = strconv.Itoa(WriteResult)
					}
				}
			}(tmpTx)
		}
		wg4tx.Wait()
		// 到此，一个交易分片执行完成，对其进行abort
		switch e.method {
		case ExecuteWithFabric:
			fabric := Fabric.NewFabric(transactions)
			fabric.TransactionSort()
		case ExecuteWithFabricpp:
			fabricPP := FabricPP.NewFabricPP(transactions)
			fabricPP.TransactionSort()
		case ExecuteWithNezha:
			nezha := Nezha.NewNeZha(transactions)
			nezha.TransactionSort()
		case ExecuteWithHarmony:
			harmony := Harmony.NewHarmony(transactions)
			harmony.TransactionSort()
		}
		var wg4exec sync.WaitGroup
		wg4exec.Add(len(transactions))
		for _, tx := range transactions {
			tmpTx := tx
			go func(tx *Blockchain.Transaction) {
				if tx.CheckAbort() {
					wg4exec.Done()
				} else {
					for _, op := range tx.GetOps() {
						if op.Type == Blockchain.OpWrite {
							Smallbank.GlobalSmallBank.Write(op.Key, op.WriteResult)
						}
					}
					wg4exec.Done()
				}
			}(tmpTx)
		}
		wg4exec.Wait()
	}
	totalAbortNumber := 0
	timeDuration := time.Since(startTime)
	for _, transactions := range e.split {
		for _, tx := range transactions {
			if tx.CheckAbort() {
				totalAbortNumber += 1
			}
			tx.Reset()
		}
	}
	return timeDuration, float64(totalAbortNumber) / float64(len(e.transactions))
}
