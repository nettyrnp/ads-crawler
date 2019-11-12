package repository

import (
	"github.com/nettyrnp/ads-crawler/api/sys/entity"
	"time"
)

type PortalsQueryOpts struct {
	SortBy entity.PortalSortField
	From   time.Time
	To     time.Time
	Desc   bool
	Limit  uint64
	Offset uint64
}
