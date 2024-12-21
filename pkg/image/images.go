package image

import (
	"dBox/pkg/filesystem"
	"dBox/pkg/model"
	"fmt"
)

func ListImages() {
	var images model.ImageList
	images = filesystem.FindImagesSaved()
	fmt.Printf("IMAGE NAME					TAG\n")
	for _, img := range images.Image {
		for _,tag := range img.Image_Tag {
			fmt.Printf("%-20s					%s\n",img.Image_Name, tag)
		}
	}
}