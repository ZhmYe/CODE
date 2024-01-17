package Utils

import "main/src/Blockchain"

// ACG address->StateSet
type ACG = map[string]StateSet

// 构建并发交易所对应的ACG
func GetACG(txs []*Transaction) ACG {
	acg := make(ACG)
	for _, tx := range txs {
		for _, op := range tx.Ops {
			_, exist := acg[op.Key]

			// 如果在acg中不存在address,新建一个StateSet
			if !exist {
				acg[op.Key] = *newStateSet()
			}

			unit := newUnit(*op, tx) // 新建操作单元
			stateSet := acg[op.Key]

			// 根据读/写操作加入到StateSet的两部分中
			switch unit.op.Type {
			case Blockchain.OpRead:
				stateSet.appendToReadSet(*unit)
			case Blockchain.OpWrite:
				stateSet.appendToWriteSet(*unit)
			}
			acg[op.Key] = stateSet
		}
	}
	return acg
}
