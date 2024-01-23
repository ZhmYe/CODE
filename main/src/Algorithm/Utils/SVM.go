package Utils

import (
	"main/src/Blockchain"
	"main/src/Smallbank"
	"sync"
)

type SVMState struct {
	version       int // 记录当前的版本号
	txs           []*Transaction
	lastWrite     string // 最后一个写操作的结果
	lastWriteFlag bool
	lastWriteTxId int // 最后一个写操作的交易id
	key           string
}
type CheckVersionResult int

const (
	Older CheckVersionResult = iota
	Latest
	HasBeenAbort
)

func NewSVMState(key string) *SVMState {
	s := new(SVMState)
	s.version = -1
	s.txs = make([]*Transaction, 0)
	s.key = key
	s.lastWrite = ""
	s.lastWriteTxId = -1
	s.lastWriteFlag = false
	return s
}
func (s *SVMState) CheckVersion(v int) CheckVersionResult {
	if s.version < v {
		return HasBeenAbort
	} else if s.version > v {
		return Older
	} else {
		return Latest
	}
}
func (s *SVMState) GetVersion() int {
	return s.version
}
func (s *SVMState) Append(tx *Transaction, version int) {
	s.txs = append(s.txs, tx)
	s.version = version
}

// UpdateVersion 如果某个交易在一个address上被abort，会影响到其他的address
// 此时这些address已被锁住，修改其最新的版本号，即最新有效的交易id
func (s *SVMState) UpdateVersion() {
	version := -1
	lastWrite := ""
	lastWriteTxId := -1
	s.lastWriteFlag = false
	for _, tx := range s.txs {
		if !tx.CheckAbort() {
			version = tx.GetId()
			// 同时需要更新最后的写操作结果和id
			for _, op := range tx.GetOps() {
				if op.Key == s.key && op.Type == Blockchain.OpWrite {
					lastWrite = op.WriteResult
					lastWriteTxId = tx.GetId()
					s.lastWriteFlag = true
				}
			}
		}
	}
	s.version = version
	s.lastWriteTxId = lastWriteTxId
	s.lastWrite = lastWrite
}
func (s *SVMState) GetValue() (string, int) {
	if s.lastWriteFlag {
		return s.lastWrite, s.version
	} else {
		return Smallbank.GlobalSmallBank.Read(s.key), s.version
	}
}

type TryAppendResult int

const (
	AppendDirectly = iota
	AppendWithAbort
	Abort
)

func (s *SVMState) TryAppend(tx *Transaction, version int) TryAppendResult {
	checkResult := s.CheckVersion(version)
	switch checkResult {
	case Latest:
		// 是最新的版本
		return AppendDirectly
	case Older:
		// 当前交易读取的版本比最新版本老
		// 说明有同时并发的交易已经修改了当前的版本
		// case1
		// 如果当前交易是序号靠前的交易，那么为了保证确定性，需要将链表中所有序号大于当前交易的交易全部abort
		// 对比的应该是最后一笔对该地址有写入操作的有效交易，即lastWriteTxId
		if tx.GetId() < s.lastWriteTxId {
			return AppendWithAbort
		} else {
			// case2
			// 如果当前交易是序号靠后的交易，那么该交易直接abort
			// 在外面通过是否有false判断是否abort，如果有直接abort
			return Abort
		}
	case HasBeenAbort:
		return Abort
	}
	return AppendDirectly
}
func (s *SVMState) ProcessTransaction(tx *Transaction, checkResult TryAppendResult) {
	switch checkResult {
	case AppendDirectly:
		// 是最新的版本，那么直接交易添加到链表中，并更新版本号
		s.Append(tx, tx.GetId()) // 将交易id作为版本号，交易id应该是单调递增的
	case AppendWithAbort:
		// 当前交易读取的版本比最新版本老
		// 说明有同时并发的交易已经修改了当前的版本
		// 如果当前交易是序号靠前的交易，那么为了保证确定性，需要将链表中所有序号大于当前交易的交易全部abort
		// 对比的应该是最后一笔对该地址有写入操作的有效交易，即lastWriteTxId
		// 将序号比当前交易大的交易全部abort，倒序遍历
		for i := len(s.txs) - 1; i >= 0; i-- {
			if s.txs[i].GetId() > tx.GetId() {
				s.txs[i].SetAbort()
			} else {
				// 本身已保证有序
				break
			}
		}
		s.Append(tx, tx.GetId())
	default:
		panic("Invalid Params")
	}
}

type SVM struct {
	Address2Transactions map[string]*SVMState
	Mutexes              map[string]*sync.Mutex
}

func NewSVM(keys []string) *SVM {
	svm := new(SVM)
	svm.Mutexes = make(map[string]*sync.Mutex)
	svm.Address2Transactions = make(map[string]*SVMState)
	for _, key := range keys {
		svm.Mutexes[key] = new(sync.Mutex)
		svm.Address2Transactions[key] = NewSVMState(key)
	}
	return svm
}
func (s *SVM) Lock(keys []string) {
	for {
		successLock := make([]string, 0)
		flag := true
		for _, key := range keys {
			if !s.tryLock(key) {
				flag = false
				break
			}
			successLock = append(successLock, key)
		}
		if !flag {
			s.UnLock(successLock)
			tmp := 0
			for k := 0; k < 100000; k++ {
				tmp++
			}
		} else {
			break
		}
	}
}
func (s *SVM) UnLock(keys []string) {
	for _, key := range keys {
		s.Mutexes[key].Unlock()
	}
}

func (s *SVM) tryLock(key string) bool {
	return s.Mutexes[key].TryLock()
}
func (s *SVM) Read(key string) (string, int) {
	return s.Address2Transactions[key].GetValue()
}
func (s *SVM) Commit(tx *Transaction, versions map[string]int) {
	keys := make([]string, 0)
	for address, _ := range versions {
		keys = append(keys, address)
	}
	s.Lock(keys)
	defer s.UnLock(keys)
	tryAppendResult := make(map[string]TryAppendResult, 0)
	// 先确认好这笔交易是否会被abort，得到所有地址的处理结果后
	for _, address := range keys {
		version := versions[address]
		result := s.Address2Transactions[address].TryAppend(tx, version)
		if result == Abort {
			tx.SetAbort()
			return
		} else {
			tryAppendResult[address] = result
		}
	}
	// 如果这笔交易不会被abort，那么真正添加这笔交易
	// 这里主要考虑Older情况中的abort，如果因为添加了这笔交易导致其它交易被abort
	// 然后最终这笔交易又被abort，会导致其它交易白白被abort且无法检测
	for address, result := range tryAppendResult {
		s.Address2Transactions[address].ProcessTransaction(tx, result)
	}

}
func (s *SVM) CommitLastWriteToDB() {
	for address, svmValue := range s.Address2Transactions {
		if svmValue.lastWriteFlag {
			Smallbank.GlobalSmallBank.Write(address, svmValue.lastWrite)
		}
	}
}
