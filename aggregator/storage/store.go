package storage

import t "github.com/zoninnik89/ad-click-agregator/aggregator/types"

type StoreInterface interface {
	AddClick(adID string)
	GetCount(adID string) *m.ClickCounter
}
