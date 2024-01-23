package Consensus

import (
	"main/src/Algorithm/Nezha"
	"main/src/Blockchain"
	"main/src/Execution"
	"sync"
	"time"
)

type Block = Blockchain.Block
type Transaction = Blockchain.Transaction
type InstanceReport struct {
	processTransactionNumber int           // 处理了多少笔交易
	processTime              time.Duration // 处理的交易
	concurrency              int           // 线程数
	abortRate                float64       // abort率
}

func NewInstanceReport() *InstanceReport {
	report := new(InstanceReport)
	report.concurrency = -1
	report.processTime = time.Duration(0)
	report.processTransactionNumber = 0
	report.abortRate = 0
	return report
}
func (r *InstanceReport) SetProcessTime(t time.Duration) {
	r.processTime = t
}
func (r *InstanceReport) SetProcessTransactionNumber(n int) {
	r.processTransactionNumber = n
}
func (r *InstanceReport) SetAbortRate(rate float64) {
	r.abortRate = rate
}
func (r *InstanceReport) GetProcessTime() time.Duration {
	return r.processTime
}
func (r *InstanceReport) GetProcessTransactionNumber() int {
	return r.processTransactionNumber
}
func (r *InstanceReport) GetAbortRate() float64 {
	return r.abortRate
}
func (r *InstanceReport) GetConcurrency() int {
	return r.concurrency
}
func (r *InstanceReport) SetConcurrency(n int) {
	r.concurrency = n
}

type InstanceMode int

const (
	Default    InstanceMode = iota // 所有线程同时处理一批交易，然后批与批之间串行
	PreBatched                     // 将所有交易按序分配为线程数个batch,batch内部串行，batch之间并行，后续需合并
	Preemptive                     // 抢占式，每个线程执行完自己的就去主线程(channel)处获得下一个交易
)

type Instance struct {
	transactions []*Transaction  // 这里直接用交易代替一个或多个区块
	block        []*Block        // 每次共识时的一个或多个区块，目前暂时不用，用交易简单替代
	mode         InstanceMode    // 执行交易的方式
	concurrency  int             // 并发数，分配到的线程数
	report       *InstanceReport // 用于最后输出统计数据
}

func NewInstance() *Instance {
	instance := new(Instance)
	instance.mode = Default
	instance.block = make([]*Block, 0)
	instance.transactions = make([]*Transaction, 0)
	instance.concurrency = 1
	instance.report = NewInstanceReport()
	return instance
}
func (i *Instance) SetConcurrency(c int) {
	i.concurrency = c
}
func (i *Instance) SetMode(mode InstanceMode) {
	i.mode = mode
}
func (i *Instance) InjectTransactions(transactions []*Transaction) {
	i.transactions = transactions
}
func (i *Instance) GetConcurrency() int {
	return i.concurrency
}
func (i *Instance) GetReport() *InstanceReport {
	return i.report
}
func (i *Instance) SetReport(processTransactionNumber int, processTime time.Duration, rate float64) {
	i.report.SetProcessTransactionNumber(processTransactionNumber)
	i.report.SetProcessTime(processTime)
	i.report.SetAbortRate(rate)
	i.report.SetConcurrency(i.concurrency)
}
func (i *Instance) GetProcessTimeString() string {
	return i.report.GetProcessTime().String()
}
func (i *Instance) GetProcessTransactionNumber() int {
	return i.report.GetProcessTransactionNumber()
}
func (i *Instance) GetAbortRate() float64 {
	return i.report.GetAbortRate()
}
func (i *Instance) RunInDefault() {
	// 每个instance获取到了一批交易和一定数量的线程数
	// 直接使用executor
	executor := Execution.NewExecutor(Execution.ExecuteWithFabric, i.concurrency, i.transactions)
	executor.SplitTransactions()
	finalTimeDuration := time.Duration(0)
	finalAbortRate := float64(0)
	for k := 0; k < 100; k++ {
		// 只需要执行时间， 在SimpleExecute里已经把每次重复执行的交易状态重置了
		timeDuration, abortRate := executor.SimpleExecute()
		finalTimeDuration += timeDuration
		finalAbortRate += abortRate
	}
	finalTimeDuration /= 100
	finalAbortRate /= 100
	i.SetReport(len(i.transactions), finalTimeDuration, finalAbortRate)
}
func (i *Instance) RunInPreBatched() {
	// instance将交易按序划分为线程数批,每个线程处理对应的批
	// 每个线程内部串行执行，然后合并，需要按序abort,类似fabric，但不能直接使用Fabric类(或将一批交易看成一笔交易？)
	finalTimeDuration := time.Duration(0)
	finalAbortRate := float64(0)
	batch := BatchTransactions(i.transactions, i.concurrency)
	for k := 0; k < 100; k++ {
		startTime := time.Now()
		var wg4batch sync.WaitGroup
		wg4batch.Add(i.concurrency)
		for b := 0; b < i.concurrency; b++ {
			tmpBatchItem := batch[b]
			go func(batchTransctions []*Transaction) {
				defer wg4batch.Done()
				// 串行相当于并发度为1
				executor := Execution.NewExecutor(Execution.ExecuteWithFabric, 1, batchTransctions)
				executor.SplitTransactions()
				executor.SimpleExecute()
			}(tmpBatchItem)
		}
		wg4batch.Wait()
		// 处理abort
		// ReadSetInBeforeBatch := make(map[string]bool, 0)
		WriteSetInBeforeBatch := make(map[string]bool, 0)
		for _, batchItem := range batch {
			batchWriteSet := make(map[string]bool, 0)
			abortWriteSet := make(map[string]bool, 0)
			for _, tx := range batchItem {
				localWriteSet := make([]string, 0)
				for _, op := range tx.GetOps() {
					// 读操作需要判断在之前的Batch中是否已经有新的写入
					if op.Type == Blockchain.OpRead {
						_, hasBeenWrite := WriteSetInBeforeBatch[op.Key]
						if hasBeenWrite {
							tx.SetAbort()
						}
						// 如果其读的地址，基于当前batch中前面的写，但前面的写已经被abort了，那么会出现级联abort
						_, hasBeenAbort := abortWriteSet[op.Key]
						if hasBeenAbort {
							tx.SetAbort()
						}
					} else {
						// 要把交易的写集全部记录下来，用于级联abort
						localWriteSet = append(localWriteSet, op.Key)
					}
				}
				if !tx.CheckAbort() {
					// 如果交易没有被abort，那么把其写集加入到batchWriteSet中
					for _, address := range localWriteSet {
						batchWriteSet[address] = true
					}
				} else {
					// 如果该交易被abort了，俺么把其写集加入到abortWriteSet中，用于后续处理级联
					for _, address := range localWriteSet {
						abortWriteSet[address] = true
					}
				}
			}
			// 一个Batch内的交易全部处理完成，得到有效的写集batchWriteSet，将其添加到WriteSetInBeforeBatch用于后续batch处理
			for address, _ := range batchWriteSet {
				WriteSetInBeforeBatch[address] = true
			}
		}
		finalTimeDuration += time.Since(startTime)
		abortNumber := 0
		for _, tx := range i.transactions {
			if tx.CheckAbort() {
				abortNumber += 1
			}
			tx.Reset()
		}
		finalAbortRate += float64(abortNumber) / float64(len(i.transactions))
	}
	finalTimeDuration /= 100
	finalAbortRate /= 100
	i.SetReport(len(i.transactions), finalTimeDuration, finalAbortRate)
}
func worker(tasks <-chan *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		executor := Execution.NewExecutor(Execution.ExecuteWithFabric, 1, []*Transaction{task})
		executor.SplitTransactions()
		executor.SimpleExecute()
	}
}
func (i *Instance) RunInPreemptive() {
	// 一共concurrency个线程，每个线程通过channel抢占式的得到下一个tx，所有tx全部在线程内乐观并发
	// 执行全部结束后，进行abort
	finalTimeDuration := time.Duration(0)
	finalAbortRate := float64(0)
	for epoch := 0; epoch < 100; epoch++ {
		startTime := time.Now()
		numTask := len(i.transactions)
		tasks := make(chan *Transaction, numTask)
		var wg4worker sync.WaitGroup
		wg4worker.Add(i.concurrency)
		for x := 0; x < i.concurrency; x++ {
			go worker(tasks, &wg4worker)
		}
		for _, transaction := range i.transactions {
			tasks <- transaction
		}
		close(tasks)
		wg4worker.Wait()
		nezha := Nezha.NewNeZha(i.transactions)
		nezha.TransactionSort()
		abortRate := nezha.GetAbortRate()
		finalAbortRate += abortRate
		finalTimeDuration += time.Since(startTime)
	}
	finalTimeDuration /= 100
	finalAbortRate /= 100
	i.SetReport(len(i.transactions), finalTimeDuration, finalAbortRate)
}
func (i *Instance) Run() {
	switch i.mode {
	case Default:
		i.RunInDefault()
	case PreBatched:
		i.RunInPreBatched()
	case Preemptive:
		i.RunInPreemptive()
	}
}
