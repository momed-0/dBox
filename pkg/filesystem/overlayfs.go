package filesystem

import (
	"fmt"
	"os"
	"log"
	"syscall"
)

const image = "./images/root_bundle"

func SetupOverlayFS(containerName string) error {
	lowerDir := image
	upperDir := fmt.Sprintf("/tmp/%s/upper", containerName) 
	workDir := fmt.Sprintf("/tmp/%s/work", containerName)   
	rootDir := fmt.Sprintf("/tmp/%s/root", containerName)   

	for _, dir := range []string{upperDir, workDir, rootDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	if err := syscall.Mount("overlay", rootDir, "overlay", 0, options); err != nil {
		return fmt.Errorf("failed to mount overlayfs: %v", err)
	}

	return nil
}

func TeardownOverlayFS(containerName string) {
	log.Printf("Tearing down OverlayFS...")
	// Implement teardown logic
}
