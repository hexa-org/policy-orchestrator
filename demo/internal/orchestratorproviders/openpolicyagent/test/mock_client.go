package openpolicyagent_test

type MockBundleClient struct {
	GetResponse    []byte
	GetErr         error
	PostStatusCode int
	PostErr        error

	ArgPostBundle []byte
}

func (m *MockBundleClient) GetDataFromBundle(_ string) ([]byte, error) {
	return m.GetResponse, m.GetErr
}

func (m *MockBundleClient) PostBundle(bundle []byte) (int, error) {
	m.ArgPostBundle = bundle
	return m.PostStatusCode, m.PostErr
}
