package model

import "math"

const (
	ARCH      = "amd64"
	OS        = "linux"
	IMAGE_DIR = "./images/"
	MAX_FLOAT64 = math.MaxFloat64
)

type AuthResponse struct {
	Token string `json:"token"`
}

type ManifestList struct {
	Manifests     []ManifestFatList `json:"manifests"`
}

type ManifestFatList struct {
	Digest    string   `json:"digest"`
	Platform  Platform `json:"platform"`
}


type Platform struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
}

type Manifest struct {
	Config Config `json:"config"`
	Layers []Config `json:"layers"`
}


type Config struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Image struct {
	Image_Name string	`json:"Image_Name"`
	Image_Tag []string	`json:"Image_Tag"`
	Latest_Tag string	`json:"Latest_Tag"`
}

type ImageList struct {
	Image []Image
}
