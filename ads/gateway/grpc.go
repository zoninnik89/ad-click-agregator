package gateway

import (

	// protoBuff "github.com/zoninnik89/commons/api"
	"github.com/zoninnik89/commons/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}
