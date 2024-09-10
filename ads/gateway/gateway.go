package gateway

import (
	"context"
	// pb "github.com/sikozonpc/commons/api"
)

type AdGateway interface {
	CheckIfAdIsValid(ctx context.Context, adId string) (bool, error)
}
