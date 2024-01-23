package Consensus

import (
	"fmt"
	"main/src/Execution"
	"main/src/Sys"
	"testing"
)

func Test4Peer(t *testing.T) {
	peer := NewPeer(4, 64, 10000, 0.6)
	peer.Run()
}
func Test4Instance(t *testing.T) {
	Sys.SetCPU(8)
	generator := Execution.NewGenerator([]float64{0.6})
	transactions := generator.GenerateTransactions(1000)[0]
	instance := NewInstance()
	instance.SetMode(PreBatched)
	instance.SetConcurrency(64)
	instance.InjectTransactions(transactions)
	instance.Run()
	report := instance.GetReport()
	fmt.Println(report.GetConcurrency(), report.GetProcessTime(), report.GetProcessTransactionNumber(), report.GetAbortRate())
}
