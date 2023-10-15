package armtestsupport

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

type fakeCountBasedPagerBuilder[T any] struct {
	numPages int
	pages    []T
	next     int
}

func NewFakeCountBasedPagerBuilder[T any](numPages int) *fakeCountBasedPagerBuilder[T] {
	return &fakeCountBasedPagerBuilder[T]{numPages: numPages, next: 0}
}

func (b *fakeCountBasedPagerBuilder[T]) AddPage(page T) {
	if len(b.pages) >= b.numPages {
		panic(fmt.Sprintf("can add maximum %d pages", b.numPages))
	}

	b.pages = append(b.pages, page)
}

func (b *fakeCountBasedPagerBuilder[T]) Pager() *runtime.Pager[T] {
	pager := runtime.NewPager(runtime.PagingHandler[T]{
		More: func(page T) bool {
			return b.next < b.numPages
		},
		Fetcher: func(ctx context.Context, t *T) (T, error) {
			aPage := b.pages[b.next]
			b.next++
			return aPage, nil
		},
	})
	return pager
}
