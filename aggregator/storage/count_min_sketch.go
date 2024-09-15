package storage

import (
	"github.com/zoninnik89/ad-click-aggregator/aggregator/types"
	"hash/fnv"
	"math"
)

type CountMinSketch struct {
	depth     int
	width     int
	counts    [][]int32
	hashFuncs []func(data []byte) uint32
}

func NewCountMinSketch(depth int, width int) *CountMinSketch {
	//initialize the 2d array
	counts := make([][]int32, depth)
	for i := range counts {
		counts[i] = make([]int32, width)
	}

	// initialize hash functions (using FNV-1a hash function)
	hashFuncs := make([]func(data []byte) uint32, depth)
	for i := range hashFuncs {
		hashFuncs[i] = createHshFunction(i)
	}

	return &CountMinSketch{
		depth:     depth,
		width:     width,
		counts:    counts,
		hashFuncs: hashFuncs,
	}
}

func createHshFunction(seed int) func([]byte) uint32 {
	return func(data []byte) uint32 {
		hash := fnv.New32a()
		hash.Write(data)
		return hash.Sum32() + uint32(seed) // adding seed to reduce collisions
	}
}

func (cms *CountMinSketch) AddClick(adID string) {
	data := []byte(adID)
	for i := 0; i < cms.depth; i++ {
		hash := cms.hashFuncs[i](data)
		index := hash % uint32(cms.width)
		cms.counts[i][index]++
	}
}

func (cms *CountMinSketch) GetCount(adID string) *types.ClickCounter {
	data := []byte(adID)
	var minCount int32 = math.MaxInt32

	for i := 0; i < cms.depth; i++ {
		hash := cms.hashFuncs[i](data)
		index := hash % uint32(cms.width)
		minCount = minOfTwo(minCount, cms.counts[i][index])
	}
	return &types.ClickCounter{
		adId:        adID,
		totalClicks: minCount,
	}
}

func minOfTwo(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
