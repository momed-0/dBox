package image

import (
	"dBox/pkg/filesystem"
	"dBox/pkg/model"
	"fmt"
)

func ListImages() {
	var images model.ImageList
	images = filesystem.FindImagesSaved()
	for _, img := range images.Image {
		fmt.Printf("Image Name: %-20s | Image Tag: %s\n", img.Image_Name, img.Image_Tag)
	}
}