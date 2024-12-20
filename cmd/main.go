package main

import (
	"dBox/pkg/container"
	"dBox/pkg/image"
	"log"
	"os"

)

func main() {

	if os.Getenv("PROCESS") == "CHILD" {
		container.Child()
		return
	}

	switch os.Args[1] {
	case "run":
		if len(os.Args) < 4  {
			log.Fatalf("Error! Usage ./main run [image name] [tag (optional)] [command] [options]")
		}
		container.ContainerInit()
	case "pull":
		if len(os.Args) < 3 {
			log.Fatalf("Error! Usage ./main pull [image name] [tag (optional)]")
		}
		tag := "latest"
		//check if the user has provided a tag for the image else use default 'latest' tag
		if len(os.Args) == 4 {
			tag = os.Args[3]
		}
		image.InitPull(os.Args[2], tag)	
	case "images":
		image.ListImages()		
	default:
		log.Printf("%s : command not supported.", os.Args[1])
		log.Fatalf("Supported Commands: run pull images!")
	}
}
