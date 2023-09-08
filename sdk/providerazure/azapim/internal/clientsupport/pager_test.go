package clientsupport_test

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/hexa-org/policy-orchestrator/sdk/core/rar"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/internal/clientsupport"
	"github.com/hexa-org/policy-orchestrator/sdk/providerazure/azapim/policystore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewArmListPageMapper_Get(t *testing.T) {
	tests := []struct {
		name     string
		errIndex int
	}{
		{name: "error on first fetch", errIndex: 0},
		{name: "error on second fetch", errIndex: 1},
		{name: "error on last fetch", errIndex: 2},
		{name: "success", errIndex: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respArr, mappedResArr := buildExpResponseAndResource()

			pager := buildTestResourcePager(tt.errIndex, respArr...)

			mapper := clientsupport.NewArmListPageMapper[testListResourceResponse, rar.ResourceActionRoles](pager, mapperFunc, "SomeCaller")
			act, err := mapper.Get()

			expError := tt.errIndex >= 0
			assert.Equal(t, expError, err != nil)
			assert.Equal(t, expError, len(act) == 0)

			if expError {
				assert.ErrorContains(t, err, "SomeCaller")
				assert.ErrorContains(t, err, fmt.Sprintf("some error at index %d", tt.errIndex))
			} else {
				i := 0
				for _, expRarList := range mappedResArr {
					assert.Equal(t, expRarList, []rar.ResourceActionRoles{act[i], act[i+1]})
					i += 2
				}
			}
		})
	}
}

type testListResourceResponse struct {
	keys   []string
	values []string
	next   int
}

var mapperFunc = func(page testListResourceResponse) []rar.ResourceActionRoles {
	rarList := make([]rar.ResourceActionRoles, 0)
	for k, aKey := range page.keys {
		aRar, _ := policystore.NvToRar(aKey, page.values[k])
		rarList = append(rarList, aRar)
	}
	return rarList
}

var (
	Page1 = testListResourceResponse{
		keys:   []string{"resrol-httppost-analytics", "resrol-httpdelete-analytics"},
		values: []string{`["Admin.Analytics", "User.Analytics"]`, `["Reporter.Analytics", "Partner.Analytics"]`},
		next:   1,
	}
	Page2 = testListResourceResponse{
		keys:   []string{"resrol-httppost-developer", "resrol-httpdelete-developer"},
		values: []string{`["Admin.Developer", "User.Developer"]`, `["Reporter.Developer", "Partner.Developer"]`},
		next:   2,
	}
	Page3 = testListResourceResponse{
		keys:   []string{"resrol-httppost-profile", "resrol-httpdelete-profile"},
		values: []string{`["Admin.Profile", "User.Profile"]`, `["Reporter.Profile", "Partner.Profile"]`},
		next:   -1,
	}
)

func buildExpResponseAndResource() ([]testListResourceResponse, [][]rar.ResourceActionRoles) {

	respArr := []testListResourceResponse{Page1, Page2, Page3}
	mappedResArr := make([][]rar.ResourceActionRoles, 0)
	for _, page := range respArr {
		rar1, _ := policystore.NvToRar(page.keys[0], page.values[0])
		rar2, _ := policystore.NvToRar(page.keys[1], page.values[1])
		mappedResArr = append(mappedResArr, []rar.ResourceActionRoles{rar1, rar2})
	}
	return respArr, mappedResArr
}

func buildTestResourcePager(errIndex int, expResp ...testListResourceResponse) *runtime.Pager[testListResourceResponse] {
	pager := runtime.NewPager(runtime.PagingHandler[testListResourceResponse]{
		More: func(page testListResourceResponse) bool {
			return page.next > 0
		},
		Fetcher: func(ctx context.Context, page *testListResourceResponse) (testListResourceResponse, error) {
			var next int
			if page != nil {
				next = page.next
			}

			if errIndex == next {
				return testListResourceResponse{}, fmt.Errorf("some error at index %d", errIndex)
			}
			return expResp[next], nil
		},
	})

	return pager
}
