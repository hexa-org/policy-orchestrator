package providers

import "net/http"

type OpaDecisionProvider struct {
}

func (o OpaDecisionProvider) BuildInput(r *http.Request) (any interface{}, err error) {
	//TODO implement me
	panic("implement me")
}

func (o OpaDecisionProvider) Allow(any interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}
