package Consensus

import (
	"main/src/Algorithm/Harmony"
	"main/src/Algorithm/Utils"
	"main/src/Blockchain"
	"main/src/Config"
	"main/src/Smallbank"
	"strconv"
	"sync"
)

var config = Config.GlobalConfig
var globalSmallBank = Smallbank.GlobalSmallBank

type SimulateEngine struct {
	acgs   []Utils.ACG
	buffer map[string]int // 存储真正写入前各地址的缓存
	blocks []*Block       // 当前要执行的区块的时候
}

func newSimulateEngine(blocks []*Block) *SimulateEngine {
	e := new(SimulateEngine)
	e.acgs = make([]Utils.ACG, 0)

	e.buffer = make(map[string]int, 0)
	e.blocks = blocks
	return e
}

// SimulateExecution 模拟执行
func (e *SimulateEngine) SimulateExecution() []Utils.ACG {
	// 并发执行粒度
	parallelingNumber := config.ParallelingNumber
	// 依次遍历每个区块
	for _, block := range e.blocks {
		// 每次取出parallelingNumber笔交易并行执行
		tmp := block.GetTransactionLength()
		if block.GetTransactionLength()%parallelingNumber != 0 {
			tmp += parallelingNumber - block.GetTransactionLength()%parallelingNumber
		}
		for j := 0; j < tmp; j += parallelingNumber {
			// 并行
			var wg4tx sync.WaitGroup
			wg4tx.Add(parallelingNumber)
			for k := 0; k < parallelingNumber; k++ {
				if j+k >= block.GetTransactionLength() {
					wg4tx.Done()
					continue
				}
				tmpTx := block.GetTransaction(j + k)
				tmpBuffer := e.buffer
				go func(tx *Transaction, wg4tx *sync.WaitGroup, buffer map[string]int) {
					defer wg4tx.Done()
					//if index+bias >= len(block.txs) {
					//	return
					//}
					//tx := block.txs[index+bias]
					for _, op := range tx.Ops {
						if op.Type == Blockchain.OpRead {
							readResult, exist := buffer[op.Key]
							if !exist {
								readResult, _ = strconv.Atoi(Smallbank.GlobalSmallBank.Read(op.Key))
							}
							op.ReadResult = strconv.Itoa(readResult)
						}
						if op.Type == Blockchain.OpWrite {
							readResult, exist := buffer[op.Key]
							if !exist {
								readResult, _ = strconv.Atoi(globalSmallBank.Read(op.Key))
							}
							amount, _ := strconv.Atoi(op.Val)
							WriteResult := readResult + amount
							//buffer[op.Key] = WriteResult
							op.WriteResult = strconv.Itoa(WriteResult)
							//globalSmallBank.Write(op.Key, strconv.Itoa(WriteResult))
						}
					}
				}(tmpTx, &wg4tx, tmpBuffer)
			}
			wg4tx.Wait()
		}
		// 这里获取到了buffer，下一个区块基于上一个区块的buffer来做，所以要先abort，然后buffer取ACG中每个地址的最后一个写
		//nezha := newNeZha(block.txs)
		//nezha.TransactionSort() // abort掉了一部分交易
		harmony := Harmony.NewHarmony(block.GetTransactions())
		harmony.TransactionSort()
		//fmt.Println(harmony.getAbortRate())
		for address, stateSet := range harmony.GetACG() {
			writeSet := stateSet.WriteSet
			if len(writeSet) == 0 {
				continue
			}
			flag := false
			for i := len(writeSet) - 1; i >= 0; i-- {
				if !writeSet[i].CheckTransactionAbort() {
					e.buffer[address], _ = strconv.Atoi(writeSet[i].GetTransactionWriteResult())
					flag = true
					break
				}
			}
			// 如果所有写集都被abort了，那么buffer内容要清空
			if !flag {
				writeSet = make([]Utils.Unit, 0)
				stateSet.WriteSet = writeSet
				delete(e.buffer, address)
			}
		}
		e.acgs = append(e.acgs, harmony.GetACG())
	}
	return e.acgs
}
