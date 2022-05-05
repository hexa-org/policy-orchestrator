package decisionproviders

import "net/http"

type DefaultProvider struct {
}

func (d DefaultProvider) BuildInput(r *http.Request) (any interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (d DefaultProvider) Allow(any interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}
