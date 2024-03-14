package models

type CreatePofileReq struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type TrustReq struct {
	ProfileStorePublic
}
