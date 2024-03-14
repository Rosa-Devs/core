package models

import "github.com/Rosa-Devs/Database/src/manifest"

type ChangeEventDb_req struct {
	manifest.Manifest
}

type NewMessageReq struct {
	Manifest manifest.Manifest `json:"manifest"`
	Msg      string            `json:"msg"`
}

type GetMessagesReq struct {
	manifest.Manifest
}
