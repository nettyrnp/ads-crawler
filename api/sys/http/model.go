package http

import "github.com/nettyrnp/ads-crawler/api/sys/entity"

type customerPortalsReq struct {
	SortBy string `json:"sort_by"`
	From   int64  `json:"from"`
	To     int64  `json:"till"`
	Desc   bool   `json:"desc"`
	Limit  uint64 `json:"limit"`
	Offset uint64 `json:"offset"`
}

type CustomerPortalsResp struct {
	Portals []*entity.Portal `json:"portals"`
	Total   int              `json:"total"`
}
