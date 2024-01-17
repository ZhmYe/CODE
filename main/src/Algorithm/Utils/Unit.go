package Utils

import "main/src/Blockchain"

type Op = Blockchain.Op
type Transaction = Blockchain.Transaction

// Unit 操作单元，每一笔交易的某一个read/write操作
type Unit struct {
	op Op           // 实际执行的操作
	tx *Transaction // 交易标识
}

func newUnit(op Op, tx *Transaction) *Unit {
	unit := new(Unit)
	unit.op = op
	unit.tx = tx
	return unit
}
func (u *Unit) CheckTransactionAbort() bool {
	return u.tx.CheckAbort()
}
func (u *Unit) GetTransactionId() int {
	return u.tx.GetId()
}
func (u *Unit) GetTransactionOps() []*Op {
	return u.tx.GetOps()
}
func (u *Unit) SetTransactionAbort() {
	u.tx.SetAbort()
}
func (u *Unit) GetTransactionSequence() int {
	return u.tx.GetSequence()
}
func (u *Unit) SetTransactionSequence(s int) {
	u.tx.SetSequence(s)
}
func (u *Unit) GetTransactionHash() string {
	return u.tx.GetTxHash()
}
func (u *Unit) GetTransactionWriteResult() string {
	return u.op.WriteResult
}
