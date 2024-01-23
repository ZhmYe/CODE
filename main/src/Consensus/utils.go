package Consensus

// 这里将总数为total的交易，划分为随机的share份，返回int[share]{每份大小}
func DivideTransactions(share int, total int) []int {
	// 后续还需要将交易比例转化为并发数，在这里直接简化，提前规定好交易的比例
	// 先不管share数不为4的
	if share != 4 {
		return []int{}
	} else {
		// 规定交易比例为1.5: 1 : 1 : 0.5
		// 即一共8份，分为3: 2 : 2 : 1
		each := total / 8
		return []int{total - 5*each, 2 * each, 2 * each, each}
	}
}
func GetConcurrencyShare(totalConcurrency int, transactionNumbers []int) []int {
	// 暂时也简化，直接返回
	// 规定比例为3 : 2 : 2 : 1
	each := totalConcurrency / 8
	return []int{totalConcurrency - each*5, each * 2, each * 2, each}
}
func BatchTransactions(transactions []*Transaction, BatchNumber int) (batch [][]*Transaction) {
	batchSize := len(transactions) / BatchNumber // 这里暂时先不考虑不能整除还有剩余交易的情况
	for i := 0; i < BatchNumber; i++ {
		batch = append(batch, transactions[i*batchSize:(i+1)*batchSize])
	}
	return batch
}
