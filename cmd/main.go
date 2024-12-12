package main

import (
	"dBox/pkg/container"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Error! Usage ./main run [command] [options]")
	}

	switch os.Args[1] {
	case "run":
		container.ContainerInit()
	case "child":
		container.Child()
	default:
		log.Fatalf("%s : command not supported.", os.Args[1])
	}
}
