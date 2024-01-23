package Evaluate

func RunE1WithDifferentParams(CPUNumbers []int, path string) {
	for _, CPUNumber := range CPUNumbers {
		EvaluateTpsAndAbortNumberWithDifferentConcurrency(CPUNumber, path)
	}
}
func RunE2WithDifferentParams(blockSizes []int, path string) {
	for _, blockSize := range blockSizes {
		EvaluateAbortRateAndTpsWithDifferentBlockSize(blockSize, path)
	}
}
