package Execution

import (
	"main/src/Algorithm/Utils"
	"main/src/Blockchain"
	"main/src/Sys"
	"strconv"
	"sync"
	"time"
)

type Transaction = Blockchain.Transaction
type PreemptiveExecutor struct {
	concurrency  int // 线程数
	transactions []*Transaction
	SVM          *Utils.SVM // 并发多版本address-based map
}

func NewPreemptiveExecutor(keys []string, concurrency int, transactions []*Blockchain.Transaction) *PreemptiveExecutor {
	e := new(PreemptiveExecutor)
	e.concurrency = concurrency
	e.transactions = transactions
	e.SVM = Utils.NewSVM(keys)
	return e
}
func (e *PreemptiveExecutor) Execute() (time.Duration, float64) {
	// 一共concurrency个线程，每个线程通过channel抢占式的得到下一个tx，基于SVM内的逻辑进行并行+串行的混合过程，在执行的过程中已完成abort
	startTime := time.Now()
	numTask := len(e.transactions)
	tasks := make(chan *Transaction, numTask)
	var wg4worker sync.WaitGroup
	wg4worker.Add(e.concurrency)
	for x := 0; x < e.concurrency; x++ {
		go func(tasks <-chan *Transaction, wg *sync.WaitGroup) {
			defer wg.Done()
			for tx := range tasks {
				versions := make(map[string]int, 0)
				Sys.GoRoutineSleep()
				for _, op := range tx.Ops {
					if op.Type == Blockchain.OpRead {
						readResult, version := e.SVM.Read(op.Key)
						op.ReadResult = readResult
						versions[op.Key] = version
					}
					if op.Type == Blockchain.OpWrite {
						readResult, version := e.SVM.Read(op.Key)
						versions[op.Key] = version
						amount, _ := strconv.Atoi(op.Val)
						readResultInt, _ := strconv.Atoi(readResult)
						WriteResult := readResultInt + amount
						op.WriteResult = strconv.Itoa(WriteResult)
					}
				}
				e.SVM.Commit(tx, versions)
				//fmt.Println(tx.GetId())
			}
		}(tasks, &wg4worker)
	}
	for _, transaction := range e.transactions {
		//fmt.Println(i)
		tasks <- transaction
	}
	close(tasks)
	wg4worker.Wait()
	//finalAbortRate += abortRate
	abortNumber := 0
	e.SVM.CommitLastWriteToDB()
	for _, tx := range e.transactions {
		if tx.CheckAbort() {
			abortNumber += 1
		}
		tx.Reset()
	}
	e.SVM.Reset()
	//for _, value := range e.SVM.Address2Transactions {
	//	fmt.Println(value.GetVersion())
	//}

	return time.Since(startTime), float64(abortNumber) / float64(len(e.transactions))
}
