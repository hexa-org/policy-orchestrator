package decisionsupportproviders

import "net/http"

type DefaultProvider struct {
}

func (d DefaultProvider) BuildInput(_ *http.Request) (any interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DefaultProvider) Allow(_ interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}
