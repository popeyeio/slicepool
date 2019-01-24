package slicepool

import (
	"errors"
	"sort"
	"sync"
)

var ErrInvalidParam = errors.New("invalid param")

type Pool struct {
	length    int
	poolSizes []int
	pools     []sync.Pool
}

var defaultPool *Pool

func init() {
	defaultPool, _ = New(1<<1, 1<<20)
}

func New(minSize, maxSize int) (*Pool, error) {
	if minSize <= 0 || maxSize < minSize {
		return nil, ErrInvalidParam
	}

	length := 1
	for chunkSize := minSize; chunkSize < maxSize; chunkSize <<= 1 {
		length++
	}

	poolSizes := make([]int, length)
	pools := make([]sync.Pool, length)

	for i := 0; i < length; i++ {
		poolSizes[i] = minSize << uint(i)
		pools[i].New = func(size int) func() interface{} {
			return func() interface{} {
				v := make([]interface{}, size)
				return &v
			}
		}(poolSizes[i])
	}

	return &Pool{
		length:    length,
		poolSizes: poolSizes,
		pools:     pools,
	}, nil
}

func Get(size int) []interface{} {
	return defaultPool.Get(size)
}

func (p *Pool) Get(size int) []interface{} {
	if p.length <= 0 || size > p.poolSizes[p.length-1] {
		return make([]interface{}, 0, size)
	}

	n := sort.Search(p.length, func(i int) bool {
		return p.poolSizes[i] >= size
	})
	v := p.pools[n].Get().(*[]interface{})
	return (*v)[:0]
}

func Put(v []interface{}) {
	defaultPool.Put(v)
}

func (p *Pool) Put(v []interface{}) {
	size := cap(v)
	if p.length <= 0 || size < p.poolSizes[0] {
		return
	}

	n := sort.Search(p.length, func(i int) bool {
		return p.poolSizes[i] > size
	})
	p.pools[n-1].Put(&v)
}
