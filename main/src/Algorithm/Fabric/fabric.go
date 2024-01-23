package Fabric

import (
	"main/src/Blockchain"
)

type Transaction = Blockchain.Transaction
type Fabric struct {
	rate float64
	txs  []*Transaction
}

func NewFabric(txs []*Transaction) *Fabric {
	fabric := new(Fabric)
	fabric.rate = 0
	fabric.txs = txs
	return fabric
}
func (fabric *Fabric) TransactionSort() {
	ReadKeys := make(map[string]bool, 0)
	WriteKeys := make(map[string]bool, 0)
	for _, tx := range fabric.txs {
		tmpReadKeys := make(map[string]bool, 0)
		tmpWriteKeys := make(map[string]bool, 0)
		for _, op := range tx.Ops {
			switch op.Type {
			case Blockchain.OpRead:
				_, exist := tmpReadKeys[op.Key]
				if !exist {
					tmpReadKeys[op.Key] = true
				}
			case Blockchain.OpWrite:
				_, exist := tmpWriteKeys[op.Key]
				if !exist {
					tmpWriteKeys[op.Key] = true
				}
			}
		}
		for readkey, _ := range tmpReadKeys {
			_, exist := WriteKeys[readkey]
			if exist {
				tx.SetAbort()
			} else {
				ReadKeys[readkey] = true
			}
		}
		if tx.CheckAbort() {
			continue
		}
		for writekey, _ := range tmpWriteKeys {
			WriteKeys[writekey] = true
		}
	}
}

func (fabric *Fabric) getAbortRate() float64 {
	abort := 0
	for _, tx := range fabric.txs {
		if tx.CheckAbort() {
			abort += 1
		}
	}
	fabric.rate = float64(abort) / float64(len(fabric.txs))
	return fabric.rate
}
