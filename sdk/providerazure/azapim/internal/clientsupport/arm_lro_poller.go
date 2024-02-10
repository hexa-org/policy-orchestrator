package clientsupport

import (
	"context"
	"time"
)
import "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"

type ArmLroPoller[T any] interface {
	PollForResult(ctx context.Context, pollFreq time.Duration) (T, error)
}
type GetPollerFunc[T any] func() (*runtime.Poller[T], error)

type armLroPoller[T any] struct {
	getPoller GetPollerFunc[T]
}

func NewArmLroPoller[T any](getPoller GetPollerFunc[T]) ArmLroPoller[T] {
	if getPoller == nil {
		panic("Cannot create poller with nil function")
	}

	return &armLroPoller[T]{
		getPoller: getPoller,
	}
}

// PollForResult
// pollFreq - duration in seconds between polling
// pass 0 to use default 30s
func (cp *armLroPoller[T]) PollForResult(ctx context.Context, pollFreq time.Duration) (T, error) {
	poller, err := cp.getPoller()
	if err != nil || poller == nil {
		return *new(T), err
	}

	return poller.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{Frequency: pollFreq})
}
