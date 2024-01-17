package Consensus

import (
	"main/src/Blockchain"
)

// baseline
func (peer *Peer) runInBaseline() {
	// 启动所有instance
	for i := 0; i < len(peer.instances); i++ {
		peer.instances[i].Start()
	}
	peer.UpdateLastExecutionTime()
	index := 0
	for {
		if peer.checkFinished() {
			//fmt.Println("四个instance全部结束")
			break
		}
		if peer.baselineCheck(index+1) && peer.checkExecutionTimeout() {
			peer.UpdateLastExecutionTime()
			// 开始新一轮执行
			//fmt.Println("开始新一轮执行...")
			//for _, instance := range peer.instances {
			//	fmt.Print(len(instance.blocks) - instance.hasExecutedIndex)
			//	fmt.Print(" ")
			//}
			//fmt.Println()
			//startTime := time.Now()
			// 取出所有要执行的交易
			txs := make([]*Transaction, 0)
			for _, instance := range peer.instances {
				txs = append(txs, instance.blocks[index].GetTransactions()...)
			}
			blocks := make([]*Block, 0)
			blocks = append(blocks, Blockchain.NewBlock(txs))
			simulateEngine := newSimulateEngine(blocks)
			simulateEngine.SimulateExecution()
			abortTxs := make([]*Transaction, 0)
			//writeAddress := make(map[string]bool, 0)
			for i, _ := range peer.instances {
				//peer.instances[index].CascadeAbort(&writeAddress)
				tmp := peer.instances[i].execute(1)
				abortTxs = append(abortTxs, tmp...)
			}
			peer.reExecute(abortTxs)
			//finishTime := time.Now()
			for _, instance := range peer.instances {
				instance.blocks[index].SignToFinish()
				instance.hasExecutedIndex++
			}
			index++
		}

	}
}
func (peer *Peer) baselineCheck(index int) bool {
	total := 0
	for _, instance := range peer.instances {
		if len(instance.blocks) >= index {
			total += 1
		}
	}
	return total == peer.instanceNumber
}
