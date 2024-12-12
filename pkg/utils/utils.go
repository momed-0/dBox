package utils

import (
	"os"
	"log"
)

func WriteToFile(filename, message string) {
	err := os.WriteFile(filename, []byte(message), 0644)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}
