package Blockchain

// OpType 操作类型
type OpType int

const (
	OpRead  OpType = iota // 读操作
	OpWrite               // 写操作
)

// Op 操作
type Op struct {
	Type        OpType // 操作类型 读/写
	Key         string // 操作的key
	Val         string // 操作的value
	ReadResult  string // 最终读到的结果
	WriteResult string // 最终要写的结果
}

// TxType 交易类型, smallbank
type TxType int

const (
	TransactSavings TxType = iota
	DepositChecking
	SendPayment
	WriteCheck
	Amalgamate
	Query
)

// Transaction 交易
type Transaction struct {
	txType   TxType // 交易类型
	Ops      []*Op  // 交易中包含的操作
	abort    bool   // 是否abort
	sequence int    // sorting时的序列号
	txHash   string // 交易哈希
	id       int    // 交易id
}

func (t *Transaction) SetAbort() {
	t.abort = true
}

func (t *Transaction) CheckAbort() bool {
	return t.abort
}
func (t *Transaction) GetId() int {
	return t.id
}
func (t *Transaction) SetId(id int) {
	t.id = id
}
func (t *Transaction) GetOps() []*Op {
	return t.Ops
}
func (t *Transaction) GetSequence() int {
	return t.sequence
}
func (t *Transaction) SetSequence(s int) {
	t.sequence = s
}
func (t *Transaction) GetTxHash() string {
	return t.txHash
}
func NewTransaction(ops []*Op, sequence int, hash string, txType TxType) *Transaction {
	t := new(Transaction)
	t.Ops = ops
	t.txHash = hash
	t.sequence = sequence
	t.txType = txType
	t.abort = false
	return t
}
func (t *Transaction) Reset() {
	t.abort = false
	t.sequence = -1
}
