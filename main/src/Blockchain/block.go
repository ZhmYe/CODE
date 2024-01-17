package Blockchain

import "time"

// Block 区块
type Block struct {
	txs         []*Transaction // 交易
	createTime  time.Time      // 被创建的时间，用于衡量等待时间
	finish      bool
	finishTime  time.Time     // 结束的时间
	processTime time.Duration // 处理时间
}

func NewBlock(txs []*Transaction) *Block {
	block := new(Block)
	block.txs = txs
	block.createTime = time.Now()
	block.finish = false
	return block
}
func (b *Block) SignToFinish() {
	b.finish = true
	b.finishTime = time.Now()
	b.processTime = time.Since(b.createTime)
}
func (b *Block) getProcessTime() time.Duration {
	return b.processTime
}
func (b *Block) GetTransactions() []*Transaction {
	return b.txs
}
func (b *Block) GetTransactionLength() int {
	return len(b.txs)
}
func (b *Block) GetTransaction(index int) *Transaction {
	return b.txs[index]
}
func (b *Block) CheckFinish() bool {
	return b.finish
}
