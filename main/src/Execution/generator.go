package Execution

import (
	"main/src/Blockchain"
	"main/src/Config"
	"main/src/Smallbank"
)

var config = Config.GlobalConfig

type Generator struct {
	smallbank *Smallbank.Smallbank
	skews     []float64
}

func NewGenerator(skews []float64) *Generator {
	g := new(Generator)
	g.skews = skews
	g.smallbank = Smallbank.GlobalSmallBank
	return g
}
func (g *Generator) GenerateTransactions(n int) (transactions [][]*Blockchain.Transaction) {
	for _, skew := range g.skews {
		g.smallbank.UpdateZipfianWithSkew(skew)
		transactions = append(transactions, g.smallbank.GenTxSet(n))
	}
	return transactions
}
