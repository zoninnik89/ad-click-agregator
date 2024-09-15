package storage

import t "github.com/zoninnik89/ad-click-aggregator/aggregator/types"

type StoreInterface interface {
	AddClick(adID string)
	GetCount(adID string) *t.ClickCounter
}
