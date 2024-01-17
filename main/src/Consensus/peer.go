package Consensus

import (
	"main/src/Algorithm/Utils"
	"main/src/Blockchain"
	"strconv"
	"sync"
	"time"
)

type Peer struct {
	instances         []Instance    // 一个节点维护n个共识instance
	instanceNumber    int           // instance数量
	timeDuration      time.Duration // 执行timeout
	lastExecutionTime time.Time     // 上一次运行执行的时间
}

func NewPeer(n int) *Peer {
	peer := new(Peer)
	peer.instanceNumber = n
	peer.instances = make([]Instance, 0)
	peer.timeDuration = time.Duration(40) * time.Millisecond
	peer.UpdateLastExecutionTime()
	for i := 0; i < n; i++ {
		instance := NewInstance(time.Duration(int(float64(40)/float64(i+1))), i)
		peer.instances = append(peer.instances, *instance)
	}
	return peer
}
func (peer *Peer) UpdateLastExecutionTime() {
	peer.lastExecutionTime = time.Now()
}
func (peer *Peer) checkExecutionTimeout() bool {
	if time.Since(peer.lastExecutionTime) >= peer.timeDuration {
		return true
	}
	return false
}
func (peer *Peer) checkFinished() bool {
	total := 0
	for _, instance := range peer.instances {
		if instance.finish && len(instance.blocks) == instance.hasExecutedIndex {
			total += 1
		}
	}
	if total == peer.instanceNumber {
		return true
	}
	return false
}
func (peer *Peer) Run() (time.Duration, float64) {
	// 启动所有instance
	for i := 0; i < len(peer.instances); i++ {
		peer.instances[i].Start()
	}
	peer.UpdateLastExecutionTime()
	startTime := time.Now()
	abortTxNumber := 0
	for {
		if peer.checkFinished() {
			//fmt.Println("四个instance全部结束")
			break
		}
		if peer.checkExecutionTimeout() {
			peer.UpdateLastExecutionTime()
			var wg4Execution sync.WaitGroup
			wg4Execution.Add(len(peer.instances))
			// 所有instance模拟执行自己的速度对应的个数，这里也就是instance的id=i+1
			execBlockNumber := make([]int, peer.instanceNumber)
			for i := 0; i < len(peer.instances); i++ {
				go func(index int, wg *sync.WaitGroup) {
					defer wg.Done()
					n := peer.instances[index].SimulateExecution(index + 1)
					execBlockNumber[index] = n
				}(i, &wg4Execution)
			}
			wg4Execution.Wait()
			cascade := make(map[string][]int)
			for _, address := range globalSmallBank.Savings {
				//cascade[address] = make([]int, peer.instanceNumber)
				tmp := make([]int, peer.instanceNumber)
				flag := false
				for i := 0; i < len(peer.instances); i++ {
					localCascade, exist := peer.instances[i].cascade[address]
					if exist {
						if localCascade != 0 {
							flag = true
						}
						//flag = true
						tmp[i] = localCascade
					} else {
						tmp[i] = 0
					}
				}
				if flag {
					cascade[address] = tmp
				}
			}
			for _, address := range globalSmallBank.Checkings {
				//cascade[address] = make([]int, peer.instanceNumber)
				tmp := make([]int, peer.instanceNumber)
				flag := false
				for i := 0; i < len(peer.instances); i++ {
					localCascade, exist := peer.instances[i].cascade[address]
					if exist {
						if localCascade != 0 {
							flag = true
						}
						//flag = true
						tmp[i] = localCascade
					} else {
						tmp[i] = 0
					}
				}
				if flag {
					cascade[address] = tmp
				}
			}
			// 到此到此了每个地址上各个instance的级联度
			// 这里可以对各个地址进行排序，现在先默认不排序
			OrderGraph := make([][]int, peer.instanceNumber)
			for i := 0; i < peer.instanceNumber; i++ {
				OrderGraph[i] = make([]int, peer.instanceNumber)
				for j := 0; j < peer.instanceNumber; j++ {
					OrderGraph[i][j] = 0
				}
			}
			for _, cascades := range cascade {
				// 先对cascades进行排序
				index := make([]int, peer.instanceNumber)
				for i := 0; i < peer.instanceNumber; i++ {
					index[i] = i
				}
				for i := 0; i < peer.instanceNumber; i++ {
					for j := i + 1; j < peer.instanceNumber; j++ {
						if cascades[i] < cascades[j] {
							index[i], index[j] = index[j], index[i]
						}
					}
				}
				// 排序后index内对应从大到小的cascade的下标
				for i := 0; i < peer.instanceNumber-1; i++ {
					pre := index[i]
					latter := index[i+1]
					// 之前没有排序过
					if OrderGraph[latter][pre] == 0 {
						OrderGraph[pre][latter] = 1
						// 为了防止成环
						for j := 0; j < peer.instanceNumber; j++ {
							if OrderGraph[j][pre] == 1 {
								OrderGraph[j][latter] = 1
							}
						}
					}
				}
			}
			order := Utils.TopologicalOrder(OrderGraph)
			abortTxs := make([]*Transaction, 0)
			writeAddress := make(map[string]bool, 0)
			for _, index := range order {
				peer.instances[index].CascadeAbort(&writeAddress)
				txs := peer.instances[index].execute(execBlockNumber[index])
				abortTxs = append(abortTxs, txs...)
			}
			abortTxNumber += len(abortTxs)
			//peer.reExecute(abortTxs)
		}
	}
	RunTime := time.Since(startTime)
	totalTxNumber := 0
	for _, instance := range peer.instances {
		for _, block := range instance.blocks {
			if block.CheckFinish() {
				totalTxNumber += block.GetTransactionLength()
			}
		}
	}
	return RunTime, float64(abortTxNumber) / float64(totalTxNumber)
}

func (peer *Peer) reExecute(txs []*Transaction) {
	for _, tx := range txs {
		for _, op := range tx.Ops {
			if op.Type == Blockchain.OpRead {
				op.ReadResult = globalSmallBank.Read(op.Key)
			} else {
				readResult, _ := strconv.Atoi(globalSmallBank.Read(op.Key))
				amount, _ := strconv.Atoi(op.Val)
				op.WriteResult = strconv.Itoa(readResult + amount)
				globalSmallBank.Write(op.Key, op.WriteResult)
			}
		}
	}
}
