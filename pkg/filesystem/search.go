package filesystem

import (
	"os"
	"log"
)


const IMAGE_DIR = "./images/"

func FindImagesSaved(imageName string,tag string) bool {
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
