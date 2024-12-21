package main

import (
	"dBox/pkg/container"
	"dBox/pkg/image"
	"dBox/pkg/filesystem"
	"log"
	"os"
	"strings"
)

func main() {

	if os.Getenv("PROCESS") == "CHILD" {
		container.Child(os.Getenv("IMAGE_NAME"),os.Getenv("IMAGE_TAG"))
		return
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 4  {
			log.Fatalf("Error! Usage ./main run [image name]:[tag (optional)] [command] [options]")
		}
		imageData := strings.Split(os.Args[2], ":")
		if len(imageData) > 2 {
			log.Fatalf("Error! Usage ./main run [image name]:[tag (optional)] [command] [options]")
		} else if len(imageData) < 2 {
			imageData = append(imageData, "latest")
		} else {
			if imageData[1] == "" {
				imageData[1] = "latest"
			}
		}
		//check if the image and version exists
		if filesystem.SearchImagesSaved(imageData[0],imageData[1]) == false {
			log.Fatalf("Image: %s with tag %s doesn't exists.First pull these image name and tag",imageData[0],imageData[1])
		}
		container.ContainerInit(imageData)
	case "pull":
		if len(os.Args) < 3 {
			log.Fatalf("Error! Usage ./main pull [image name]:[tag (optional)]")
		}
		imageData := strings.Split(os.Args[2], ":")
		//check if the user has provided a tag for the image else use default 'latest' tag
		if len(imageData) > 2 {
			log.Fatalf("Error! Usage ./main pull [image name]:[tag (optional)]")
		} else if len(imageData) < 2 {
			imageData = append(imageData, "latest")
		} else {
			if imageData[1] == "" {
				imageData[1] = "latest"
			}
		}
		image.InitPull(imageData[0],imageData[1])	
	case "images":
		image.ListImages()		
	default:
		log.Printf("%s : command not supported.", os.Args[1])
		log.Fatalf("Supported Commands: run pull images!")
	}
}
