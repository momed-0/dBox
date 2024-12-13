package filesystem

import (
	"os"
	"log"
	"dBox/pkg/model"
)


const IMAGE_DIR = "./images/"

func SearchImagesSaved(imageName string,tag string) bool {
	entries, err := os.ReadDir(IMAGE_DIR)
	if err != nil {
        log.Fatalf("Failed to read saved files in %s , error : %v",IMAGE_DIR,err)
    }

	for _, file := range entries {
		if file.Name() == imageName{
			 return true
		}
    }
	return false
}

func FindImagesSaved() model.ImageList {
	entries, err := os.ReadDir(IMAGE_DIR)
	if err != nil {
        log.Fatalf("Failed to read saved files in %s , error : %v",IMAGE_DIR,err)
    }
	var Images model.ImageList
	for _, file := range entries {
		Images.Image = append(Images.Image, model.Image{Image_Name:file.Name(), Image_Tag: "latest"})
    }
	return Images
}