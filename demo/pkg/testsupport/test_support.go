package testsupport

import "time"

type TestData interface {
	SetUp()
	TearDown()
}

func WithSetUp[T TestData](data T, test func(data T)) {
	data.SetUp()
	time.Sleep(50)
	test(data)
	data.TearDown()
}
