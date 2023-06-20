package armclientsupport_test

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/google/uuid"
	"github.com/hexa-org/policy-orchestrator/internal/orchestratorproviders/microsoftazure/azureapim/armclientsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testListResourceResponse struct {
	keys   []string
	values []string
	next   int
}

type mappedResource struct {
	theMap map[string]string
}

var mapperFunc = func(page testListResourceResponse) []mappedResource {
	aMap := make(map[string]string)
	for k, aKey := range page.keys {
		aMap[aKey] = page.values[k]
	}
	return []mappedResource{{theMap: aMap}}
}

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
			respArr, mappedResArr := buildExpResponseAndResource(3)

			pager := buildTestResourcePager(tt.errIndex, respArr...)
			mapper := armclientsupport.NewArmListPageMapper[testListResourceResponse, mappedResource](pager, mapperFunc, "SomeCaller")
			act, err := mapper.Get()

			expError := tt.errIndex >= 0
			assert.Equal(t, expError, err != nil)
			assert.Equal(t, expError, len(act) == 0)
			assert.NotEqual(t, expError, len(act) == 3)
			if expError {
				assert.ErrorContains(t, err, "SomeCaller")
				assert.ErrorContains(t, err, fmt.Sprintf("some error at index %d", tt.errIndex))
			} else {
				assert.Equal(t, mappedResArr[0], act[0])
				assert.Equal(t, mappedResArr[1], act[1])
				assert.Equal(t, mappedResArr[2], act[2])
			}
		})
	}
}

func TestNewArmListPageMapper_GetOne_FromMulPages(t *testing.T) {
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
			respArr, mappedResArr := buildExpResponseAndResource(3)
			pager := buildTestResourcePager(tt.errIndex, respArr...)
			mapper := armclientsupport.NewArmListPageMapper[testListResourceResponse, mappedResource](pager, mapperFunc, "SomeCaller")
			resp, err := mapper.GetOne()

			expError := tt.errIndex == 0
			assert.Equal(t, expError, err != nil)
			if expError {
				assert.ErrorContains(t, err, "SomeCaller")
				assert.ErrorContains(t, err, fmt.Sprintf("some error at index %d", tt.errIndex))
				assert.Equal(t, mappedResource{}, resp)
			} else {
				assert.Equal(t, mappedResArr[0], resp)
			}
		})
	}
}

func TestNewArmListPageMapper_GetOne_HasOnlyOnePage(t *testing.T) {
	tests := []struct {
		name     string
		errIndex int
	}{
		{name: "error on fetch", errIndex: 0},
		{name: "success", errIndex: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respArr, mappedResArr := buildExpResponseAndResource(1)
			pager := buildTestResourcePager(tt.errIndex, respArr...)
			mapper := armclientsupport.NewArmListPageMapper[testListResourceResponse, mappedResource](pager, mapperFunc, "SomeCaller")
			resp, err := mapper.GetOne()

			expError := tt.errIndex == 0
			assert.Equal(t, expError, err != nil)
			if expError {
				assert.ErrorContains(t, err, "SomeCaller")
				assert.ErrorContains(t, err, fmt.Sprintf("some error at index %d", tt.errIndex))
				assert.Equal(t, mappedResource{}, resp)
			} else {
				assert.Equal(t, mappedResArr[0], resp)
			}
		})
	}
}

func buildExpResponseAndResource(count int) ([]testListResourceResponse, []mappedResource) {
	if count < 1 {
		panic("require count > 0")
	}

	respArr := make([]testListResourceResponse, 0)
	mappedResArr := make([]mappedResource, 0)
	for i := 0; i < count; i++ {
		keys := []string{uuid.NewString(), uuid.NewString()}
		values := []string{uuid.NewString(), uuid.NewString()}
		kvMap := make(map[string]string)
		for i, aKey := range keys {
			kvMap[aKey] = values[i]
		}

		next := -1
		if i < (count - 1) {
			next = i + 1
		}
		respArr = append(respArr, testListResourceResponse{
			keys:   keys,
			values: values,
			next:   next,
		})

		mappedResArr = append(mappedResArr, mappedResource{theMap: kvMap})
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
