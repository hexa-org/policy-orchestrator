package clientsupport

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	logger "golang.org/x/exp/slog"
)

type ArmListPageMapper[T any, R any] interface {
	GetOne() (R, error)
	Get() ([]R, error)
}
type armListPageMapper[T any, R any] struct {
	caller   string // only used for error msg
	internal *runtime.Pager[T]
	mapper   func(page T) []R
}

func NewArmListPageMapper[T any, R any](armPager *runtime.Pager[T], mapper func(page T) []R, caller string) ArmListPageMapper[T, R] {
	return &armListPageMapper[T, R]{internal: armPager, mapper: mapper, caller: caller}
}

// GetOne - processes the first page only.
// Even if there are multiple pages this stops after fetching the first page
// Even if there is an error fetching first page, we do not fetch next page
func (p *armListPageMapper[T, R]) GetOne() (R, error) {
	var r R
	list, err := p.get(true)
	if err == nil && len(list) > 0 {
		r = list[0]
	}
	return r, err
}

func (p *armListPageMapper[T, R]) Get() ([]R, error) {
	return p.get(false)
}

func DoListAndMap[T any, R any](p *runtime.Pager[T], m func(page T) []R, caller string) ([]R, error) {
	pageMapper := NewArmListPageMapper(p, m, caller)
	resRoles, err := pageMapper.Get()
	if err != nil || len(resRoles) == 0 {
		return []R{}, err
	}
	return resRoles, nil
}

func (p *armListPageMapper[T, R]) get(firstOnly bool) ([]R, error) {
	output := make([]R, 0)
	pager := p.internal
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			parsedError := ParseResponseError(err)
			logger.Error("error in", "caller=", p.caller, "Error", parsedError)
			return []R{}, fmt.Errorf("error in %s. %v", p.caller, parsedError)
		}

		one := p.mapper(page)
		if len(one) > 0 {
			output = append(output, one...)
		}

		if firstOnly {
			break
		}
	}

	return output, nil
}
