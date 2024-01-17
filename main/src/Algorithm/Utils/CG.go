package Utils

import (
	"main/src/Blockchain"
	"math"
)

type Graph = [][]int
type CG struct {
	graph    Graph
	txs      []*Transaction
	subGraph [][]int
	index    int
	dfn      []int
	low      []int
	stack    []int
	visited  []bool
	cycles   [][]int
}

func NewCG(txs []*Transaction) *CG {
	cg := new(CG)
	cg.txs = txs
	cg.graph = buildConflictGraph(txs)
	cg.index = 0
	cg.dfn = make([]int, len(txs))
	cg.low = make([]int, len(txs))
	cg.stack = make([]int, 0)
	cg.visited = make([]bool, len(txs))
	for i, _ := range cg.visited {
		cg.visited[i] = false
	}
	cg.subGraph = make([][]int, 0)
	cg.cycles = make([][]int, 0)
	return cg
}
func buildConflictGraph(txs []*Transaction) Graph {
	// 初始化邻接矩阵，默认全部为0
	graph := make(Graph, len(txs))
	for i, _ := range graph {
		graph[i] = make([]int, len(txs))
		for j, _ := range graph[i] {
			graph[i][j] = 0
		}
	}
	for i, _ := range txs {
		if i == 0 {
			continue
		}
		// 获取第i个交易的读写集
		ReadKeysIni := make(map[string]bool, 0)
		WriteKeysIni := make(map[string]bool, 0)
		for _, op := range txs[i].Ops {
			if op.Type == Blockchain.OpRead {
				ReadKeysIni[op.Key] = true
			} else {
				WriteKeysIni[op.Key] = true
			}
		}
		for j, _ := range txs[:i] {
			// 获取第j个交易的读写集,j = 0 ~ i -1
			ReadKeysInj := make(map[string]bool, 0)
			WriteKeysInj := make(map[string]bool, 0)
			for _, op := range txs[j].Ops {
				if op.Type == Blockchain.OpRead {
					ReadKeysInj[op.Key] = true
				} else {
					WriteKeysInj[op.Key] = true
				}
			}
			for writekey, _ := range WriteKeysIni {
				_, exist := ReadKeysInj[writekey]
				if exist {
					graph[j][i] = 1
				}
			}
			for writekey, _ := range WriteKeysInj {
				_, exist := ReadKeysIni[writekey]
				if exist {
					graph[i][j] = 1
				}
			}
		}
	}
	return graph
}
func (cg *CG) tarjan(u int) {
	cg.dfn[u] = cg.index
	cg.low[u] = cg.index
	cg.index++
	cg.stack = append(cg.stack, u)
	cg.visited[u] = true
	for v, _ := range cg.graph[u] {
		if v == u || cg.graph[u][v] == 0 {
			continue
		}
		if !cg.visited[v] {
			cg.tarjan(v)
			cg.low[u] = int(math.Min(float64(cg.low[u]), float64(cg.low[v])))
		} else {
			for i, _ := range cg.stack {
				if cg.stack[i] == v {
					cg.low[u] = int(math.Min(float64(cg.low[u]), float64(cg.dfn[v])))
				}
			}
		}
	}
	if cg.dfn[u] == cg.low[u] {
		subG := make([]int, 0)
		subG = append(subG, u)
		v := cg.stack[len(cg.stack)-1]
		cg.stack = cg.stack[:len(cg.stack)-1]
		for u != v {
			subG = append(subG, v)
			v = cg.stack[len(cg.stack)-1]
			cg.stack = cg.stack[:len(cg.stack)-1]
		}
		cg.subGraph = append(cg.subGraph, subG)
	}
}
func (cg *CG) GetSubGraph() {
	for i, _ := range cg.txs {
		if !cg.visited[i] {
			cg.tarjan(i)
		}
	}
}
func (cg *CG) buildSubGraph(indexes []int) Graph {
	g := make(Graph, len(indexes))
	for i, _ := range g {
		g[i] = make([]int, len(indexes))
		for j, _ := range g[i] {
			g[i][j] = 0
		}
	}
	for i, _ := range indexes {
		for j, _ := range indexes {
			if cg.graph[indexes[i]][indexes[j]] == 1 {
				g[i][j] = 1
			}
		}
	}
	return g
}

func (cg *CG) GetAllCycles() {
	cg.GetSubGraph() // 利用tarjan算法将cg分解为多个强连通分量
	//fmt.Println(len(cg.subGraph))
	total := 0
	for _, indexes := range cg.subGraph {
		if len(indexes) == 1 {
			continue
		}
		total += 1
		g := cg.buildSubGraph(indexes)
		cycles := findCycles(g)
		for _, cycle := range cycles {
			for i, bias := range cycle {
				cycle[i] = indexes[bias]
			}
			cg.cycles = append(cg.cycles, cycle)
		}
	}
}

func (cg *CG) TransactionAbort() {
	counter := make(map[int]int, 0)
	for _, cycle := range cg.cycles {
		//fmt.Println(cycle)
		for _, tx := range cycle {
			_, exist := counter[tx]
			if exist {
				counter[tx] += 1
			} else {
				counter[tx] = 1
			}
		}
	}
	//fmt.Println(counter)
	stillCycle := make(map[int]bool, 0)
	for c, _ := range cg.cycles {
		stillCycle[c] = false
	}
	for !checkStillCycle(stillCycle) {
		txid := getMaxFromCounter(counter)
		cg.txs[txid].SetAbort()
		for j := 0; j < len(cg.txs); j++ {
			cg.graph[txid][j] = -1
			cg.graph[j][txid] = -1
		}
		for c, cycle := range cg.cycles {
			if stillCycle[c] {
				continue
			}
			contain := false
			for i, _ := range cycle {
				if cycle[i] == txid {
					//cycle = append(cycle[:i], cycle[i+1:]...)
					stillCycle[c] = true
					contain = true
					break
				}
			}
			if contain {
				for _, id := range cycle {
					counter[id]--
				}
			}
		}
		counter[txid] = 0
	}
}
func (cg *CG) GetGraph() Graph {
	return cg.graph
}
