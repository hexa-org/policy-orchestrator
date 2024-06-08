package testsupport

import (
	"sync"
	"time"
)

type TestData interface {
	SetUp()
	TearDown()
}

var mu sync.Mutex

func WithSetUp[T TestData](data T, test func(data T)) {
	mu.Lock()
	defer mu.Unlock()
	data.SetUp()
	time.Sleep(50)
	test(data)
	data.TearDown()

}
