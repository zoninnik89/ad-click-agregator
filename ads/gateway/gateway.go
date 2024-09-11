package gateway

import (
	"context"
	// protoBuff "github.com/zoninnik89/commons/api"
)

type AdGateway interface {
	CheckIfAdIsValid(ctx context.Context, adId string) (bool, error)
}
