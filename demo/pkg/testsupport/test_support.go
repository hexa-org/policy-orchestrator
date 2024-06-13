package testsupport

import (
	"sync"
)

type TestData interface {
	SetUp()
	TearDown()
}

var mu sync.Mutex

func WithSetUp[T TestData](data T, test func(data T)) {
	mu.Lock()

	data.SetUp()
	test(data)
	data.TearDown()

	mu.Unlock()
}
