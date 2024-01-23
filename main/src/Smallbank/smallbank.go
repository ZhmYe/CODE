package Smallbank

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/syndtr/goleveldb/leveldb"
	"main/src/Blockchain"
	"main/src/Config"
	"math/rand"
	"strconv"
	"time"
)

var config = Config.GlobalConfig

type Op = Blockchain.Op
type Transaction = Blockchain.Transaction
type Smallbank struct {
	Savings   []string
	Checkings []string
	txid      int
	db        *leveldb.DB
	zipfian   *Zipfian
	r         *rand.Rand
}

func (s *Smallbank) GetKeys() ([]string, []string) {
	return s.Savings, s.Checkings
}

// TransactSavings 向储蓄账户增加一定余额
func (s *Smallbank) TransactSavings(account string, amount int) *Transaction {
	r := Op{
		Type: Blockchain.OpRead,
		Key:  account,
	}
	w := Op{
		Type: Blockchain.OpWrite,
		Key:  account,
		Val:  strconv.Itoa(amount),
	}
	return Blockchain.NewTransaction([]*Op{&r, &w}, -1, strconv.Itoa(s.txid), Blockchain.TransactSavings)
}

// DepositChecking 向支票账户增加一定余额
func (s *Smallbank) DepositChecking(account string, amount int) *Transaction {
	r := Op{
		Type: Blockchain.OpRead,
		Key:  account,
	}
	w := Op{
		Type: Blockchain.OpWrite,
		Key:  account,
		Val:  strconv.Itoa(amount),
	}
	return Blockchain.NewTransaction([]*Op{&r, &w}, -1, strconv.Itoa(s.txid), Blockchain.DepositChecking)
}

// SendPayment 在两个支票账户间转账
func (s *Smallbank) SendPayment(accountA string, accountB string, amount int) *Transaction {
	ra := Op{
		Type: Blockchain.OpRead,
		Key:  accountA,
	}
	rb := Op{
		Type: Blockchain.OpRead,
		Key:  accountB,
	}
	wa := Op{
		Type: Blockchain.OpWrite,
		Key:  accountA,
		Val:  strconv.Itoa(-amount),
	}
	wb := Op{
		Type: Blockchain.OpWrite,
		Key:  accountB,
		Val:  strconv.Itoa(amount),
	}
	return Blockchain.NewTransaction([]*Op{&ra, &rb, &wa, &wb}, -1, strconv.Itoa(s.txid), Blockchain.SendPayment)
}

// WriteCheck 减少一个支票账户
func (s *Smallbank) WriteCheck(account string, amount int) *Transaction {
	r := Op{
		Type: Blockchain.OpRead,
		Key:  account,
	}
	w := Op{
		Type: Blockchain.OpWrite,
		Key:  account,
		Val:  strconv.Itoa(-amount),
	}
	return Blockchain.NewTransaction([]*Op{&r, &w}, -1, strconv.Itoa(s.txid), Blockchain.WriteCheck)
}

// Amalgamate 将储蓄账户的资金全部转到支票账户
func (s *Smallbank) Amalgamate(saving string, checking string) *Transaction {
	ra := Op{
		Type: Blockchain.OpRead,
		Key:  saving,
	}
	rb := Op{
		Type: Blockchain.OpRead,
		Key:  checking,
	}
	wa := Op{
		Type: Blockchain.OpWrite,
		Key:  saving,
		Val:  strconv.Itoa(0),
	}
	wb := Op{
		Type: Blockchain.OpWrite,
		Key:  checking,
		Val:  strconv.Itoa(0),
	}
	return Blockchain.NewTransaction([]*Op{&ra, &rb, &wa, &wb}, -1, strconv.Itoa(s.txid), Blockchain.Amalgamate)

}

// Query 查询第i个用户的saving和checking
func (s *Smallbank) Query(saving string, checking string) *Transaction {
	ra := Op{
		Type: Blockchain.OpRead,
		Key:  saving,
	}
	rb := Op{
		Type: Blockchain.OpRead,
		Key:  checking,
	}
	return Blockchain.NewTransaction([]*Op{&ra, &rb}, -1, strconv.Itoa(s.txid), Blockchain.Query)
}

func (s *Smallbank) GetRandomAmount() int {
	return RandomRange(1e3, 1e4)
}
func (s *Smallbank) GetNormalRandomIndex() int {
	return int(s.zipfian.Next(s.r))
	//n := len(s.savings)
	//hotRateCheck := rand.Float64()
	//if hotRateCheck < config.HotKeyRate {
	//	return int(rand.Float64() * float64(n) * config.HotKey)
	//} else {
	//	return int(rand.Float64()*float64(n)*(1-config.HotKey)) + int(float64(n)*config.HotKey)
	//}
	//for {
	//	x := int(rand.NormFloat64()*config.StdDiff) + n/2
	//	if x >= 0 && x < n {
	//		return x
	//	}
	//}
}

func (s *Smallbank) GetRandomTx() *Transaction {
	s.txid++
	switch rand.Int() % 6 {
	case 0:
		i := s.GetNormalRandomIndex()
		amount := s.GetRandomAmount()
		return s.TransactSavings(s.Savings[i], amount)
	case 1:
		i := s.GetNormalRandomIndex()
		amount := s.GetRandomAmount()
		return s.DepositChecking(s.Checkings[i], amount)
	case 2:
		i := s.GetNormalRandomIndex()
		j := s.GetNormalRandomIndex()
		for j == i {
			j = s.GetNormalRandomIndex()
		}
		amount := s.GetRandomAmount()
		return s.SendPayment(s.Checkings[i], s.Checkings[j], amount)
	case 3:
		i := s.GetNormalRandomIndex()
		amount := s.GetRandomAmount()
		return s.WriteCheck(s.Checkings[i], amount)
	case 4:
		i := s.GetNormalRandomIndex()
		return s.Amalgamate(s.Savings[i], s.Checkings[i])
	default:
		i := s.GetNormalRandomIndex()
		return s.Query(s.Savings[i], s.Checkings[i])
	}

	panic("err")
}

func (s *Smallbank) GenTxSet(n int) []*Transaction {
	txs := make([]*Transaction, n)
	for i := range txs {
		txs[i] = s.GetRandomTx()
		txs[i].SetId(s.txid) // 加入交易id
	}
	return txs
}

// RandomRange [l, r)
func RandomRange(l, r int) int {
	return rand.Intn(r-l) + l
}

// Read 从leveldb中读
func (s *Smallbank) Read(key string) string {
	val, _ := s.db.Get([]byte(key), nil)
	return string(val)
}

// Write 向leveldb中写
func (s *Smallbank) Write(key, val string) {
	s.db.Put([]byte(key), []byte(val), nil)
}

// Update 更新leveldb的数据
func (s *Smallbank) Update(key, val string) {
	s.db.Put([]byte(key), []byte(val), nil)
}
func (s *Smallbank) UpdateZipfian() {
	s.zipfian = NewZipfianWithItems(int64(Config.GlobalConfig.OriginKeys), Config.GlobalConfig.ZipfianConstant)
	fmt.Print("Update Zipfian Skew to ")
	fmt.Println(Config.GlobalConfig.ZipfianConstant)
}
func (s *Smallbank) UpdateZipfianWithSkew(skew float64) {
	s.zipfian = NewZipfianWithItems(int64(Config.GlobalConfig.OriginKeys), skew)
	fmt.Print("Update Zipfian Skew to ")
	fmt.Println(s.zipfian.zipfianConstant)
}
func GenSaving(n int) ([]string, []int) {
	saving := make([]string, n)
	amount := make([]int, n)
	for i := range saving {
		saving[i] = uuid.NewString()
		amount[i] = RandomRange(1e4, 1e5)
	}
	return saving, amount
}
func GenChecking(n int) ([]string, []int) {
	checking := make([]string, n)
	amount := make([]int, n)
	for i := range checking {
		checking[i] = uuid.NewString()
		amount[i] = RandomRange(1e3, 1e4)
	}
	return checking, amount
}
func NewSmallbank(path string) *Smallbank {
	// 为特定数量的用户创建一个支票账户和一个储蓄账户，第i个用户的储蓄金地址为savings[i],支票地址为checkings[i]
	saving, savingAmount := GenSaving(config.OriginKeys)
	checking, checkingAmount := GenChecking(config.OriginKeys)
	s := &Smallbank{
		Savings:   saving,
		Checkings: checking,
	}

	var err error
	s.db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		panic("open leveldb failed!")
	}
	for i := range s.Savings {
		s.db.Put([]byte(s.Savings[i]), []byte(strconv.Itoa(savingAmount[i])), nil)
		s.db.Put([]byte(s.Checkings[i]), []byte(strconv.Itoa(checkingAmount[i])), nil)
	}
	s.txid = 0
	s.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	s.zipfian = NewZipfianWithItems(int64(config.OriginKeys), config.ZipfianConstant)
	return s
}
func TestSmallbank(output bool) *Smallbank {
	smallbank := NewSmallbank(config.Path)
	fmt.Println("init smallbank success")
	if !output {
		return smallbank
	}
	return smallbank
}

var GlobalSmallBank = TestSmallbank(true)
