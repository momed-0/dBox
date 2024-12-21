package filesystem

import (
	"os"
	"log"
	"dBox/pkg/model"
	"dBox/pkg/utils"
	"path/filepath"
)

func SearchImagesSaved(imageName string,tag string) bool {
	entries, err := os.ReadDir(model.IMAGE_DIR)
	if err != nil {
        log.Fatalf("Failed to read saved files in %s , error : %v",model.IMAGE_DIR,err)
    }

	for _, file := range entries {
		if file.Name() == imageName {
			 //find all the tags and check if this tag exist
			 path := filepath.Join(model.IMAGE_DIR, file.Name())
			 tags , err := utils.ReadMetaData(path)
			 if err != nil {
				log.Fatalf("Failed to read metadata: %v",err)
			}
			 for _,entry := range tags {
				if entry == tag {
					return true
				}
			 }
		}
    }
	return false
}

func FindImagesSaved() model.ImageList {
	entries, err := os.ReadDir(model.IMAGE_DIR)
	if err != nil {
        log.Fatalf("Failed to read saved files in %s , error : %v",model.IMAGE_DIR,err)
    }
	var Images model.ImageList
	for _, file := range entries {
		
		//search all tags for this image
		path := filepath.Join(model.IMAGE_DIR, file.Name())
		tags , err := utils.ReadMetaData(path)
		if err != nil {
			log.Fatalf("Failed to read metadata: %v",err)
		}
		Images.Image = append(Images.Image, model.Image{Image_Name:file.Name(), Image_Tag: tags})
    }
	return Images
}