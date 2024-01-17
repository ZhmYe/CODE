package FabricPP

import (
	"main/src/Algorithm/Utils"
	"main/src/Blockchain"
)

type Transaction = Blockchain.Transaction
type FabricPP struct {
	rate  float64
	txs   []*Transaction
	cg    Utils.CG
	order []int
}

func NewFabricPP(txs []*Transaction) *FabricPP {
	fabricPp := new(FabricPP)
	fabricPp.rate = 0
	fabricPp.txs = txs
	return fabricPp
}

func (f *FabricPP) TransactionSort() {
	f.cg = *Utils.NewCG(f.txs)
	f.cg.GetSubGraph()
	//fmt.Println(len(f.cg.subGraph))
	f.cg.GetAllCycles()
	//fmt.Println(len(f.cg.cycles))
	f.cg.TransactionAbort()
	f.DAGSort()
}

func (f *FabricPP) DAGSort() {
	f.order = Utils.TopologicalOrder(f.cg.GetGraph())
	//fmt.Println(len(f.order))
}
func (f *FabricPP) getAbortRate() float64 {
	abort := 0
	for _, tx := range f.txs {
		if tx.CheckAbort() {
			abort += 1
		}
	}
	f.rate = float64(abort) / float64(len(f.txs))
	return f.rate
}
