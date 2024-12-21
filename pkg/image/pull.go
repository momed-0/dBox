package image

import (
	"log"
	"dBox/pkg/filesystem"
	"dBox/pkg/request"
	"dBox/pkg/model"
)


func InitPull(imageName string,tag string) {
	log.Printf("Checking if the image %s already exits!", imageName)
	if filesystem.SearchImagesSaved(imageName,tag) == true {
		log.Printf("Image: %s with tag %s already exists\n",imageName,tag)
		return
	}
	log.Printf("Authenticating with docker hub..")
	authData, err := request.GetAuthToken(imageName)
	if err != nil {
		log.Fatalf("Failed to fetch auth token: %v", err)
	}
	log.Printf("Trying to fetch Fat Manifest list corresponding to %s..",imageName)
	manifestList, err := request.FetchManifestList(imageName, tag, authData.Token)
	if err != nil {
		log.Fatalf("Failed to fetch manifest list: %v", err)
	}
	// find the curresponding digest for the specific architecture and os
	for _, manifest := range manifestList.Manifests {
		if manifest.Platform.Architecture == model.ARCH && manifest.Platform.OS == model.OS{
			err := request.FetchManifest(imageName,tag, manifest.Digest,authData.Token)
			if err != nil {
				log.Fatalf("Error:", err)
			}	
		}
	}
}





