package Consensus

import (
	"fmt"
	"main/src/Execution"
	"strconv"
	"sync"
)

// Peer Peer专门用于实验三，一个Peer启动n个instance，每个instance根据处理的交易数划分线程数，每个instance采用default的运行方式
type Peer struct {
	instances        []*Instance // 一个节点维护n个共识instance
	totalConcurrency int         // 总的并发数
}

func (p *Peer) GetInstanceNumber() int {
	return len(p.instances)
}
func (p *Peer) SetConcurrency(c int) {
	p.totalConcurrency = c
}
func (p *Peer) AddInstance(n int, concurrency int, skew float64, mode InstanceMode) {
	generator := Execution.NewGenerator([]float64{skew})
	transactions := generator.GenerateTransactions(n)[0]
	instance := NewInstance()
	instance.SetMode(mode)
	instance.SetConcurrency(concurrency)
	instance.InjectTransactions(transactions)
	p.instances = append(p.instances, instance)
}
func NewPeer(instanceNumber int, totalConcurrency int, transactionNumber int, skew float64) *Peer {
	p := new(Peer)
	p.SetConcurrency(totalConcurrency)
	eachInstanceTransactionNumber := DivideTransactions(instanceNumber, transactionNumber)
	eachInstanceConcurrency := GetConcurrencyShare(totalConcurrency, eachInstanceTransactionNumber)
	for i := 0; i < instanceNumber; i++ {
		p.AddInstance(eachInstanceTransactionNumber[i], eachInstanceConcurrency[i], skew, Default)
	}
	p.InstanceConcurrencyLog()
	return p
}
func (p *Peer) InstanceConcurrencyLog() {
	for i, instance := range p.instances {
		fmt.Println("Instance " + strconv.Itoa(i) + " Concurrency: " + strconv.Itoa(instance.GetConcurrency()))
	}
}
func (p *Peer) Run() (reports []*InstanceReport) {
	var wg4instance sync.WaitGroup
	wg4instance.Add(p.GetInstanceNumber())
	for _, instance := range p.instances {
		tmp := instance
		go func(instance *Instance) {
			defer wg4instance.Done()
			instance.Run()
		}(tmp)
	}
	wg4instance.Wait()
	for _, instances := range p.instances {
		reports = append(reports, instances.GetReport())
	}
	return reports
}
