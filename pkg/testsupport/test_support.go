package testsupport

type TestData interface {
	SetUp()
	TearDown()
}

func WithSetUp[T TestData](data T, test func(data T)) {
	data.SetUp()
	test(data)
	data.TearDown()
}
