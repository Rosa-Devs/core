package models

import "github.com/Rosa-Devs/Database/src/manifest"

type CreateManifestReq struct {
	Name string `json:"name"`
	Opts string `json:"opts"`
}

type CreateManifestRes struct {
	Manifest string `json:"manifest"`
}

type AddManifest struct {
	Manifest string `json:"manifest"`
}

type DeleteManifest struct {
	manifest.Manifest
}
