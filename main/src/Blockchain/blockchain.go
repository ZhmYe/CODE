package Blockchain

// BlockChain 区块链
// 在所有的instance产生的块处理完成后，按排序结果加入到一个大的Block中，这里暂时先留着不用写
type BlockChain struct {
	blocks []*Block
}
