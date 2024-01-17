package Nezha

import (
	"main/src/Algorithm/Utils"
	"main/src/Blockchain"
)

type Unit = Utils.Unit
type StateSet = Utils.StateSet
type Transaction = Blockchain.Transaction

// NeZha Nezha实例
type NeZha struct {
	acg   Utils.ACG
	rate  float64
	txs   []*Transaction
	order []int
}

func NewNeZha(txs []*Transaction) *NeZha {
	nezha := new(NeZha)
	nezha.rate = 0
	nezha.txs = txs
	nezha.acg = *new(Utils.ACG)
	nezha.order = make([]int, 0)
	return nezha
}
func (nezha *NeZha) getACG() {
	nezha.acg = Utils.GetACG(nezha.txs)
}
func (nezha *NeZha) getAbortRate() float64 {
	abort := 0
	for _, tx := range nezha.txs {
		if tx.CheckAbort() {
			abort += 1
		}
	}
	nezha.rate = float64(abort) / float64(len(nezha.txs))
	return nezha.rate
}

// Transaction Sort
func getMinSeq(sortedRSet []Unit) int {
	minSeq := 100000000
	for _, unit := range sortedRSet {
		if unit.GetTransactionSequence() < minSeq {
			minSeq = unit.GetTransactionSequence()
		}
	}
	return minSeq
}
func getMaxSeq(sortedRSet []Unit) int {
	maxSeq := -1
	for _, unit := range sortedRSet {
		if unit.GetTransactionSequence() > maxSeq {
			maxSeq = unit.GetTransactionSequence()
		}
	}
	return maxSeq
}
func getSortedRSet(Rw StateSet) []Unit {
	sortedRSet := make([]Unit, 0)
	for _, unit := range Rw.ReadSet {
		if unit.GetTransactionSequence() != -1 {
			sortedRSet = append(sortedRSet, unit)
		}
	}
	return sortedRSet
}
func getSortedWSet(Rw StateSet) []Unit {
	sortedWSet := make([]Unit, 0)
	for _, unit := range Rw.WriteSet {
		if unit.GetTransactionSequence() != -1 {
			sortedWSet = append(sortedWSet, unit)
		}
	}
	return sortedWSet
}

// TransactionSort 利用ACG对交易进行排序
func (nezha *NeZha) TransactionSort() {
	nezha.getACG()
	initialSeq := 0
	// 这里加上对address的排序结果sorted_address，然后下面通过遍历sorted_address来得到
	for _, Rw := range nezha.acg {
		maxRead := -1
		writeSeq := -1
		sortedRSet := getSortedRSet(Rw)
		ReadSetTxHash := make(map[string]bool, 0) // 用于判断是否有同意交易的读写在同一个key上
		// line 4 - 15
		if len(sortedRSet) == 0 {
			for _, unit := range Rw.ReadSet {
				unit.SetTransactionSequence(initialSeq)
				sortedRSet = append(sortedRSet, unit)
				_, exist := ReadSetTxHash[unit.GetTransactionHash()]
				if !exist {
					ReadSetTxHash[unit.GetTransactionHash()] = true
				}
			}
			maxRead = initialSeq
		} else {
			minSeq := getMinSeq(sortedRSet)
			maxSeq := getMaxSeq(sortedRSet)
			maxRead = maxSeq
			for _, unit := range Rw.ReadSet {
				if unit.GetTransactionSequence() == -1 {
					unit.SetTransactionSequence(minSeq)
					sortedRSet = append(sortedRSet, unit)
				}
				_, exist := ReadSetTxHash[unit.GetTransactionHash()]
				if !exist {
					ReadSetTxHash[unit.GetTransactionHash()] = true
				}
			}
		}
		// line 16 - 19
		sortedWSet := getSortedWSet(Rw)
		for _, unit := range sortedWSet {
			_, exist := ReadSetTxHash[unit.GetTransactionHash()]
			if exist {
				unit.SetTransactionSequence(maxRead + 1)
				maxRead += 1
			}
		}
		// line 20 - 24
		for _, unit := range sortedWSet {
			if unit.GetTransactionSequence() < maxRead {
				unit.SetTransactionAbort()
			}
		}
		// line 25 - 29
		if len(Rw.ReadSet) == 0 {
			writeSeq = initialSeq
		} else {
			writeSeq = maxRead + 1
		}
		// line 30 - 35
		for _, unit := range Rw.WriteSet {
			if unit.GetTransactionSequence() == -1 {
				unit.SetTransactionSequence(writeSeq)
				writeSeq += 1
			}
		}
	}
	nezha.Sort()
}
func (nezha *NeZha) Sort() {
	seq := 0
	flag := false
	for {
		for i, tx := range nezha.txs {
			if tx.CheckAbort() {
				continue
			}
			if tx.GetSequence() == seq {
				flag = true
				nezha.order = append(nezha.order, i)
			}
		}
		if !flag {
			break
		}
		flag = false
		seq++
	}
}
