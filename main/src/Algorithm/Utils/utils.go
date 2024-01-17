package Utils

func sort(a []int) []int {
	for i := 0; i < len(a); i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
	return a
}

func checkInPath(n int, path []int) bool {
	for _, p := range path {
		if p == n {
			return true
		}
	}
	return false
}
func findCycle(graph Graph, target int, index int, path []int, result *[][]int) {
	if graph[index][target] == 1 {
		tmp := sort(append(path, index))
		exist := false
		for _, r := range *result {
			if len(r) == len(tmp) {
				same := true
				for k, _ := range r {
					if r[k] != tmp[k] {
						same = false
						break
					}
				}
				exist = same
			}
			if exist {
				break
			}
		}
		if !exist {
			*result = append(*result, tmp)
		}
	} else {
		for i, _ := range graph[index] {
			if graph[index][i] == 1 && !checkInPath(i, path) {
				findCycle(graph, target, i, append(path, index), result)
			}
		}
	}
}
func findCycles(graph Graph) [][]int {
	results := make([][]int, 0)
	for i, _ := range graph {
		findCycle(graph, i, i, *new([]int), &results)
	}
	return results
}
func checkStillCycle(m map[int]bool) bool {
	for _, flag := range m {
		if !flag {
			return false
		}
	}
	return true
}
func getMaxFromCounter(m map[int]int) int {
	maxCount := 0
	maxid := -1
	for txid, count := range m {
		if count > maxCount {
			maxCount = count
			maxid = txid
		}
	}
	return maxid
}
func getDegree(DAG [][]int, index int) int {
	// i->j DAG[i][j] = 1
	degree := 0
	abort := 0
	for i := 0; i < len(DAG); i++ {
		if DAG[i][index] == 1 && i != index {
			degree += 1
		}
		if DAG[index][i] == -1 {
			abort += 1
		}
	}
	if abort == len(DAG) {
		return -1
	}
	return degree
}
func TopologicalOrder(DAG [][]int) []int {
	degrees := make([]int, len(DAG))
	for i, _ := range degrees {
		degrees[i] = getDegree(DAG, i)
	}
	sortResult := make([]int, 0)
	visited := make(map[int]bool, 0)
	for k := 0; k < len(DAG); k++ {
		for i := 0; i < len(DAG); i++ {
			_, flag := visited[i]
			if flag {
				continue
			}
			if degrees[i] == 0 {
				// 取出度数为0的点加入到结果中，并将它连的所有点取消连接
				sortResult = append(sortResult, i)
				visited[i] = true
				for j := 0; j < len(DAG); j++ {
					if DAG[i][j] == 1 {
						degrees[j] -= 1
					}
				}
				break
			}
		}
	}
	//fmt.Println(len(sortResult))
	return sortResult
}
