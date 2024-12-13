package model


const (
	ARCH      = "amd64"
	OS        = "linux"
	IMAGE_DIR = "./images/"
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
	Image_Name string
	Image_Tag string
}

type ImageList struct {
	Image []Image
}