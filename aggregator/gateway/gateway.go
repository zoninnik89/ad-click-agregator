package gateway

import (
	"context"
	// protoBuff "github.com/zoninnik89/commons/api"
)

type AggregatorGateway interface {
	StubInterface(ctx context.Context, adID string) (bool, error)
}
