package Config

import "runtime"

// Config 全局配置定义
type Config struct {
	LogPath           string
	OriginKeys        int // 初始Key的数量
	HotKey            float64
	HotKeyRate        float64 // 有HotKeyRate的交易访问HotKey的状态
	Path              string
	ZipfianConstant   float64
	BlockSize         int // 区块大小
	instanceNumber    int // instance个数
	ParallelingNumber int // 并发粒度
}

var GlobalConfig = Config{OriginKeys: 10000, HotKey: 0.2, HotKeyRate: 1, Path: "leveldb", ZipfianConstant: 0.6, BlockSize: 200, instanceNumber: 4, ParallelingNumber: runtime.NumCPU()}
