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
			log.Fatalf("Error! Usage ./main run [image name] [command] [options]")
		}
		container.ContainerInit()
	case "pull":
		image.InitPull(os.Args[2], "latest")	
	case "images":
		image.ListImages()		
	default:
		log.Fatalf("%s : command not supported.", os.Args[1])
	}
}
